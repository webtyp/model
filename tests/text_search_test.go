package model_test

import (
	"testing"

	"webtyp.com/model"
)

// searchKind stands for input.Text/Phone/Email/Rut (model cannot import input).
type searchKind struct{ model.Kind }

func (searchKind) SearchableText() {}

func patientSchema() []model.Field {
	return []model.Field{
		{Name: "id", Type: searchKind{model.Text()}, DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: searchKind{model.Text()}, ServerOnly: true},
		{Name: "name", Type: searchKind{model.Text()}},
		{Name: "rut", Type: searchKind{model.Text()}},
		{Name: "phone", Type: searchKind{model.Text()}},
		{Name: "notes", Type: model.Text()},
	}
}

func TestTextSearch_FieldsSkipKeyServerOnlyAndPlainKinds(t *testing.T) {
	got := model.TextSearchFields(patientSchema())
	if len(got) != 3 || got[0] != 2 || got[1] != 3 || got[2] != 4 {
		t.Errorf("TextSearchFields = %v, want [2 3 4] (name, rut, phone)", got)
	}
}

func TestTextSearch_ValueJoinsFieldsInSearchForm(t *testing.T) {
	id, tenant, name, rut, phone, notes := "p1", "t1", "Lucía Pérez", "12.345.678-K", "+56 9 1234 5678", "Ficha 12"
	got := model.TextSearchOf(patientSchema(), []any{&id, &tenant, &name, &rut, &phone, &notes})
	if got != "lucia perez 12345678k 56912345678" {
		t.Errorf("TextSearchOf = %q, want %q", got, "lucia perez 12345678k 56912345678")
	}
}

func searchOf(t *testing.T, name, rut, phone string) string {
	t.Helper()
	id, tenant, notes := "p1", "t1", "x"
	return model.TextSearchOf(patientSchema(), []any{&id, &tenant, &name, &rut, &phone, &notes})
}

func TestTextSearch_DigitsOfTwoFieldsNeverMerge(t *testing.T) {
	if got := searchOf(t, "Lucía Pérez", "13047980-0", "+56912345678"); got != "lucia perez 130479800 56912345678" {
		t.Errorf("got %q", got)
	}
}

func TestTextSearch_EmptyFieldsAreSkipped(t *testing.T) {
	if got := searchOf(t, "José Núñez", "", ""); got != "jose nunez" {
		t.Errorf("got %q", got)
	}
}

func TestTextSearch_NoValuesIsEmpty(t *testing.T) {
	if got := searchOf(t, "", "", ""); got != "" {
		t.Errorf("got %q", got)
	}
}
