---
PLAN: "feat: ResourceOf and FeatureResource — a permission comes from its table, or is declared once by the library that owns the feature"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 593671046190705875
---

> Read AGENTS.md at the repo root before any change.

# PLAN — the permission comes from the table (step 1 of 3)

## Why

`model.Resource` is a `string`, so any word compiles; seven libraries use six conventions
(`"service"`, `ResourceDevice`, `"staff_manager"`, `"users"`, `access_role`, `lan_identity`), and the
UI gates a screen by one string while the server enforces another. Decision D26 of
`veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md`: the permission of a table comes from the
table's `Definition` (as Django creates permissions per model,
https://docs.djangoproject.com/en/5.1/topics/auth/default/#default-permissions, and Kubernetes RBAC
names the API type, https://kubernetes.io/docs/reference/access-authn-authz/rbac/); what is not a
table is declared once by the library that owns it. Step 1 (this plan) adds the two constructors
**without changing the `Resource` type**, so nothing breaks. Step 2 moves every repo to them;
step 3 makes `Resource` refuse a plain string.

## Design gate (api-design)

1. **Prior art.** Django (permission per model, created on migrate), Kubernetes (resource = API
   type), and this repo's own `Action`: a closed type that stopped drifting once a typo no longer
   compiled (`rbac.go`, «Closed on purpose»).
2. **Names.** `model.ResourceOf(ServiceModel)` — «the resource of the service table».
   `model.FeatureResource("api_explorer")` — «a resource that is a feature, not a table».
3. **Ledger.** Concepts +2 constructors now, −1 way after step 3 (no string literals).
4. **Where.** `model`, next to `Resource` (`rbac.go`).
5. **Deletes.** Nothing in step 1; step 2 deletes every `Resource*` constant that repeats a table.

## Stage 1 — `rbac.go`

Next to `type Resource string`:

```go
// ResourceOf is the permission of a table: the name of its Definition. A
// Definition with no primary-key field is not a table and gets "" — an enforcer
// rejects a guarded route with an empty resource at startup (AccessGuarded).
//
//	reg.Operation(OpListServices, h).Requires(model.ResourceOf(&ServiceModel), model.Read)
func ResourceOf(def *Definition) Resource {
	if def == nil {
		return ""
	}
	for _, f := range def.Fields {
		if f.IsPK() {
			return Resource(def.Name)
		}
	}
	return ""
}

// FeatureResource is the permission of something that is not a table — the API
// explorer, the database diagram. Declare it once, as a package variable, in the
// library that owns the feature; never inline it at a call site.
//
//	var ResourceIntrospection = model.FeatureResource("api_explorer")
func FeatureResource(name string) Resource { return Resource(name) }
```

Update the doc comment of `Resource`: «Build it with ResourceOf (a table) or FeatureResource (a
feature); a plain string will stop compiling».

Also fix the contradiction in the file's header comment (`rbac.go:5-9`: «a module's IDENTITY and the
RESOURCE that protects it are the same name»): replace it with «the resource that protects a record
is its table (ResourceOf); a module protects each table it owns separately». And in `interface.go`
(the `Model` comment, `ModuleNaming // ModelName() → routing, RBAC resource`), change to
`→ routing`.

## Stage 2 — `tests/resource_test.go` (new)

| Test | Input | Expected |
|---|---|---|
| `TestResourceOf_Table` | a `Definition{Name: "service", Fields: {{Name: "id", DB: &FieldDB{PK: true}}, …}}` | `"service"` |
| `TestResourceOf_NotATable` | a `Definition` with no PK field (an args DTO) | `""` |
| `TestResourceOf_Nil` | `nil` | `""` |
| `TestFeatureResource` | `"api_explorer"` | `"api_explorer"` |

## Stage 3 — docs

`README.md` (RBAC section): the two constructors, one example each, and «step 3 will make a plain
string stop compiling». `ARCHITECTURE.md`: why (the drift, Django/Kubernetes, `Action` precedent).

## Delivery check

```
gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission: when the check passes, commit, push and open the PR. Stuck 30
  minutes: open the PR listing what fails.
- `git diff --stat main` lists only: `rbac.go`, `interface.go`, `tests/resource_test.go`,
  `README.md`, `ARCHITECTURE.md`, this plan.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `rbac.go`, `interface.go` | compiles |
| 2 | `tests/resource_test.go` | 4 tests pass |
| 3 | `README.md`, `ARCHITECTURE.md` | sections present |

## Executor notes
- `TestResourceOfIsTheModuleIdentity` test in `tests/rbac_test.go` was removed because `ResourceOf` no longer accepts `ModuleNaming`. This was a consequence of the signature change for `ResourceOf`.
