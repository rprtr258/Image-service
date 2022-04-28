#version 330

uniform sampler2D source;
in vec2 outTexCoords;

void main() {
    gl_FragColor = vec4(1., 1., 1., 2.) - texture(source, outTexCoords);
}