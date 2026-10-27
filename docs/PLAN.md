---
PLAN: "feat: Field.Row and Field.Options — a field says where it shows in a list row and which list fills its options"
TAG: v0.2.7
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 2373001823178949158
PR: https://github.com/webtyp/model/pull/7
REVIEWED_COMMIT: 0228cca309f2b282e25889d3956d2f5fc92209a1
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp.com/model`: the list row is declared on the fields

## Why

A list row (title, detail, badge) is today built by a hand-written method on each record (`Item()` in every
module's `view.go`). Nothing ties a row text to the field it comes from, so the framework cannot keep a row in step
with what is saved, and a module also repeats by hand which list fills each select (`view.Lookup`). From now on the
`model.Definition` says it, on each field:

```go
{Name: "name", Type: input.Text(), Label: "Name", Row: model.RowTitle},
{Name: "sku", Type: input.Text(), Label: "SKU", Row: model.RowDetail},
{Name: "zone_id", Type: input.Select(), Label: "Zone", Options: &ZoneModel}, // options without a foreign key
```

`Schema()` returns these very `Fields` at runtime (`ormc` generates `func (m *X) Schema() []model.Field { return
XModel.Fields }`), so `RowOf` builds the row from the schema and the pointers, with no generated code per model.
`ValidateRow` is the one rule; `ormc` calls it at generation time (its own plan) and renderers may call it at startup.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

## Ecosystem search

- `model.ReadValues(schema, ptrs)` (`field.go:365`) reads every field value without reflection → **reuse** in `RowOf`.
- `Field.Ref *Definition` (`field.go:111`) already names the table a foreign key points to → a select with `Ref`
  takes its options from that table; `Options` is the same pointer **without** a foreign key.
- `view.Item{ID, Label, Description, LeadTop, LeadMain, LeadBottom}` (`webtyp/view`) is the hand-built row it will
  replace (a later `view` plan). Nothing in `webtyp/*` or `veltylabs/*` declares row slots on fields → **new** here.
- `fmt.Convert(v).String()` (`webtyp.com/fmt`) is how `form` turns a field value into text (`form/load.go:28`) → reuse.

## Design gate

1. **Prior art.** Admin frameworks declare the row by field: Laravel Nova (`->showOnIndex()` on the field), Strapi's
   «main field», Payload CMS `useAsTitle: 'name'`, Sanity `preview.select {title, subtitle}`, Django `list_display`.
   We mark the field itself (Nova/Strapi style): a closed set, no field name written twice.
2. **Novice-name test.** `Row: model.RowTitle` reads «in the row, this is the title»; `Options: &ZoneModel` reads
   «its options are the zones». `model.RowOf(record)` — «the row of this record».
3. **Complexity ledger.** Concepts +3 (`RowSlot`, `Row`, `Options`) / −3 later (`Item()`, `view.Itemizer`,
   `view.Lookup`); lines per module −10 to −30 later; ways to build a row −1 later.
4. **Where it belongs.** `model` owns the field declaration and already reads values through the schema
   (`ValidateFields`, `ReadValues`); the row is a fact about the record, not about a renderer.
5. **What it deletes.** Nothing here; later plans delete `Item()` in 16 modules, `view.Itemizer` and `view.Lookup`.

## Stage 1 — the field slots (`field.go`)

Add to `Field`, after `Ref`:

```go
	// Row says where this field shows in a list row of its record: RowTitle (the
	// one main text), RowDetail (secondary text; several RowDetail fields are joined
	// with " · " in declaration order) or RowBadge (the leading badge, e.g. a date or
	// an hour). The zero value keeps the field out of the row. See RowOf.
	Row RowSlot
	// Options is the Definition whose records fill this field's choices (a select)
	// when the field is NOT a foreign key; a foreign key says it with Ref, which
	// fills the choices too. Setting both is an error (ValidateRow).
	Options *Definition
```

## Stage 2 — `row.go` (new)

```go
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
```

If `fmt.Convert(x).String()` does not exist with that shape in `webtyp.com/fmt`, stop and name it in the PR.

## Stage 3 — tests (`tests/row_test.go`, new, `package model_test`)

Write a test record by hand (no generator here):

```go
type rowDevice struct{ Id, Name, Access, Location string; Port int64 }

var rowDeviceFields = model.Fields{
	{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
	{Name: "name", Type: model.Text(), Row: model.RowTitle},
	{Name: "access", Type: model.Text(), Row: model.RowDetail},
	{Name: "location", Type: model.Text(), Row: model.RowDetail},
	{Name: "port", Type: model.Int()},
}

func (d *rowDevice) Schema() []model.Field { return rowDeviceFields }
func (d *rowDevice) Pointers() []any       { return []any{&d.Id, &d.Name, &d.Access, &d.Location, &d.Port} }
```

| Test | Input | Expected |
|---|---|---|
| `TestRowOf_TitleAndJoinedDetails` | `{Id:"a1", Name:"Cesar", Access:"local", Location:"Oficina"}` | `Row{ID:"a1", Title:"Cesar", Detail:"local · Oficina"}` |
| `TestRowOf_EmptyDetailSkipped` | `Location: ""` | `Detail == "local"` (no trailing separator) |
| `TestRowOf_NotInRowIgnored` | `Port: 8080` | no field of `Row` contains `"8080"` |
| `TestValidateRow_Valid` | `rowDeviceFields` | `nil` |
| `TestValidateRow_TwoTitles` | two `RowTitle` fields `a`, `b` | error containing `two fields marked RowTitle: a, b` |
| `TestValidateRow_TwoBadges` | two `RowBadge` + one title | error containing `two fields marked RowBadge` |
| `TestValidateRow_DetailWithoutTitle` | one `RowDetail`, no title | error containing `need a field marked RowTitle` |
| `TestValidateRow_RefAndOptions` | field `zone_id` with `Ref: &d` and `Options: &d` | error containing `both Ref and Options: zone_id` |
| `TestValidateRow_NoRowIsValid` | fields without `Row` | `nil` |

## Stage 4 — docs

- `README.md`: a short section «List rows and select options» with the three-line example of «Why» and `RowOf`.
- `docs/` architecture file (if present): one paragraph — the row is declared on the fields, read through the
  schema, validated by `ValidateRow` (also at generation time by `ormc`).

## Stage 5 — translations

`langc sync .`, then fill every empty value it reports in `lang.json` for every language in `"languages"`. Zero
untranslated keys and zero reuse warnings.

## Delivery

- `gotest` green. A PR that does not compile, or whose tests could not run, is a failed delivery.
  Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.
- No `map`, no `reflect`, no stdlib `fmt`/`strings` in non-test code (this package compiles to WASM).
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `field.go` | `Row`, `Options` on `Field` |
| 2 | `row.go` | `RowSlot`, `Row`, `ValidateRow`, `RowOf` |
| 3 | `tests/row_test.go` | behaviour |
| 4 | `README.md`, `docs/*` | document |
| 5 | `lang.json` | translate |
