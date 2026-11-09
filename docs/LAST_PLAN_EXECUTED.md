---
PLAN: "feat: model.Rules, Definition.Rules/MFA and model.ModuleDefinition — the data a module declares (contracts only; the judge lives in definitioncheck)"
EXECUTOR: antigravity
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit after every stage and stop. Never ask whether to
> proceed. This plan runs **locally**: do not tag, push or publish anything; the owner reviews first.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## Goal

A module's `definition.go` (the file `definitionc` reads, formerly `model.go`) declares, for **each action in `Allow`**, whether the entity has a business rule
of its own (`AddMyBusinessRule`) or crud's checks are enough (`HaveNoBusinessRule`), which allowed actions also require a second factor (`MFA`), and the module
itself (`ModuleDefinition`). This repo adds **only these data types** (contracts, isomorphic: they compile into the browser and the server alike).
**No checking code lives here** (decision D44): the rule "every allowed action declares its rule" is judged by the `definitioncheck` library, which `definitionc` and `crud` call.
`model` stays free of tooling and of the error texts that would otherwise enter the dictionary of every WASM consumer.

## Ecosystem search

Searched `Rules`, `ModuleDefinition`, `Entities`, `HaveNoBusinessRule`, `AddMyBusinessRule` over `webtyp/*` and `veltylabs/*` (`grep -rn`): only
`crud.Table` hooks (the consumer, changed in its own plan) and `Definition.Allow`/`model.Allower` (`definition.go`) exist.
Decision: **extend** `model` (the single schema repo) with data only; the judge is a new library (`definitioncheck` plan).

## Design gate

1. **Prior art** (unverified in this session; the owner chose the shape): Django REST `ModelViewSet` — declare the model, hook only what differs; protobuf — the schema is data,
   the checks are a separate tool. Difference here: no `reflect`/generics, so the choice is declared as constants in a literal `definitionc` reads statically.
2. **Novice-name test**: `model.AddMyBusinessRule` = "I write a business rule for this action"; `model.HaveNoBusinessRule` = "this action has no business rule of mine";
   `Rules` = the rule choice per action; `ModuleDefinition.Entities` = the entities the module serves.
3. **Ledger**: concepts +3 (`RuleMode`, `Rules`, `ModuleDefinition`); files to touch to offer an action: +1 line in `Rules` per action
   (replaces the hand-written mount, removed in the `definitionc` plan); ways to do the same: −1 (the hand-written `MountOperations`).
4. **Where**: `model` is the one schema repo and holds data; the judge is `definitioncheck`; `crud` and `definitionc` read these types.
5. **Deletes**: nothing in this repo; the `definitionc` plan deletes `ops.go` mounting in every module.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## APIs used (all inside this repo — read them before writing)

| symbol | exact signature | file |
|---|---|---|
| `Action` | `type Action uint8`; constants `Create, Read, Update, Delete`; `AllActions` | `rbac.go:83` |
| `Definition` | struct with `Name string`, `Fields Fields`, `Wire bool`, `Allow Action` | `definition.go:12` |

If a signature differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — the code (no tests: these are data types; their behaviour is judged and tested in `definitioncheck`)

`definition.go`: add to `Definition`, after `Allow`:

```go
	// Rules says, for each action in Allow, whether the entity has a business rule of its own (AddMyBusinessRule) or crud's
	// checks are enough (HaveNoBusinessRule). An allowed action left undeclared is an error (definitioncheck), never a default.
	//
	//	Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule}
	Rules Rules
	// MFA is the set of allowed actions that also require a second factor. It must be a subset of Allow (definitioncheck).
	//
	//	MFA: model.Update | model.Delete
	MFA Action
```

New file `rules.go` (no build tag, no import; compiles to WASM, no `map`):

```go
package model

// RuleMode is what one allowed action does about the business. The zero value means "undeclared" and is refused by definitioncheck.
type RuleMode uint8

const (
	// AddMyBusinessRule: the entity has a business rule of its own, run before crud stores; definitionc creates its rule file to complete.
	AddMyBusinessRule RuleMode = iota + 1
	// HaveNoBusinessRule: no business rule of mine; crud validates the schema and stores, nothing to write.
	HaveNoBusinessRule
)

// Rules declares a RuleMode per action. Each field is for the action of the same name.
type Rules struct {
	Create RuleMode
	Read   RuleMode
	Update RuleMode
	Delete RuleMode
}

// Of returns the mode declared for exactly one action; a set of actions, or anything that is not one of the four, returns 0.
//
//	Rules{Create: AddMyBusinessRule}.Of(Create) // AddMyBusinessRule
func (r Rules) Of(a Action) RuleMode {
	switch a {
	case Create:
		return r.Create
	case Read:
		return r.Read
	case Update:
		return r.Update
	case Delete:
		return r.Delete
	}
	return 0
}
```

(`Of` stays here because `crud` reads it on every action at run time; it holds no error text.)

New file `module_definition.go`:

```go
package model

// ModuleDefinition is the one declaration of a domain module: its name and the entities it serves. definitionc generates
// the module's operations from it; nothing else lists the module's entities.
//
//	var Definition = model.ModuleDefinition{Name: "patient_directory", Entities: []model.Definition{PatientModel}}
type ModuleDefinition struct {
	Name     string       // the module identity, e.g. "patient_directory"
	Entities []Definition // every entity of the module, in declaration order
}
```

`gotest` passes (the package builds and vets; no new test). Commit: `feat: Rules, MFA and ModuleDefinition (data only)`.

## Stage 2 — documentation

- `README.md` (≤ 150 lines, the four ordered sections of the template in AGENTS.md): in *Why This Exists* add the Without/With of declaring
  `Rules` (without: a hand-written mount per entity that can be forgotten; with: one line per action); in *Step-by-Step Guide* add a compilable
  `ModuleDefinition` + `Rules` example; say that checking a declaration is the job of `definitioncheck`.
- `docs/ARCHITECTURE.md`: a section on `Rules`/`ModuleDefinition` in the four-section format; state that `definitioncheck` judges them and `crud` and `definitionc` consume them.
- Do not link `docs/PLAN.md` from permanent docs.

## Delivery check

1. `gotest` (builds, vets, runs; a PR that does not compile, or whose tests could not run, is a failed delivery).
2. `git diff --stat main` lists only: `definition.go`, `rules.go`, `module_definition.go`, `README.md`, `docs/ARCHITECTURE.md`.
3. `grep -rn "interface{}\|\bany\b\|webtyp.com/fmt\|func.*Check" rules.go module_definition.go` → empty (typed, no imports, no checking code).
4. `lang.json` is unchanged (this plan adds no text).

| Stage | Files | Commit |
|---|---|---|
| 1 | `definition.go`, `rules.go`, `module_definition.go` | `feat: …` |
| 2 | `README.md`, `docs/ARCHITECTURE.md` | `docs: …` |
