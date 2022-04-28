#version 330

uniform sampler2D source;
in vec2 outTexCoords;

void main() {
    vec3 outColor = vec3(0.3, 0.1, 0.2);
    gl_FragColor = texture(source, outTexCoords) * vec4(outColor, 1.0f);
}