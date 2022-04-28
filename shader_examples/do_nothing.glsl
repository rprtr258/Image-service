#version 330

uniform sampler2D source;
in vec2 outTexCoords;

void main() {
    gl_FragColor = texture(source, outTexCoords);
}