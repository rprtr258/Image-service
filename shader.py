# # Requirements # #
# pip install glfw pyopengl pyrr pillow

import sys

import glfw
from OpenGL.GL import *
import OpenGL.GL.shaders
import numpy
from PIL import Image


# Initialize glfw
if not glfw.init():
    print("Couldn't initialize GLFW")
    exit(1)

# Create window
window = glfw.create_window(1, 1, "My OpenGL window", None, None)  # Size (1, 1) for show nothing in window
# window = glfw.create_window(800, 600, "My OpenGL window", None, None)

# Terminate if any issue
if not window:
    glfw.terminate()
    print("Couldn't create window")
    exit(1)

# Set context to window
glfw.make_context_current(window)

# Initial data
quad_data = [
#   positions     texture coordinates
    -1., -1., 0., 0., 0.,
     1., -1., 0., 1., 0.,
     1.,  1., 0., 1., 1.,
    -1.,  1., 0., 0., 1.
]
quad = numpy.array(quad_data, dtype=numpy.float32)
# Vertices indices order
indices_data = [
    0, 1, 2,
    2, 3, 0
]
indices = numpy.array(indices_data, dtype=numpy.uint32)

# Vertex shader
vertex_shader = """
#version 330
layout(location = 0) in vec3 position;
layout(location = 1) in vec2 inTexCoords;
out vec2 outTexCoords;

void main() {
    gl_Position = vec4(position, 1.0);
    outTexCoords = inTexCoords;
}
"""

# Fragment shader
fragment_shader = """
#version 330

uniform sampler2D source;
in vec2 outTexCoords;

float intensityFactor = 2.;

void main() {
    ivec2 textureSize2d = textureSize(source, 0); // Width and height of texture image

    // coloring
    //vec3 outColor = vec3(0.3, 0.1, 0.2);
    //gl_FragColor = texture(source, outTexCoords) * vec4(outColor, 1.0f);

    // inversion
    //gl_FragColor = vec4(1., 1., 1., 2.) - texture(source, outTexCoords);

    // RGB
    //vec2 uv = gl_FragCoord.xy / textureSize2d;
    //vec3 color;
    //if (uv.x < 1./3.) {
    //    color = vec3(1., 0., 0.);
    //} else if (uv.x < 2./3.) {
    //    color = vec3(0., 1., 0.);
    //} else {
    //    color = vec3(0., 0., 1.);
    //}
    //gl_FragColor = vec4(texture(source, outTexCoords).rgb * color, 1.);

    // remove body color
    // TODO: smooth
    //vec4 c = texture(source, outTexCoords);
    //float threshold = 60. / 255.;
    //if (abs(c.r - 74./255.) < threshold && abs(c.g - 38./255.) < threshold && abs(c.b - 26./255.) < threshold) {
    //    gl_FragColor = vec4(vec3((c.r + c.g + c.b) / 3.), 1.);
    //} else {
    //    gl_FragColor = c;
    //}

    // mirror horizontally
    if (gl_FragCoord.x / textureSize2d.x < 0.5) {
        gl_FragColor = texture(source, vec2(textureSize2d.x - outTexCoords.x, outTexCoords.y));
    } else {
        gl_FragColor = texture(source, outTexCoords);
    }

    //gl_FragColor = vec4((k1*t1 + k2*t2 + k3*t3)/(k1 + k2 + k3), 1.0f);
    //gl_FragColor = vec4(s00, 1.0f);
}
"""

# Compile shaders
try:
    shader = OpenGL.GL.shaders.compileProgram(
        OpenGL.GL.shaders.compileShader(vertex_shader, GL_VERTEX_SHADER),
        OpenGL.GL.shaders.compileShader(fragment_shader, GL_FRAGMENT_SHADER)
    )
except OpenGL.GL.shaders.ShaderCompilationError as e:
    print("Error compiling shader:")
    for x in e.args:
        print(x)
    exit(1)


# VBO
vertex_buffer_object = glGenBuffers(1)
glBindBuffer(GL_ARRAY_BUFFER, vertex_buffer_object)
glBufferData(GL_ARRAY_BUFFER, quad.itemsize * len(quad), quad, GL_STATIC_DRAW)


# EBO
entity_buffer_object = glGenBuffers(1)
glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, entity_buffer_object)
glBufferData(GL_ELEMENT_ARRAY_BUFFER, indices.itemsize * len(indices), indices, GL_STATIC_DRAW)

# Configure positions of initial data
glVertexAttribPointer(0, 3, GL_FLOAT, GL_FALSE, quad.itemsize * 5, ctypes.c_void_p(0))
glEnableVertexAttribArray(0)

# Configure texture coordinates of initial data
glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, quad.itemsize * 5, ctypes.c_void_p(12))
glEnableVertexAttribArray(1)


# Texture
texture = glGenTextures(1)
# Bind texture
glBindTexture(GL_TEXTURE_2D, texture)
# Texture wrapping params
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT)
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT)
# Texture filtering params
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)

# Open image
image = Image.open(sys.argv[1])
img_data = image.convert("RGBA").tobytes()
glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA, image.width, image.height, 0, GL_RGBA, GL_UNSIGNED_BYTE, img_data)

# Create render buffer with size (image.width x image.height)
rb_obj = glGenRenderbuffers(1)
glBindRenderbuffer(GL_RENDERBUFFER, rb_obj)
glRenderbufferStorage(GL_RENDERBUFFER, GL_RGBA, image.width, image.height)

# Create frame buffer
fb_obj = glGenFramebuffers(1)
glBindFramebuffer(GL_FRAMEBUFFER, fb_obj)
glFramebufferRenderbuffer(GL_FRAMEBUFFER, GL_COLOR_ATTACHMENT0, GL_RENDERBUFFER, rb_obj)

# Check frame buffer (that simple buffer should not be an issue)
status = glCheckFramebufferStatus(GL_FRAMEBUFFER)
if status != GL_FRAMEBUFFER_COMPLETE:
    print("incomplete framebuffer object")
    exit(1)

# Install program
glUseProgram(shader)

# Bind framebuffer and set viewport size
glBindFramebuffer(GL_FRAMEBUFFER, fb_obj)
glViewport(0, 0, image.width, image.height)

# Draw the quad which covers the entire viewport
glDrawElements(GL_TRIANGLES, 6, GL_UNSIGNED_INT, None)

# Read the data and create the image
image_buffer = glReadPixels(0, 0, image.width, image.height, GL_RGBA, GL_UNSIGNED_BYTE)
image_out = numpy.frombuffer(image_buffer, dtype=numpy.uint8).reshape(image.height, image.width, 4)
img = Image.fromarray(image_out, 'RGBA')
img.save(r"image_out.png")

