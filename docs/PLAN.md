---
PLAN: "feat: Field.Label and Field.Help — the English display texts of a field"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — model: `Field.Label` and `Field.Help`

Phase **T1** of the master plan `SOURCE_SELECTION_MASTER_PLAN.md` (orchestration only — everything
this plan needs is inline). `webtyp/form` (phase T2) depends on the tag this plan produces.

Read [AGENTS.md](../AGENTS.md) first (WASM rules, tests in `tests/`, display text in English).

## Why

A form shows each field's label. Today `model.Field` has no place for it, so `webtyp/form` shows the
column name: a device form in mjosefa-cms reads `Name`, `Ip`, `Type`, `Is_active`, and the checkbox
says `is_active`. With translations moving to data (`config/lang.json`, see
`https://github.com/webtyp/lang/blob/main/docs/PLAN.md`), the label must be an **English text**
declared next to the field. That text is the translation key: `form` shows it through
`lang.Translate`, and the `langc` generator collects it as a key.

Fields also need a **persistent help text** ("Format: 12.345.678-9", "Only if the patient is a
minor"). A placeholder cannot carry it: it disappears while typing, has low contrast, and screen
readers announce it inconsistently. W3C/WAI, Nielsen Norman Group and GOV.UK all say a placeholder
shows an example, never instructions. The placeholder stays what it is today: a format example owned
by the input type (`input.IP()` → `example: 192.168.1.1`). Instructions go in `Help`.

## Design gate

1. **Prior art.**
   - Django `verbose_name` on a model field: a human label used by forms and admin, and wrapped
     for translation.
   - Rails `human_attribute_name` (labels from locale files keyed by attribute).
   - Laravel `$attributes` labels in language files.
   - Go `gorm`/`ent` have no label, so UIs re-derive one.

   - For help text: Django `help_text`, Material "hint", GOV.UK "hint text".

   We follow Django: label and help beside the field, in the source language, translated at display
   time.
2. **Novice-name test.** `Label: "IP address"` — "the field's label". `Help: "Format: 12.345.678-9"`
   — "the field's help". Both are the words form APIs use.
3. **Complexity ledger.**
   ```
   Concepts the developer must learn   +2 (Label, Help)
   Files they must touch to do X       +0 (the label lives where the field is declared)
   Lines at the call site              +1 per labelled field (optional: no Label → humanised Name), +1 per field with help
   Ways to do the same thing           +0
   ```
4. **Where it belongs.** `model.Field` is the single schema that `form` already reads, so the label
   belongs there, not in `form` or a side table.
5. **What it deletes.** Nothing here. In `form` (phase T2) it replaces the fallback "placeholder as
   label".

## Stage 1 — the field

In `field.go`, add to `Field`, after `Name`:

```go
	// Label is the text a person sees for this field, in ENGLISH
	// ("IP address", "Active"). It is a translation key for webtyp.com/lang:
	// UIs show it through lang.Translate, and the langc generator collects it.
	// Empty → UIs derive it from Name ("is_active" → "is active").
	Label string
	// Help is a persistent instruction shown under the field, in ENGLISH
	// ("Format: 12.345.678-9"). It is a translation key, like Label. Put
	// instructions here, never in a placeholder (which only shows an example
	// and disappears while typing). Empty → no help is shown.
	Help string
```

No other code changes: `Schema()` already returns the `Definition`'s `Fields`, so the value reaches
every consumer at run time. `ormc` parses `Definition` literals and silently ignores keys it does
not handle, so it needs no change. Do NOT add `Label` handling to anything in this repo besides the
field itself.

## Stage 2 — test and docs

- `tests/field_test.go`: a `Definition` whose field has `Label: "IP address"` and
  `Help: "Format: 192.168.1.1"`; its `Fields[i].Label` and `.Help` equal those texts, and a field
  without them has both `""`.
- `docs/API_FIELD.md` and the README's Quick Start: one example field with `Label` and `Help`. Add:
  "English text; it is the translation key — never write another language here. Instructions go in
  Help, never in a placeholder".

## Acceptance

- `gotest` passes (includes WASM).
- Every `*_test.go` at the root starts with `// Root-level test (justified):` (today:
  `interface_test.go`, `rbac_test.go` — add the comment if they truly need unexported identifiers,
  otherwise `git mv` them to `tests/` as `package model_test`).

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Field | `field.go` |
| 2 | Test and docs | `tests/field_test.go`, `docs/API_FIELD.md`, `README.md`, root `*_test.go` |
