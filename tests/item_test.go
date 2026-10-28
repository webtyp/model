package model_test

import (
	"strings"
	"testing"

	. "webtyp.com/model"
)

func TestJoinItem(t *testing.T) {
	if got := JoinItem("local", "", "Oficina"); got != "local · Oficina" {
		t.Fatalf("expected 'local · Oficina', got %q", got)
	}
	if got := JoinItem(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	if got := JoinItem("a"); got != "a" {
		t.Fatalf("expected 'a', got %q", got)
	}
}

func TestValidateItem_Valid(t *testing.T) {
	fields := []Field{
		{Name: "id", Type: Text(), DB: &FieldDB{PK: true}},
		{Name: "name", Type: Text(), Item: ItemLabel},
		{Name: "sku", Type: Text(), Item: ItemDescription},
	}
	if err := ValidateItem(fields); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateItem_NoMarksIsValid(t *testing.T) {
	fields := []Field{
		{Name: "name", Type: Text()},
		{Name: "sku", Type: Text()},
	}
	if err := ValidateItem(fields); err != nil {
		t.Fatalf("expected valid when no marks and no PK, got: %v", err)
	}
}

func TestValidateItem_TwoLabels(t *testing.T) {
	fields := []Field{
		{Name: "id", Type: Text(), DB: &FieldDB{PK: true}},
		{Name: "a", Type: Text(), Item: ItemLabel},
		{Name: "b", Type: Text(), Item: ItemLabel},
	}
	err := ValidateItem(fields)
	if err == nil || !strings.Contains(err.Error(), "two fields marked ItemLabel: a, b") {
		t.Fatalf("expected error containing 'two fields marked ItemLabel: a, b', got: %v", err)
	}
}

func TestValidateItem_TwoBadges(t *testing.T) {
	fields := []Field{
		{Name: "id", Type: Text(), DB: &FieldDB{PK: true}},
		{Name: "a", Type: Text(), Item: ItemLabel},
		{Name: "b", Type: Text(), Item: ItemBadge},
		{Name: "c", Type: Text(), Item: ItemBadge},
	}
	err := ValidateItem(fields)
	if err == nil || !strings.Contains(err.Error(), "two fields marked ItemBadge") {
		t.Fatalf("expected error containing 'two fields marked ItemBadge', got: %v", err)
	}
}

func TestValidateItem_DescriptionWithoutLabel(t *testing.T) {
	fields := []Field{
		{Name: "id", Type: Text(), DB: &FieldDB{PK: true}},
		{Name: "desc", Type: Text(), Item: ItemDescription},
	}
	err := ValidateItem(fields)
	if err == nil || !strings.Contains(err.Error(), "need a field marked ItemLabel") {
		t.Fatalf("expected error containing 'need a field marked ItemLabel', got: %v", err)
	}
}

func TestValidateItem_NoPK(t *testing.T) {
	fields := []Field{
		{Name: "name", Type: Text(), Item: ItemLabel},
	}
	err := ValidateItem(fields)
	if err == nil || !strings.Contains(err.Error(), "primary key field") {
		t.Fatalf("expected error containing 'primary key field', got: %v", err)
	}
}

func TestValidateItem_RefAndOptions(t *testing.T) {
	dummy := &Definition{Name: "dummy"}
	fields := []Field{
		{Name: "zone_id", Type: Text(), Ref: dummy, Options: dummy},
	}
	err := ValidateItem(fields)
	if err == nil || !strings.Contains(err.Error(), "both Ref and Options: zone_id") {
		t.Fatalf("expected error containing 'both Ref and Options: zone_id', got: %v", err)
	}
}
