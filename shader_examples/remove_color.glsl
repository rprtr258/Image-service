#version 330

uniform sampler2D source;
in vec2 outTexCoords;

void main() {
    // TODO: smooth
    vec4 c = texture(source, outTexCoords);
    float threshold = 60. / 255.;
    vec3 color_to_remove = vec3(74., 38., 26.) / 255; // body color i guess
    if (max(max(abs(c.r - color_to_remove.r), abs(c.g - color_to_remove.g)), abs(c.b - color_to_remove.b)) < threshold) {
        c = vec4(vec3((c.r + c.g + c.b) / 3.), 1.);
    }
    gl_FragColor = c;
}