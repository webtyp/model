---
PLAN: "feat: FieldDB.Revision — the column that counts a row's updates, for conditional updates (decision D48)"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit after every stage, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## Goal

Two requests that read the same row and update it must not both win. An entity declares the column that counts its updates; `crud` rejects an update that carries an older value (409) and the store applies it only if the column still holds that value.
This repo adds **only the declaration** (data, isomorphic): the flag and its accessor. Checking a declaration is `definitioncheck`'s job; the behaviour is `crud`'s and `orm`'s.

## Ecosystem search

Searched `Revision` over `webtyp/*` and `veltylabs/*`: `appointment_booking` and `payment_manager` keep a hand-written `Revision` field and compare it by hand (`booking.go`, `repository.go`: `ErrConflict`). Decision: **extend** `FieldDB` (the database metadata of a field, where `PK`, `Unique` and `AutoInc` already live).

## Design gate

1. **Prior art**: optimistic locking with a version column (JPA `@Version`, Rails `lock_version`).
2. **Novice-name test**: `FieldDB{Revision: true}` on the integer field `revision` = «this column is the row's revision».
3. **Ledger**: concepts +1 (`Revision`); hand-written revision comparisons in modules: −N (each module deletes its own).
4. **Where**: `model` owns field metadata; `docs/diagrams/PLACEMENT.md` is unchanged.
5. **Deletes**: nothing here.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Stage 1 — failing test first (`tests/field_revision_test.go`, package `model_test`)

| Test | setup → call → expected |
|---|---|
| `TestField_IsRevision` | `Field{Name: "revision", Type: Int(), DB: &FieldDB{Revision: true}}.IsRevision()` → true; a field with `DB: nil`, with `DB: &FieldDB{PK: true}` and with `DB: &FieldDB{Unique: true}` → false |

Add its row to `docs/TESTS.md` (format: `| [TestField_IsRevision](../tests/field_revision_test.go) | Field IsRevision |`). Run `gotest`: it fails to compile. Commit: `test: Field.IsRevision (red)`.

## Stage 2 — the code (`field.go`)

In `FieldDB`, after `AutoInc bool`:

```go
	// Revision marks the integer column that counts the row's updates. crud writes 1 when the row is created, rejects an update that carries an
	// older value (409) and writes value+1 on each update; a client never writes it directly. At most one per entity, on an Int field (definitioncheck).
	//
	//	{Name: "revision", Type: model.Int(), DB: &model.FieldDB{Revision: true}}
	Revision bool
```

After `IsUnique`:

```go
// IsRevision returns true if the field is the row's revision counter.
func (f Field) IsRevision() bool { return f.DB != nil && f.DB.Revision }
```

`gotest` passes. Commit: `feat: FieldDB.Revision and Field.IsRevision`.

## Stage 3 — documentation

`README.md` and `docs/ARCHITECTURE.md`: one paragraph and a literal example of the revision field (the line above), saying who enforces it (`definitioncheck` judges, `crud` and `orm` apply). No `lang.json` change (no text added).

## Delivery check

1. `gotest` passes. 2. `git diff --stat main` lists only `field.go`, `tests/field_revision_test.go`, `docs/TESTS.md`, `README.md`, `docs/ARCHITECTURE.md`. 3. `lang.json` is unchanged.
