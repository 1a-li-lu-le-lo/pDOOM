// Copyright NU Cybernetics. p(DOOM) — research prototype.
// GLSL for the Event Horizon scene: a radial glow disc, the accretion band (soft
// edges plus a crisp median line) and orbiting points whose motion is computed in
// the vertex shader from seeded per-point attributes, so no buffer is written on
// the CPU per frame. Colours are converted to the output colour space at the end
// of each fragment program, as three.js's built-in materials do.

/** Shared by the glow disc and the band: passes the local XY position through. */
export const PLANE_VERT = /* glsl */ `
varying vec2 vPos;
void main() {
  vPos = position.xy;
  gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
}
`;

export const GLOW_FRAG = /* glsl */ `
uniform float uRadius;
uniform float uHole;
uniform float uGlow;
uniform float uOpacity;
uniform vec3 uHot;
uniform vec3 uWarm;
uniform vec3 uCold;
varying vec2 vPos;
void main() {
  float d = length(vPos) / uRadius;
  float hole = smoothstep(uHole, uHole + 0.08, d);
  vec3 col = mix(uHot, uWarm, smoothstep(0.0, 0.35, d));
  col = mix(col, uCold, smoothstep(0.35, 0.7, d));
  float a = mix(uGlow, uGlow * 0.5, smoothstep(0.0, 0.35, d));
  a = mix(a, 0.25, smoothstep(0.35, 0.7, d));
  a = mix(a, 0.0, smoothstep(0.7, 1.0, d));
  gl_FragColor = vec4(col, a * hole * uOpacity);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
`;

export const BAND_FRAG = /* glsl */ `
uniform float uInner;
uniform float uOuter;
uniform float uFeather;
uniform float uMid;
uniform float uTime;
uniform float uOpacity;
uniform float uLine;
uniform vec3 uColor;
varying vec2 vPos;
void main() {
  float d = length(vPos);
  float band = smoothstep(uInner, uInner + uFeather, d) * (1.0 - smoothstep(uOuter - uFeather, uOuter, d));
  float ang = atan(vPos.y, vPos.x);
  float drift = 0.82 + 0.18 * sin(ang * 3.0 + uTime * 0.05) * sin(ang * 5.0 - uTime * 0.03);
  float line = 1.0 - smoothstep(0.0, 0.016, abs(d - uMid));
  float a = band * drift * uOpacity + line * uLine;
  gl_FragColor = vec4(uColor, min(a, 0.9));
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
`;

export const POINT_VERT = /* glsl */ `
uniform float uTime;
uniform float uPixelRatio;
uniform float uScale;
attribute float aRadius;
attribute float aAngle;
attribute float aSpeed;
attribute float aSize;
attribute float aAlpha;
attribute float aHeight;
varying float vAlpha;
void main() {
  float ang = aAngle + uTime * aSpeed;
  vec3 p = vec3(cos(ang) * aRadius, aHeight, sin(ang) * aRadius);
  vec4 mv = modelViewMatrix * vec4(p, 1.0);
  gl_Position = projectionMatrix * mv;
  gl_PointSize = aSize * uPixelRatio * uScale / max(0.001, -mv.z);
  vAlpha = aAlpha;
}
`;

export const POINT_FRAG = /* glsl */ `
uniform vec3 uColor;
uniform float uOpacity;
varying float vAlpha;
void main() {
  vec2 c = gl_PointCoord - vec2(0.5);
  float d = length(c) * 2.0;
  float m = 1.0 - smoothstep(0.5, 1.0, d);
  gl_FragColor = vec4(uColor, m * vAlpha * uOpacity);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
`;
