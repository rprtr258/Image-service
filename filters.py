from time import asctime
import os

import requests
import numpy as np
from PIL import Image
from sklearn.cluster import KMeans

from convolution import apply_filter
from hilbert import hilbert_curve_filter

def get_next_filename():
    return str(asctime())

def load_image(url):
    imid = get_next_filename() # TODO: cache files by url
    image_filename = f"img/{imid}.orig.png" # TODO: save jpgs as jpg, not png
    with open(image_filename, "wb") as f:
        f.write(requests.get(url).content)
    im = Image.open(image_filename)
    return np.array(im), imid

# TODO: offload work to workers
def apply_convolution(url, kernel):
    image, imid = load_image(url)
    filtered = apply_filter(image, kernel)
    filtered_filename = f"img/{imid}.res.png"
    filtered.save(filtered_filename)
    return filtered_filename

def cluster_filter(url, N):
    image, imid = load_image(url)
    X = []
    for i in range(image.shape[0]):
        for j in range(image.shape[1]):
            X.append(image[i][j])
    X = np.array(X)
    kmeans = KMeans(n_clusters=N, random_state=0).fit(X)
    print("Clusters")
    colors = kmeans.cluster_centers_
    r = np.zeros(image.shape)
    for i in range(r.shape[0]):
        for j in range(r.shape[1]):
            d = [np.dot(image[i][j] - color, image[i][j] - color) for color in colors]
            k = 0
            for h in range(1, len(colors)):
                if d[h] < d[k]:
                    k = h
            r[i][j] = colors[k]
    filtered = Image.fromarray(r.astype(np.uint8), "RGB")
    filtered_filename = f"img/{imid}.res.png"
    filtered.save(filtered_filename)
    return filtered_filename

def hilbert_curve(url):
    image, imid = load_image(url)
    tmp = hilbert_curve_filter(image)
    filtered_filename = f"img/{imid}.res.png"
    tmp.save(filtered_filename)
    return filtered_filename

def hilbert_darken(url):
    image, imid = load_image(url)
    tmp = hilbert_curve_filter(image)
    tmp = tmp.resize((image.shape[1], image.shape[0]), resample=Image.BICUBIC)
    tmpa = np.array(tmp)
    for i in range(image.shape[0]):
        for j in range(image.shape[1]):
            for k in range(3):
                if image[i][j][k] < tmpa[i][j][k]:
                    tmpa[i][j][k] = image[i][j][k]
    filtered_filename = f"img/{imid}.res.png"
    Image.fromarray(tmpa.astype(np.uint8), "RGB").save(filtered_filename)
    return filtered_filename

# TODO: don't call bash OMG
def transfer_style(url, style_name):
    image, imid = load_image(url)
    os.chdir("fast-style-transfer/")
    os.system(f"python evaluate.py --in-path '../img/{imid}.orig.png' --out-path ../ --checkpoint ../ckpts/{style_name}.ckpt")
    os.chdir("..")
    os.system(f"mv '{imid}.orig.png' 'img/{imid}.res.png'")
    os.system(f"rm '{imid}.orig.png'")
    return f"img/{imid}.res.png"

import glfw
from OpenGL.GL import *
import OpenGL.GL.shaders
# # Requirements # #
# pip install glfw pyopengl pyrr pillow
def shader_filter(url, fragment_shader_source):
    # Initialize glfw
    if not glfw.init():
        print("Couldn't initialize GLFW")
        exit(1)

    # Create window
    window = glfw.create_window(1, 1, "My OpenGL window", None, None)  # Size (1, 1) for show nothing in window

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
    quad = np.array(quad_data, dtype=np.float32)
    # Vertices indices order
    indices_data = [
        0, 1, 2,
        2, 3, 0
    ]
    indices = np.array(indices_data, dtype=np.uint32)

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
    }"""
    fragment_shader = fragment_shader_source

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
    imid = get_next_filename() # TODO: cache files by url
    image_filename = f"img/{imid}.orig.png" # TODO: save jpgs as jpg, not png
    with open(image_filename, "wb") as f:
        f.write(requests.get(url).content)
    image = Image.open(image_filename)
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
    image_out = np.frombuffer(image_buffer, dtype=np.uint8).reshape(image.height, image.width, 4)
    img = Image.fromarray(image_out, 'RGBA')
    filtered_filename = f"img/{imid}.res.png"
    img.save(filtered_filename)
    glfw.terminate() # TODO: do earlier?
    return filtered_filename

