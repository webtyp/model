package model

import "webtyp.com/fmt"

// RowSlot is where a field shows in a list row. The zero value is "not shown".
type RowSlot uint8

const (
	RowNone   RowSlot = iota // not in the row
	RowTitle                 // the main text; exactly one per row
	RowDetail                // secondary text; several are joined with " · "
	RowBadge                 // the leading badge (a date, an hour); at most one
)

// rowDetailSeparator joins several RowDetail values.
const rowDetailSeparator = " · "

// Row is the projection of one record as a list row, built from its schema by RowOf.
type Row struct {
	ID     string // the primary key, as text
	Title  string
	Detail string
	Badge  string
}

const (
	errRowTwoTitles  = "model: two fields marked RowTitle: "
	errRowTwoBadges  = "model: two fields marked RowBadge: "
	errRowNoTitle    = "model: fields marked RowDetail or RowBadge need a field marked RowTitle"
	errRefAndOptions = "model: a field cannot have both Ref and Options: "
)

// ValidateRow reports the first row rule fields break: at most one RowTitle and one
// RowBadge, a RowTitle whenever any field is in the row, never Ref and Options on
// the same field. ormc calls it when it generates a model; a renderer may call it
// at startup. Example: if err := model.ValidateRow(ServiceModel.Fields); err != nil { … }
func ValidateRow(fields []Field) error {
	title, badge, inRow := "", "", false
	for _, f := range fields {
		if f.Ref != nil && f.Options != nil {
			return fmt.Err(errRefAndOptions + f.Name)
		}
		switch f.Row {
		case RowTitle:
			if title != "" {
				return fmt.Err(errRowTwoTitles + title + ", " + f.Name)
			}
			title = f.Name
		case RowBadge:
			if badge != "" {
				return fmt.Err(errRowTwoBadges + badge + ", " + f.Name)
			}
			badge = f.Name
			inRow = true
		case RowDetail:
			inRow = true
		}
	}
	if inRow && title == "" {
		return fmt.Err(errRowNoTitle)
	}
	return nil
}

// RowOf builds the list row of a record from its schema: the primary key as ID and
// each field marked Row in its slot. Example: row := model.RowOf(&device)
func RowOf(f Fielder) Row {
	schema := f.Schema()
	vals := ReadValues(schema, f.Pointers())
	var r Row
	for i, field := range schema {
		if vals[i] == nil {
			continue
		}
		text := fmt.Convert(vals[i]).String()
		if field.IsPK() && r.ID == "" {
			r.ID = text
		}
		switch field.Row {
		case RowTitle:
			r.Title = text
		case RowBadge:
			r.Badge = text
		case RowDetail:
			if text == "" {
				continue
			}
			if r.Detail != "" {
				r.Detail += rowDetailSeparator
			}
			r.Detail += text
		}
	}
	return r
}
