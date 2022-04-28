#version 330

uniform sampler2D source;
in vec2 outTexCoords;

void main() {
    ivec2 textureSize2d = textureSize(source, 0); // Width and height of texture image
    vec2 coords = (gl_FragCoord.x / textureSize2d.x < 0.5)?
        vec2(1.0 - outTexCoords.x, outTexCoords.y):
        outTexCoords;
    gl_FragColor = texture(source, coords);
}