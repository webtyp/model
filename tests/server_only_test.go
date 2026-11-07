package model_test

import (
	"testing"

	"webtyp.com/model"
)

type serverOnlyRec struct {
	TenantID string
	Name     string
}

func (r *serverOnlyRec) Schema() []model.Field {
	return []model.Field{
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "name", Type: model.Text(), NotNull: true},
	}
}
func (r *serverOnlyRec) Pointers() []any { return []any{&r.TenantID, &r.Name} }

func TestValidateInput_SkipsServerOnly(t *testing.T) {
	rec := &serverOnlyRec{Name: "Ana"}
	if err := model.ValidateInput(model.ActionCreate, rec); err != nil {
		t.Fatalf("empty ServerOnly NotNull field must be valid input, got %v", err)
	}
}

func TestValidateInput_StillChecksClientFields(t *testing.T) {
	rec := &serverOnlyRec{}
	if err := model.ValidateInput(model.ActionCreate, rec); model.FieldOf(err) != "name" {
		t.Fatalf("want error on name, got %v", err)
	}
}

func TestValidateFields_ChecksServerOnly(t *testing.T) {
	rec := &serverOnlyRec{Name: "Ana"}
	if err := model.ValidateFields(model.ActionCreate, rec); model.FieldOf(err) != "tenant_id" {
		t.Fatalf("ValidateFields must still require tenant_id, got %v", err)
	}
}
