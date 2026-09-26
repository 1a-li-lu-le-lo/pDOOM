// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package model turns a validated data snapshot into a deterministic candidate
// result: external-forecast aggregations, the research-mode experimental causal
// model, the 0–100 indexes, sensitivity runs, driver explanations and the
// change record. Nothing in this package reads the wall clock or uses unseeded
// randomness; the same snapshot and options always yield byte-identical output.
package model

import "math"

// Mulberry32 is a small 32-bit PRNG. It is bit-for-bit identical to the
// JavaScript reference used by @pdoom/model-core so that the Scenario Lab in
// the browser reproduces the Go model:
//
//	a |= 0; a = (a + 0x6D2B79F5) | 0;
//	t = Math.imul(a ^ (a >>> 15), 1 | a);
//	t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
//	return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
type Mulberry32 struct{ a uint32 }

// NewMulberry32 seeds the generator. Seeds are reduced modulo 2^32, which is
// what the JavaScript `a |= 0` does to the bit pattern.
func NewMulberry32(seed int64) *Mulberry32 { return &Mulberry32{a: uint32(seed)} }

// Next returns a uniform float64 in [0, 1).
func (m *Mulberry32) Next() float64 {
	m.a += 0x6D2B79F5
	t := m.a
	t = (t ^ (t >> 15)) * (1 | t)
	t = (t + (t^(t>>7))*(61|t)) ^ t
	return float64(t^(t>>14)) / 4294967296
}

// Normal returns one standard normal deviate by the Box–Muller transform using
// exactly two uniforms (the sine partner is discarded so that the draw order is
// trivially reproducible in other languages). u1 is kept away from zero.
func (m *Mulberry32) Normal() float64 {
	u1 := m.Next()
	u2 := m.Next()
	if u1 < 1e-12 {
		u1 = 1e-12
	}
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}
