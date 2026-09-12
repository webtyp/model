---
PLAN: "fix!: FielderSlice stops embedding Fielder — a list has no columns"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Phase A (GATE)** of
> [`LIST_CONTRACT_MASTER_PLAN.md`](https://github.com/webtyp/docs/blob/main/LIST_CONTRACT_MASTER_PLAN.md).
> `webtyp/ormc` (phase B) cannot start before this ships a tag.

# Plan — `webtyp.com/model`: a list is not a row, so stop making it answer like one

## 0. Context (verified against the repo — do not re-diagnose)

`FielderSlice` embeds `Fielder`:

```go
type Fielder interface {
	Schema() []Field
	Pointers() []any
}

type FielderSlice interface {
	Fielder          // ← this line is the defect
	Len() int
	At(i int) Fielder
	Append() Fielder
}
```

A list is a sequence of rows; it has no columns of its own. The schema belongs
to the element, reached through `At(i)`/`Append()`. So every generated list type
is forced to answer with a lie — `ormc` emits this 96 times across the monorepo:

```go
func (s *UserList) Schema() []model.Field { return nil }
func (s *UserList) Pointers() []any       { return nil }
```

Nothing calls them: `json/encode.go` and `json/decode.go` reach the rows via
`Len()`/`At()`/`Append()` and type-assert the **element**, never the list.

The harm is that having them makes the lie true for the compiler. A list
satisfies `model.Fielder`, so `Accepts(&auth.UserList{})` compiles, and
`mcp/tool_schema.go` believes it:

```go
func inputSchemaOf(m model.Fielder) string {
	fields := m.Schema()
	if len(fields) == 0 { return EmptyInputSchema }
```

The MCP tool is then published **advertising that it takes no arguments** — no
error, no log. This plan makes that state unrepresentable.

**This is not a size optimization.** It was measured first: removing the two
stubs saves ~27 bytes per list type, which on a real 1.140.006-byte WASM client
carrying 10 list types is 0,02 %. Do not justify or scope this change by binary
size — see §2 of the master plan.

**Anti-footgun.** Narrowing an interface cannot break an implementer: a type
with extra methods still satisfies a smaller interface. So this change alone
must NOT touch any generated `_orm.go`, and must NOT be "completed" by deleting
stubs anywhere else — that is phases B and C, in other repos. Do not add a
`replace`, do not vendor, do not edit consumers from here.

## Design gate (api-design — five answers)

### 1. Prior art

| Concern | Frameworks | Why we differ |
|---|---|---|
| Collection vs record contract | **Go stdlib `sort.Interface`** (`Len`/`Less`/`Swap` — a collection contract that says nothing about the element's shape) | Exactly the shape we are moving to: the collection exposes traversal, the element exposes itself. We were the odd one out. |
| Collection vs record contract | **Java** `Collection<E>` vs `E`, **C#** `IEnumerable<T>` vs `T` | Neither makes the collection implement the element's interface. A `List<Person>` is not a `Person`. Our `FielderSlice: Fielder` said it was. |
| Schema introspection over a set | **encoding/json** (`reflect` walks the element), **GORM** (`Model(&User{})` names the record even when scanning into a slice) | Both keep "what are the columns" a question about the record. We do too — we just stopped also asking the slice. |

### 2. Novice-name test

No new name. The change is a deletion: `FielderSlice` read aloud is "a slice of
Fielders", and a slice of things is not itself one of those things. The current
shape is what a junior cannot predict.

### 3. Complexity ledger

```
Concepts the developer must learn   +0 / −1 (nobody learns "a list reports nil columns")
Files they must touch to do X       +0 / −0
Lines at the call site              +0 / −0   (consumers unchanged; only generated code shrinks)
Ways to do the same thing           +0 / −1 (was: ask the list OR the element for a schema)
```

### 4. Where it belongs

`model` owns both interfaces. The fix is one line here; anywhere else it would
be a workaround for this package's shape.

### 5. What it deletes

- `Fielder` from `FielderSlice`'s method set, and with it `ModelSlice`'s
  (it embeds `FielderSlice`).
- Downstream, in later phases: 96 `Schema() nil` + 101 `Pointers() nil`.

## Quality rules

```
RULE: no stdlib in this package beyond what it already imports — webtyp/fmt only.
RULE: every repeated string is a named constant; string literals forbidden in logic.
RULE: do not touch any *_orm.go file in this repo or any other — phases B and C.
```

## Stage 1 — narrow `FielderSlice`

**File:** `interface.go`.

```go
// FielderSlice is implemented by generated code to allow
// iteration over a slice of structs without reflection.
//
// It deliberately does NOT embed Fielder: a list is a sequence of rows and has
// no columns of its own. The schema belongs to the element, reached through
// At(i) and Append(). While it did embed Fielder, every generated list had to
// answer Schema() with nil — which made a list satisfy Fielder, so it could be
// passed where a record was required and the reader silently saw "no fields".
type FielderSlice interface {
	Len() int
	At(i int) Fielder
	Append() Fielder
}
```

Remove **only** the embedded `Fielder` line and add the paragraph above. Leave
`Fielder`, `ModelSlice`, `Model` and every other declaration untouched —
`ModelSlice` inherits the narrowing because it embeds `FielderSlice`.

## Stage 2 — a test that pins the contract

**File:** `interface_test.go` (new; this package currently has only
`rbac_test.go`).

The point of this change is that an illegal state stops compiling, so the test
must assert the shape, not a runtime value.

1. Declare a list type that implements ONLY the narrowed contract — no
   `Schema`, no `Pointers` — and assert it satisfies `FielderSlice`:
   ```go
   type onlySlice []*probeRow

   func (s *onlySlice) Len() int               { return len(*s) }
   func (s *onlySlice) At(i int) model.Fielder { return (*s)[i] }
   func (s *onlySlice) Append() model.Fielder  { v := &probeRow{}; *s = append(*s, v); return v }

   var _ model.FielderSlice = (*onlySlice)(nil)
   ```
   (`probeRow` is a minimal `Fielder`: a struct with a real `Schema()` and
   `Pointers()`.) This line failing to compile is the regression signal — it is
   what proves `Fielder` is no longer required.
2. Assert the element still carries the schema, so the narrowing did not move
   the capability somewhere useless:
   ```go
   var s onlySlice
   if got := len(s.Append().Schema()); got != 2 {
       t.Errorf("element schema = %d fields, want 2", got)
   }
   ```
3. Add a comment on the file stating that a list must NOT be given `Schema()` or
   `Pointers()` "to be helpful" — that is the defect this plan removed.

Write the test in the package's existing style (`rbac_test.go` is the
reference). Do not import anything outside `webtyp/fmt` and `testing`.

## Acceptance criteria

1. `go build ./...`, `go vet ./...`, `go test ./...` green.
2. `grep -n "Fielder" interface.go` → `FielderSlice` no longer lists it among
   its embedded interfaces (it still appears in `At`/`Append` signatures and in
   `Model`, which is correct).
3. `grep -rn "_orm.go" .` → this change touched none.
4. `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` → only hits that
   predate this change.

## Out of scope

- `ormc` still emitting the two stubs — phase B. Generated code that still has
  them keeps compiling: extra methods never break a narrower interface.
- Regenerating any consumer — phase C.
- Splitting the codec entry points so lists stop being `Encodable` — measured at
  0,12 % of a real binary and **rejected**; do not do it here.

| Stage | Files | Action |
|---|---|---|
| 1 | `interface.go` | `FielderSlice` stops embedding `Fielder`; document why |
| 2 | `interface_test.go` (new) | compile-time proof that a list needs no schema |
