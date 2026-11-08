package model_test

import (
	"testing"

	"webtyp.com/model"
)

func TestDefinition(t *testing.T) {
	// Construct a Definition literal
	var UserModel = model.Definition{
		Name: "user",
		Fields: model.Fields{
			{Name: "id", Type: model.Int(), DB: &model.FieldDB{PK: true, AutoInc: true}},
			{Name: "name", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 2}},
			{Name: "email", Type: model.Text(), NotNull: true},
		},
	}

	// Assert Name and number of fields
	if UserModel.Name != "user" {
		t.Errorf("expected Name 'user', got %q", UserModel.Name)
	}
	if len(UserModel.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(UserModel.Fields))
	}

	// Definition.Field("email") returns the correct field
	f, ok := UserModel.Field("email")
	if !ok {
		t.Error("expected field 'email' to be found")
	}
	if f.Name != "email" {
		t.Errorf("expected field name 'email', got %q", f.Name)
	}

	// Definition.Field("nope") returns false
	_, ok = UserModel.Field("nope")
	if ok {
		t.Error("expected field 'nope' not to be found")
	}

	// Assert that Fields is assignable from a []Field literal without conversion (alias)
	checkFields := func(f model.Fields) {}
	checkFields([]model.Field{
		{Name: "test", Type: model.Text()},
	})

	// Assert that Kind Struct(ref) preserves the Ref pointer
	SomeModel := &model.Definition{Name: "some"}
	fStruct := model.Field{Name: "nested", Type: model.Struct(SomeModel)}
	if rk, ok := fStruct.Type.(model.RefKind); !ok || rk.Ref() != SomeModel {
		t.Error("Ref pointer not preserved in Kind")
	}

	// Assert that the zero-value of Ref is nil
	var fZero model.Field
	if fZero.Ref != nil {
		t.Error("expected zero-value Ref to be nil")
	}

	// Assert that Exclude defaults to false and can be set explicitly.
	if fZero.Exclude {
		t.Error("expected zero-value Exclude to be false")
	}
	fExcluded := model.Field{Name: "password_hash", Type: model.Text(), Exclude: true}
	if !fExcluded.Exclude {
		t.Error("expected Exclude to be true when set")
	}

	// Assert Ref disambiguated by Type: scalar FK usage (Ref + FieldDB.RefColumn/OnDelete).
	StaffModel := &model.Definition{Name: "staff"}
	fFK := model.Field{
		Name: "staff_id", Type: model.Int(), NotNull: true, Ref: StaffModel,
		DB: &model.FieldDB{RefColumn: "id", OnDelete: "CASCADE"},
	}
	if fFK.Ref != StaffModel {
		t.Error("expected scalar FK Ref to be preserved")
	}
	if fFK.DB.RefColumn != "id" || fFK.DB.OnDelete != "CASCADE" {
		t.Error("expected FieldDB.RefColumn/OnDelete to be preserved")
	}

	// Zero-value of the new FieldDB members must be empty strings.
	var dbZero model.FieldDB
	if dbZero.RefColumn != "" || dbZero.OnDelete != "" {
		t.Error("expected zero-value RefColumn/OnDelete to be empty")
	}
}

func TestDefinition_AllowZeroIsClosed(t *testing.T) {
	d := model.Definition{}
	if d.Allow.Has(model.Read) {
		t.Error("expected zero Allow to not have Read")
	}
	if d.Allow.Has(model.Create) {
		t.Error("expected zero Allow to not have Create")
	}
	if d.Allow != 0 {
		t.Errorf("expected zero Allow, got %v", d.Allow)
	}
}

func TestDefinition_AllowHas(t *testing.T) {
	d := model.Definition{Allow: model.Create | model.Read}
	if !d.Allow.Has(model.Create) {
		t.Error("expected Allow to have Create")
	}
	if !d.Allow.Has(model.Read) {
		t.Error("expected Allow to have Read")
	}
	if d.Allow.Has(model.Delete) {
		t.Error("expected Allow not to have Delete")
	}
	if d.Allow.Has(model.Update) {
		t.Error("expected Allow not to have Update")
	}
}

func TestKey_JoinSplitRoundTrip(t *testing.T) {
	wire := model.JoinKey("r1", "e7")
	if wire != "r1/e7" {
		t.Fatalf("expected 'r1/e7', got %q", wire)
	}
	parts, ok := model.SplitKey(wire, 2)
	if !ok {
		t.Fatal("expected SplitKey to succeed")
	}
	if len(parts) != 2 || parts[0] != "r1" || parts[1] != "e7" {
		t.Fatalf("unexpected parts: %v", parts)
	}
}

func TestKey_SplitWrongCountNotOK(t *testing.T) {
	_, ok := model.SplitKey("a/b/c", 2)
	if ok {
		t.Error("expected SplitKey('a/b/c', 2) to fail")
	}
	_, ok = model.SplitKey("a/b", 3)
	if ok {
		t.Error("expected SplitKey('a/b', 3) to fail")
	}
}

func TestKey_SplitEmptyPartNotOK(t *testing.T) {
	cases := []string{"a/", "/b", "a//b", ""}
	for _, tc := range cases {
		_, ok := model.SplitKey(tc, 2)
		if ok {
			t.Errorf("expected SplitKey(%q, 2) to fail", tc)
		}
	}
}

