#version 330

uniform sampler2D source;
in vec2 outTexCoords;

void main() {
    ivec2 textureSize2d = textureSize(source, 0); // Width and height of texture image
    vec2 uv = gl_FragCoord.xy / textureSize2d;
    vec3 color;
    if (uv.x < 1./3.) {
        color = vec3(1., 0., 0.);
    } else if (uv.x < 2./3.) {
        color = vec3(0., 1., 0.);
    } else {
        color = vec3(0., 0., 1.);
    }
    gl_FragColor = vec4(texture(source, outTexCoords).rgb * color, 1.);
}