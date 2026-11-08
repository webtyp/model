---
PLAN: "feat: Definition.Allow (closed by default) and composite-key ids (KeySeparator, JoinKey, SplitKey)"
EXECUTOR: antigravity
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit after every stage and stop. Never ask whether to
> proceed. This plan runs **locally** and is **published as soon as the delivery check passes** (no `replace` anywhere).
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

# Plan: `model.Definition.Allow`

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Why

Decisions D19/D20 of `webtyp/docs/ARCHITECTURE_MASTER_PLAN.md`: the framework persists a table and a module only
declares which actions the table offers through the API. One declaration read by the server (`crud`, via code `ormc`
generates) and by the screen (`view`), replacing `view.Ops.Allow` and the four hand-mounted operations per entity.

## Design gate

1. **Prior art:** Rails `resources :posts, only: [:create, :show]` (https://guides.rubyonrails.org/routing.html#restricting-the-routes-created),
   Laravel `Route::resource(...)->only([...])` (https://laravel.com/docs/11.x/controllers#restful-partial-resource-routes),
   Spring Data REST `@RepositoryRestResource(exported = false)` (https://docs.spring.io/spring-data/rest/reference/customizing/configuring-the-rest-url-path.html).
   All: a resource lists the actions it offers; the rest do not exist.
2. **Name:** `Allow` — the word `view.Ops.Allow` already uses for the same concept; «PostModel allows create, read, update».
3. **Ledger:** concepts ±0 (the concept moves from `view.Ops` to the model); ways to do the same thing −1 (A6.5 deletes `view.Ops.Allow`).
4. **Where:** `model` owns `Definition` and `Action`.
5. **Deletes:** nothing here; enables the deletions of A6.2–A6.5.

## Stage 1 — `definition.go`

```go
type Definition struct {
	Name   string
	Fields Fields
	Wire   bool
	// Allow is what the table offers through the API (model.Create | model.Read | …). Zero = nothing: the table is
	// internal to its module. Read is not implied: a table only the module reads leaves it out.
	//
	//	var PostModel = model.Definition{Name: "post", Allow: model.Create | model.Read | model.Update, Fields: …}
	Allow Action
}

// Allower is implemented by the struct ormc generates from a Definition: what the table offers through the API.
// crud mounts only these actions and a screen shows only their buttons.
type Allower interface {
	Allow() Action
}

// TableSlice is a list of a table's rows that also tells what the table allows. ormc implements it on <Entity>List,
// so a screen asking for it does not compile with a list generated before Allow existed.
//
//	view.NewCallerLister(caller, ops, func() model.TableSlice { return &PatientList{} })
type TableSlice interface {
	ModelSlice
	Allower
}
```

Composite primary keys (D22) — add to `definition.go`:

```go
// KeySeparator joins the values of a composite primary key into the one id a record has on the wire, in schema order.
// A key value never contains it (crud rejects it on create).
const KeySeparator = "/"

// JoinKey is the wire id of a composite key: JoinKey("r1", "e7") == "r1/e7".
func JoinKey(parts ...string) string

// SplitKey is the inverse of JoinKey; ok is false when id does not have exactly n parts or a part is empty.
//
//	room, equipment, ok := …  // parts, ok := model.SplitKey("r1/e7", 2) → ["r1","e7"], true
func SplitKey(id string, n int) (parts []string, ok bool)
```
(write them with `webtyp.com/fmt` helpers, no stdlib `strings`). Tests: `TestKey_JoinSplitRoundTrip`, `TestKey_SplitWrongCountNotOK` (`SplitKey("a/b/c", 2)` → false), `TestKey_SplitEmptyPartNotOK` (`"a/"` → false).

Add a check to the existing `Definition` validation used by `ormc`/tests (if `Definition` has a `Validate` or the repo
has a definition checker; `grep -n "func (d Definition)" *.go`): `Allow != 0` together with `Wire: true` is an error —
`ErrAllowOnWire = "model: a Wire definition is a transport format, it cannot Allow table actions; remove Allow or Wire"`.
If no such checker exists, add none (ormc enforces it in A6.4).

## Stage 2 — tests (`tests/definition_test.go`, row in `docs/TESTS.md`)

| Test | Expected |
|---|---|
| `TestDefinition_AllowZeroIsClosed` | `Definition{}.Allow.Has(Read)` → false |
| `TestDefinition_AllowHas` | `Definition{Allow: Create \| Read}.Allow.Has(Create)` → true, `.Has(Delete)` → false |

## Stage 3 — documentation

`docs/ARCHITECTURE.md`: one paragraph «Allow — what a table offers», linking D20 by behaviour (no master link).
`README.md`: add `Allow` to the `Definition` example.

## Delivery check

`gotest`. A tree that does not compile or whose tests did not run is a failed delivery.

| Stage | Files |
|---|---|
| 1 | `definition.go` |
| 2 | `tests/definition_test.go`, `docs/TESTS.md` |
| 3 | `docs/ARCHITECTURE.md`, `README.md` |
