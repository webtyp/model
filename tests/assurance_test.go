package model_test

import (
	"testing"

	. "webtyp.com/model"
)

// The zero value of Assurance is the weakest claim: a request that nobody
// authenticated proves nothing. Anything stronger must be set on purpose.
func TestAssuranceZeroValueIsNone(t *testing.T) {
	var a Assurance
	if a != AssuranceNone {
		t.Errorf("zero value of Assurance is %d; it must be AssuranceNone", a)
	}
}

// Enforcers compare levels with <, so the order is part of the contract.
func TestAssuranceIsOrdered(t *testing.T) {
	if !(AssuranceNone < AssuranceSingleFactor && AssuranceSingleFactor < AssuranceMultiFactor) {
		t.Error("Assurance must order None < SingleFactor < MultiFactor")
	}
}

// Logs and introspection read the word, never the number.
func TestAssuranceString(t *testing.T) {
	cases := []struct {
		a    Assurance
		want string
	}{
		{AssuranceNone, "none"},
		{AssuranceSingleFactor, "single_factor"},
		{AssuranceMultiFactor, "multi_factor"},
		{Assurance(99), "none"},
	}
	for _, c := range cases {
		if got := c.a.String(); got != c.want {
			t.Errorf("Assurance(%d).String() = %q, want %q", c.a, got, c.want)
		}
	}
}
