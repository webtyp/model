---
PLAN: "fix: model.FieldError unwraps its cause; NotNull on a Bool field accepts false"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 11594222981559106768
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> **Publishing:** you publish only by OPENING THE PULL REQUEST. Submit after Stage 1 and keep committing.
> **Never ask a yes/no question about progress.**
> **Before you submit:** `git status` must be clean and `git diff --stat main` must list the files of the stages.

# Plan: two small defects of `model`

## Why (read once)

1. `model.FieldError` wraps a cause (`Err`) but has no `Unwrap`, so `errors.Is(err, ErrX)` never sees through it: callers that wrap a sentinel in a `FieldError` cannot test for it.
2. `NotNull` on a `Bool` field rejects `false` because `validateFields` treats the zero value of every non-text field as missing; `false` is a value. Modules had to leave `NotNull` off booleans (`is_active`) and say so in a comment.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```
No new dependency, no `replace`. Read files by ranges (`sed -n 1,80p file`).

## Stage 1 — `error.go`: `Unwrap`

After `func (e *FieldError) FailedField() string { return e.Field }` (line 62) add:

```go
// Unwrap answers the cause, so errors.Is and errors.As see through the field.
func (e *FieldError) Unwrap() error { return e.Err }
```

## Stage 2 — `field.go`: a Bool is never «missing»

In `validateFields`, in the `default:` branch (the lines `// Non-text fields: only check NotNull (zero value check)` and the `if` under it), change the condition from

```go
			if field.NotNull && IsZeroPtr(ptrs[i], ft) {
```

to

```go
			if field.NotNull && ft != FieldBool && IsZeroPtr(ptrs[i], ft) { // false is a value: NotNull on a bool only means the column is NOT NULL
```

and put the same sentence in the doc comment of `Field.NotNull` (`field.go:100`). Do not touch `IsZeroPtr` (other callers use it: the primary-key check stays as it is).

## Stage 3 — tests (`tests/`, one row each in `docs/TESTS.md`, `| [Name](../tests/<file>) | rule |`)

`tests/error_test.go` (append; add the imports `errors` and `webtyp.com/model` if missing):

```go
func TestFieldError_UnwrapsItsCause(t *testing.T) {
	cause := errors.New("email registered")
	var err error = &model.FieldError{Field: "email", Err: cause}
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is must see the cause through a FieldError")
	}
}
```

`tests/field_test.go` (append; same package `tests`):

```go
type flagRec struct{ Active bool }

func (r *flagRec) IsNil() bool { return r == nil }
func (r *flagRec) Schema() []model.Field {
	return []model.Field{{Name: "active", Type: model.Bool(), NotNull: true}}
}
func (r *flagRec) Pointers() []any                     { return []any{&r.Active} }
func (r *flagRec) EncodeFields(w model.FieldWriter)    { w.Bool("active", r.Active) }
func (r *flagRec) DecodeFields(rd model.FieldReader) {
	if v, ok := rd.Bool("active"); ok {
		r.Active = v
	}
}

type countRec struct{ N int64 }

func (r *countRec) IsNil() bool { return r == nil }
func (r *countRec) Schema() []model.Field {
	return []model.Field{{Name: "n", Type: model.Int(), NotNull: true}}
}
func (r *countRec) Pointers() []any                  { return []any{&r.N} }
func (r *countRec) EncodeFields(w model.FieldWriter) { w.Int("n", r.N) }
func (r *countRec) DecodeFields(rd model.FieldReader) {
	if v, ok := rd.Int("n"); ok {
		r.N = v
	}
}

func TestNotNull_BoolAcceptsFalse(t *testing.T) {
	if err := model.ValidateFields(model.ActionCreate, &flagRec{Active: false}); err != nil {
		t.Fatalf("false is a value, NotNull must accept it: %v", err)
	}
}

func TestNotNull_IntStillRejectsZero(t *testing.T) {
	if err := model.ValidateFields(model.ActionCreate, &countRec{N: 0}); err == nil {
		t.Fatal("NotNull on an int must still reject zero")
	}
}
```
(If `FieldWriter.Int` / `FieldReader.Int` have another exact name, `go doc webtyp.com/model FieldWriter` shows it: use the real one, say which in the PR.)

## Delivery check

`gotest` passes on every destination; the three tests exist with their rows; the pull request is open. A tree that does not compile or whose tests did not run is a failed delivery.
