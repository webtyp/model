package model_test

import (
	"testing"

	"webtyp.com/model"
)

// Root-level test (justified): tests for basic Resource construction functions.

func TestResourceOf_Table(t *testing.T) {
	def := &model.Definition{
		Name: "service",
		Fields: model.Fields{
			{Name: "id", DB: &model.FieldDB{PK: true}},
			{Name: "name"},
		},
	}
	got := model.ResourceOf(def)
	if got != "service" {
		t.Errorf("ResourceOf() = %q; want %q", got, "service")
	}
}

func TestResourceOf_NotATable(t *testing.T) {
	// A Definition with no PK field (an args DTO)
	def := &model.Definition{
		Name: "service_args",
		Fields: model.Fields{
			{Name: "name"},
		},
	}
	got := model.ResourceOf(def)
	if got != "" {
		t.Errorf("ResourceOf() = %q; want %q", got, "")
	}
}

func TestResourceOf_Nil(t *testing.T) {
	got := model.ResourceOf(nil)
	if got != "" {
		t.Errorf("ResourceOf(nil) = %q; want %q", got, "")
	}
}

func TestFeatureResource(t *testing.T) {
	got := model.FeatureResource("api_explorer")
	if got != "api_explorer" {
		t.Errorf("FeatureResource() = %q; want %q", got, "api_explorer")
	}
}
