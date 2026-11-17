package model_test

import (
	"testing"

	"webtyp.com/model"
)

func TestConfigurable_AnyIndustryReturnsFreshSlice(t *testing.T) {
	a := model.AnyIndustry()
	if len(a) != 4 {
		t.Fatalf("want 4 industries, got %d", len(a))
	}
	a[0] = "tampered"
	b := model.AnyIndustry()
	if b[0] != model.IndustryHumanHealth {
		t.Fatalf("AnyIndustry mutated shared state: got %q", b[0])
	}
}

func TestConfigurable_SupportsIndustry(t *testing.T) {
	healthOnly := model.ModuleDefinition{Name: "clinical_encounter", Industries: []model.Industry{model.IndustryHumanHealth}}
	if !model.SupportsIndustry(healthOnly, model.IndustryHumanHealth) {
		t.Fatal("expected IndustryHumanHealth to be supported")
	}
	if model.SupportsIndustry(healthOnly, model.IndustryGeneralBusiness) {
		t.Fatal("expected IndustryGeneralBusiness to be rejected")
	}
	generic := model.ModuleDefinition{Name: "client_directory", Industries: model.AnyIndustry()}
	if !model.SupportsIndustry(generic, model.IndustryAnimalCare) {
		t.Fatal("expected generic module to support IndustryAnimalCare")
	}
}

func TestConfigurable_IconForIndustryFallsBackToDefault(t *testing.T) {
	def := model.ModuleDefinition{
		Name:    "client_directory",
		Icon:    "users",
		IconFor: []model.IndustryIcon{{Industry: model.IndustryAnimalCare, Icon: "paw"}},
	}
	if got := def.IconForIndustry(model.IndustryAnimalCare); got != "paw" {
		t.Fatalf("want paw, got %q", got)
	}
	if got := def.IconForIndustry(model.IndustryEducation); got != "users" {
		t.Fatalf("want fallback users, got %q", got)
	}
}

func TestConfigurable_RoleUseForAppliesOverride(t *testing.T) {
	role := model.Role{
		Code:      "reception",
		Name:      "Reception",
		Use:       model.WorksWithAllRecords,
		Overrides: []model.ModuleUseOverride{{Module: "cash_manager", Use: model.SeesOwnRecords}},
	}
	if got := role.UseFor("client_directory"); got != model.WorksWithAllRecords {
		t.Fatalf("want WorksWithAllRecords, got %v", got)
	}
	if got := role.UseFor("cash_manager"); got != model.SeesOwnRecords {
		t.Fatalf("want SeesOwnRecords override, got %v", got)
	}
}
