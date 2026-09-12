package model

import "testing"

// A list must NOT be given Schema() or Pointers() "to be helpful" — that is
// the defect this file pins: while FielderSlice embedded Fielder, every
// generated list answered Schema() with nil, which made a list satisfy
// Fielder and let it be passed where a record was required.

// probeRow is a minimal Fielder: a record with a real two-column schema.
type probeRow struct {
	a string
	b string
}

func (r *probeRow) Schema() []Field {
	return []Field{{Name: "a"}, {Name: "b"}}
}

func (r *probeRow) Pointers() []any {
	return []any{&r.a, &r.b}
}

// onlySlice implements ONLY the narrowed contract — no Schema, no Pointers.
type onlySlice []*probeRow

func (s *onlySlice) Len() int         { return len(*s) }
func (s *onlySlice) At(i int) Fielder { return (*s)[i] }
func (s *onlySlice) Append() Fielder  { v := &probeRow{}; *s = append(*s, v); return v }

var _ FielderSlice = (*onlySlice)(nil)

// La lista atraviesa filas; el esquema vive en el elemento.
func TestSliceNeedsNoSchema(t *testing.T) {
	var s onlySlice
	if got := len(s.Append().Schema()); got != 2 {
		t.Errorf("element schema = %d fields, want 2", got)
	}
	if s.Len() != 1 {
		t.Errorf("Len = %d after one Append, want 1", s.Len())
	}
}
