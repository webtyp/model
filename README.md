# webtyp/model
<img src="docs/img/badges.svg">

Core schema, validation, and codec types for the webtyp ecosystem.

## Why This Exists

`model` provides the single source of truth for schemas, validation, and serialization contracts. Without it, each domain module repeats manual mounts and wire contracts that can be forgotten or drift. With `model`, entities declare their schema, allowed actions, business rule modes, and MFA requirements as pure data that generators (`definitionc`) and harnesses (`crud`) consume.

| Without | With |
|---|---|
| A hand-written mount per entity that can be forgotten | One line per action in `Rules` (`AddMyBusinessRule` / `HaveNoBusinessRule`) |
| Fragile string tags (`input:"required"`) with runtime failures | Typed `model.Definition` with compile-time symbol checking |
| Reflection and `map` in serialization | Zero-alloc typed codecs ([`Encodable`](docs/API_CODEC.md)/[`Decodable`](docs/API_CODEC.md)) compiling to WASM |
| Ad-hoc validation logic in every handler | Fail-closed validation contracts ([`Permitted`](docs/API_PERMITTED.md)) on each field |

Checking a declaration's validity (e.g. that every allowed action declares its rule mode and MFA is a subset of `Allow`) is the job of [`webtyp.com/definitioncheck`](../definitioncheck/).

## Before You Start

- **Isomorphic package**: compiles to WebAssembly with TinyGo. No `reflect`, no `map`, no stdlib `fmt`/`strings` (use `webtyp.com/fmt`).
- **Data contracts only**: `model` defines data structures (`Rules`, `ModuleDefinition`, `Definition`). Validation tooling lives in `definitioncheck`.
- **AST introspection**: generators like `definitionc` parse `Definition` literals statically from source files.

## Step-by-Step Guide

### Step 1: Declare an Entity with Rules and Module Definition

**Goal:** Define a typed entity with explicit CRUD permissions, rule modes, MFA protection, and register it in a module definition.

```go
package main

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

var UserModel = model.Definition{
	Name:  "user",
	Allow: model.Create | model.Read | model.Update,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.HaveNoBusinessRule,
		Update: model.AddMyBusinessRule,
	},
	MFA: model.Update,
	Fields: model.Fields{
		{Name: "id", Type: model.Int(), DB: &model.FieldDB{PK: true, AutoInc: true}},
		{Name: "name", Label: "Full Name", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 2}},
		{Name: "email", Label: "Email", Type: model.Text(), NotNull: true},
	},
}

var UserModule = model.ModuleDefinition{
	Name:     "user_directory",
	Entities: []model.Definition{UserModel},
}

func main() {
	mode := UserModel.Rules.Of(model.Create)
	fmt.Println(UserModule.Name, len(UserModule.Entities), mode == model.AddMyBusinessRule)
}
```

**Expected result:** Output `user_directory 1 true`. Checking that this declaration is valid is performed statically by `definitioncheck`.

### A revision counter: two edits of the same row must not both win

Mark one integer field as the row's revision and `crud` rejects an update that was made from an older read (409), while `orm` applies it only if the column still holds the value the client read:

```go
{Name: "revision", Type: model.Int(), DB: &model.FieldDB{Revision: true}}
```

`Field.IsRevision()` tells whether a field is that counter. `definitioncheck` refuses a revision that is not an integer, a second one, or the primary key.

## Documentation Map

| Order | Document | What question it answers & when to read it |
|---|---|---|
| 1 | [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Architectural decisions, no-reflection constraints, and rules |
| 2 | [`docs/API_FIELD.md`](docs/API_FIELD.md) | Field kinds, schema metadata, and reflection-less struct introspection |
| 3 | [`docs/API_CODEC.md`](docs/API_CODEC.md) | Zero-allocation typed serialization contracts (`Encodable`/`Decodable`) |
| 4 | [`docs/API_PERMITTED.md`](docs/API_PERMITTED.md) | Validation contracts, character whitelists, and length constraints |
| 5 | [`docs/WHY_64BIT_ONLY.md`](docs/WHY_64BIT_ONLY.md) | Rationale for `int64` and `float64` as standard numeric types |
| 6 | [`docs/API_EXTERNAL_CALL.md`](docs/API_EXTERNAL_CALL.md) | External call declarations for entities |
