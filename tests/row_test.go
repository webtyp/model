package model_test

import (
	"strings"
	"testing"
	"webtyp.com/model"
)

// Root-level test (justified): Testing the behaviour of RowOf and ValidateRow on basic Fielder logic.

type rowDevice struct {
	Id, Name, Access, Location string
	Port                       int64
}

var rowDeviceFields = model.Fields{
	{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
	{Name: "name", Type: model.Text(), Row: model.RowTitle},
	{Name: "access", Type: model.Text(), Row: model.RowDetail},
	{Name: "location", Type: model.Text(), Row: model.RowDetail},
	{Name: "port", Type: model.Int()},
}

func (d *rowDevice) Schema() []model.Field { return rowDeviceFields }
func (d *rowDevice) Pointers() []any       { return []any{&d.Id, &d.Name, &d.Access, &d.Location, &d.Port} }

func TestRowOf_TitleAndJoinedDetails(t *testing.T) {
	device := rowDevice{Id: "a1", Name: "Cesar", Access: "local", Location: "Oficina"}
	row := model.RowOf(&device)

	if row.ID != "a1" {
		t.Errorf("expected ID 'a1', got %q", row.ID)
	}
	if row.Title != "Cesar" {
		t.Errorf("expected Title 'Cesar', got %q", row.Title)
	}
	if row.Detail != "local · Oficina" {
		t.Errorf("expected Detail 'local · Oficina', got %q", row.Detail)
	}
}

func TestRowOf_EmptyDetailSkipped(t *testing.T) {
	device := rowDevice{Id: "a2", Name: "Julia", Access: "local", Location: ""}
	row := model.RowOf(&device)

	if row.Detail != "local" {
		t.Errorf("expected Detail 'local', got %q", row.Detail)
	}
}

func TestRowOf_NotInRowIgnored(t *testing.T) {
	device := rowDevice{Id: "a3", Name: "Test", Access: "remote", Location: "Server", Port: 8080}
	row := model.RowOf(&device)

	if strings.Contains(row.Title, "8080") || strings.Contains(row.Detail, "8080") || strings.Contains(row.Badge, "8080") {
		t.Errorf("expected Port '8080' to be ignored, got %+v", row)
	}
}

func TestValidateRow_Valid(t *testing.T) {
	err := model.ValidateRow(rowDeviceFields)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestValidateRow_TwoTitles(t *testing.T) {
	fields := model.Fields{
		{Name: "a", Row: model.RowTitle},
		{Name: "b", Row: model.RowTitle},
	}
	err := model.ValidateRow(fields)
	if err == nil || !strings.Contains(err.Error(), "two fields marked RowTitle: a, b") {
		t.Errorf("expected error containing 'two fields marked RowTitle: a, b', got %v", err)
	}
}

func TestValidateRow_TwoBadges(t *testing.T) {
	fields := model.Fields{
		{Name: "title", Row: model.RowTitle},
		{Name: "b1", Row: model.RowBadge},
		{Name: "b2", Row: model.RowBadge},
	}
	err := model.ValidateRow(fields)
	if err == nil || !strings.Contains(err.Error(), "two fields marked RowBadge") {
		t.Errorf("expected error containing 'two fields marked RowBadge', got %v", err)
	}
}

func TestValidateRow_DetailWithoutTitle(t *testing.T) {
	fields := model.Fields{
		{Name: "detail", Row: model.RowDetail},
	}
	err := model.ValidateRow(fields)
	if err == nil || !strings.Contains(err.Error(), "need a field marked RowTitle") {
		t.Errorf("expected error containing 'need a field marked RowTitle', got %v", err)
	}
}

func TestValidateRow_RefAndOptions(t *testing.T) {
	def := &model.Definition{}
	fields := model.Fields{
		{Name: "zone_id", Ref: def, Options: def},
	}
	err := model.ValidateRow(fields)
	if err == nil || !strings.Contains(err.Error(), "both Ref and Options: zone_id") {
		t.Errorf("expected error containing 'both Ref and Options: zone_id', got %v", err)
	}
}

func TestValidateRow_NoRowIsValid(t *testing.T) {
	fields := model.Fields{
		{Name: "id", Type: model.Text()},
		{Name: "name", Type: model.Text()},
	}
	err := model.ValidateRow(fields)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
