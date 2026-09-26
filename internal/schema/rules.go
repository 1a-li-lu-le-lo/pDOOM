// Copyright NU Cybernetics. p(DOOM) — research prototype.

package schema

import (
	"fmt"
	"strconv"
	"strings"
)

// Rule is a parsed editorial rule condition. The grammar is deliberately tiny:
//
//	expr    := clause ( ("and" | "or") clause )*
//	clause  := "true" | "always" | "false" | ident op number
//	op      := ">=" | "<=" | "==" | "!=" | ">" | "<"
//
// "and" binds tighter than "or". Identifiers are variable names such as
// uncertainty, coverage, cpi, csi, ipi, epi, air. Unknown variables make Eval
// return an error rather than a silent false.
type Rule struct {
	orGroups [][]clause
	text     string
}

type clause struct {
	constant *bool
	ident    string
	op       string
	value    float64
}

// ParseRule parses a rule condition string.
func ParseRule(text string) (Rule, error) {
	r := Rule{text: text}
	src := strings.TrimSpace(text)
	if src == "" {
		return r, fmt.Errorf("rule: empty condition")
	}
	for _, orPart := range splitKeyword(src, "or") {
		var group []clause
		for _, andPart := range splitKeyword(orPart, "and") {
			c, err := parseClause(strings.TrimSpace(andPart))
			if err != nil {
				return r, fmt.Errorf("rule %q: %w", text, err)
			}
			group = append(group, c)
		}
		r.orGroups = append(r.orGroups, group)
	}
	return r, nil
}

// splitKeyword splits on a lowercase keyword surrounded by whitespace.
func splitKeyword(s, kw string) []string {
	fields := strings.Fields(s)
	var parts []string
	var cur []string
	for _, f := range fields {
		if strings.EqualFold(f, kw) {
			parts = append(parts, strings.Join(cur, " "))
			cur = nil
			continue
		}
		cur = append(cur, f)
	}
	parts = append(parts, strings.Join(cur, " "))
	return parts
}

func parseClause(s string) (clause, error) {
	switch strings.ToLower(s) {
	case "true", "always":
		t := true
		return clause{constant: &t}, nil
	case "false", "never":
		f := false
		return clause{constant: &f}, nil
	}
	for _, op := range []string{">=", "<=", "==", "!=", ">", "<"} {
		if i := strings.Index(s, op); i > 0 {
			ident := strings.TrimSpace(s[:i])
			rhs := strings.TrimSpace(s[i+len(op):])
			if ident == "" || strings.ContainsAny(ident, " <>=!") {
				return clause{}, fmt.Errorf("bad identifier in %q", s)
			}
			v, err := strconv.ParseFloat(rhs, 64)
			if err != nil {
				return clause{}, fmt.Errorf("bad number in %q", s)
			}
			return clause{ident: ident, op: op, value: v}, nil
		}
	}
	return clause{}, fmt.Errorf("cannot parse clause %q", s)
}

// Eval evaluates the rule against variable values. Variables whose value is
// unavailable should be absent from vars; a clause on an absent variable is an error.
func (r Rule) Eval(vars map[string]float64) (bool, error) {
	for _, group := range r.orGroups {
		all := true
		for _, c := range group {
			ok, err := c.eval(vars)
			if err != nil {
				return false, err
			}
			if !ok {
				all = false
				break
			}
		}
		if all {
			return true, nil
		}
	}
	return false, nil
}

func (c clause) eval(vars map[string]float64) (bool, error) {
	if c.constant != nil {
		return *c.constant, nil
	}
	v, ok := vars[c.ident]
	if !ok {
		return false, fmt.Errorf("rule variable %q unavailable", c.ident)
	}
	switch c.op {
	case ">=":
		return v >= c.value, nil
	case "<=":
		return v <= c.value, nil
	case ">":
		return v > c.value, nil
	case "<":
		return v < c.value, nil
	case "==":
		return v == c.value, nil
	case "!=":
		return v != c.value, nil
	}
	return false, fmt.Errorf("unknown operator %q", c.op)
}

// Variables lists the variable names referenced by the rule.
func (r Rule) Variables() []string {
	seen := map[string]bool{}
	var out []string
	for _, g := range r.orGroups {
		for _, c := range g {
			if c.ident != "" && !seen[c.ident] {
				seen[c.ident] = true
				out = append(out, c.ident)
			}
		}
	}
	return out
}

// String returns the original condition text.
func (r Rule) String() string { return r.text }
