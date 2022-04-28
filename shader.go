package fimgs

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile shader source:\n%v\nerror:\n%v", source, log)
	}
	return shader, nil
}

func newTexture(file string) (int, int, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)
	return rgba.Rect.Size().X, rgba.Rect.Size().Y, nil
}

// requires libgl1-mesa-dev, xorg-dev packages
func ShaderFilter(sourceImageFilename, resultImageFilename, fragmentShaderSource string) error {
	if err := glfw.Init(); err != nil {
		return fmt.Errorf("couldn't initialize glfw: %q", err)
	}
	defer glfw.Terminate()

	// FIX: panics sometimes in webserver mode
	window, err := glfw.CreateWindow(1, 1, "Thou shalt not exist", nil, nil) // Size (1, 1) for show nothing in window
	if err != nil {
		return fmt.Errorf("couldn't create window: %q", err)
	}

	// Set context to window
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		return fmt.Errorf("couldn't initialize glow: %q", err)
	}

	// Initial data
	quad := []float32{
		// [x, y, z=0] positions [u=(x+1)/2, v=(y+1)/2] texture coordinates
		-1., -1., 0., 0., 0.,
		1., -1., 0., 1., 0.,
		1., 1., 0., 1., 1.,
		-1., 1., 0., 0., 1.,
	}
	// Vertices indices order
	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	// Vertex shader
	vertex_shader := `#version 330
layout(location = 0) in vec3 position;
layout(location = 1) in vec2 inTexCoords;
out vec2 outTexCoords;

void main() {
    gl_Position = vec4(position, 1.0);
    outTexCoords = inTexCoords;
}`

	// Compile shaders
	vertexShader, err := compileShader(vertex_shader, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("error compiling vertex shader:\n%q", err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("error compiling fragment shader:\n%s", err)
	}
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var vertex_buffer_object uint32
	gl.GenBuffers(1, &vertex_buffer_object)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertex_buffer_object)
	gl.BufferData(gl.ARRAY_BUFFER, len(quad)*4, gl.Ptr(quad), gl.STATIC_DRAW)

	// EBO
	var entity_buffer_object uint32
	gl.GenBuffers(1, &entity_buffer_object)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, entity_buffer_object)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Configure positions of initial data
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 4*5, 0)
	gl.EnableVertexAttribArray(0)

	// Configure texture coordinates of initial data
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 4*5, 12)
	gl.EnableVertexAttribArray(1)

	imageWidth, imageHeight, err := newTexture(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("error loading texture: %q", err)
	}

	// Create render buffer with size (image.width x image.height)
	var rb_obj uint32
	gl.GenRenderbuffers(1, &rb_obj)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rb_obj)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.RGBA, int32(imageWidth), int32(imageHeight))

	// Create frame buffer
	var fb_obj uint32
	gl.GenFramebuffers(1, &fb_obj)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb_obj)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, rb_obj)

	// FIX: sometimes fail with 0
	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("incomplete framebuffer object, status is %d, gl error is %q", status, gl.GetError())
	}

	// Install program
	gl.UseProgram(program)

	// Bind framebuffer and set viewport size
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb_obj)
	gl.Viewport(0, 0, int32(imageWidth), int32(imageHeight))

	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

	var data []uint8 = make([]uint8, 4*imageHeight*imageWidth)
	gl.ReadPixels(0, 0, int32(imageWidth), int32(imageHeight), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&data)).Data))

	image_out := image.RGBA{
		Pix:    data,
		Stride: imageWidth * 4,
		Rect:   image.Rect(0, 0, imageWidth, imageHeight),
	}
	if err = saveImage(&image_out, resultImageFilename); err != nil {
		return fmt.Errorf("error saving file: %q", err)
	}
	return nil
}
