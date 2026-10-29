package model_test

import (
	"testing"
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// Root-level test (justified): tests error.go behavior which is critical for error formatting and status response mapping.

func TestKindOf_Declared(t *testing.T) {
	if k := model.KindOf(model.NotFound("x")); k != model.ErrorNotFound {
		t.Errorf("got %v, want %v", k, model.ErrorNotFound)
	}
	if k := model.KindOf(model.Conflict("x")); k != model.ErrorConflict {
		t.Errorf("got %v, want %v", k, model.ErrorConflict)
	}
	if k := model.KindOf(model.Invalid("x")); k != model.ErrorInvalid {
		t.Errorf("got %v, want %v", k, model.ErrorInvalid)
	}
}

func TestKindOf_PlainIsInternal(t *testing.T) {
	if k := model.KindOf(fmt.Err("boom")); k != model.ErrorInternal {
		t.Errorf("got %v, want %v", k, model.ErrorInternal)
	}
}

type errTestManualFielder struct {
	schema []model.Field
	ptrs   []any
}

func (f *errTestManualFielder) Schema() []model.Field { return f.schema }
func (f *errTestManualFielder) Pointers() []any { return f.ptrs }

func TestFieldOf_RequiredField(t *testing.T) {
	type userActionMock struct {
		Name string
	}
	schema := []model.Field{
		{Name: "name", Type: model.Text(), NotNull: true},
	}
	m := &userActionMock{Name: ""}
	f := &errTestManualFielder{
		schema: schema,
		ptrs:   []any{&m.Name},
	}

	err := model.ValidateFields(model.ActionCreate, f)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if field := model.FieldOf(err); field != "name" {
		t.Errorf("FieldOf(err) = %q, want %q", field, "name")
	}

	if k := model.KindOf(err); k != model.ErrorInvalid {
		t.Errorf("KindOf(err) = %v, want %v", k, model.ErrorInvalid)
	}

	// Copy expected text format from tests/field_test.go logic (fmt.Err(f.Name, "required") -> "name required")
	expectedErrText := "name required"
	if err.Error() != expectedErrText {
		t.Errorf("err.Error() = %q, want %q", err.Error(), expectedErrText)
	}
}

func TestFieldOf_TooLong(t *testing.T) {
	type userActionMock struct {
		Name string
	}
	schema := []model.Field{
		{Name: "short_name", Type: model.Text(), Permitted: model.Permitted{Maximum: 3}},
	}
	m := &userActionMock{Name: "abcd"}
	f := &errTestManualFielder{
		schema: schema,
		ptrs:   []any{&m.Name},
	}

	err := model.ValidateFields(model.ActionCreate, f)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if field := model.FieldOf(err); field != "short_name" {
		t.Errorf("FieldOf(err) = %q, want %q", field, "short_name")
	}
}

func TestFieldOf_NotAField(t *testing.T) {
	if field := model.FieldOf(model.NotFound("x")); field != "" {
		t.Errorf("got %q, want empty string", field)
	}
}

func TestErrorKind_String(t *testing.T) {
	cases := []struct {
		k model.ErrorKind
		w string
	}{
		{model.ErrorInternal, "internal"},
		{model.ErrorInvalid, "invalid"},
		{model.ErrorNotFound, "not found"},
		{model.ErrorConflict, "conflict"},
	}

	for _, c := range cases {
		if c.k.String() != c.w {
			t.Errorf("ErrorKind(%d).String() = %q, want %q", c.k, c.k.String(), c.w)
		}
	}
}
