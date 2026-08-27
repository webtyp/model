---
PLAN: "feat: describe a policy so introspection can name who holds a permission"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `PolicyDescriber`: make an `Authorizer` able to say who it grants

## Why this exists

`model.Authorizer` answers one question: *may this user do this?* It is a
closure, so nothing can ask it the inverse question: *who can do this?*

That gap is not academic. In `veltylabs/misitio`, six of eleven registered
routes require a `(Resource, Action)` pair that the app's policy grants to **no
role at all**. Every one answers `403` to every caller, including the
administrator. The routes look correctly declared, the policy looks correctly
written, the tests are green, and the API is dead. Nothing in the ecosystem can
detect that today, because no type can enumerate a policy.

This plan adds the smallest thing that closes it: an optional interface a policy
owner implements, plus one free function that reports the empty case.

## Scope

This repo is the vocabulary of the ecosystem: types, no machinery. Everything
below is **additive** — no existing signature changes, no behaviour changes.
Nothing here enforces anything; enforcement stays in the routers.

## Anti-footguns

- `tinywasm/model` has **no dependencies** and must keep none. Everything below
  uses only what is already in this package. Do not import `tinywasm/fmt`,
  `tinywasm/json`, `strings`, or anything else.
- Tests in this repo are **in-package** at the module root (`package model`,
  e.g. `rbac_test.go`). Do not create a `tests/` directory.
- `Action` is a bit mask and `Access`'s zero value is `AccessGuarded`. Do not
  "simplify" either.

---

## Stage 1 — `RoleGrant` and `PolicyDescriber`

File: **`rbac.go`** (append; do not restructure the file).

```go
// RoleGrant binds a role code to one Grant. It is the shape a policy is
// DECLARED in, and the shape introspection reads it back out in.
//
// It is deliberately a pair and not a map: a role holds several grants, a
// grant is held by several roles, and a slice of pairs states that without
// picking a winner. It also keeps this package free of map iteration order,
// which the WASM targets in this ecosystem avoid.
type RoleGrant struct {
	Role  RoleCode
	Grant Grant
}

// PolicyDescriber is implemented by whoever OWNS an Authorizer when it can
// enumerate what it grants.
//
// It is optional on purpose. An Authorizer that cannot describe itself keeps
// working exactly as before; introspection then reports the permission a route
// requires without being able to say who holds it. What it must never do is
// report "nobody" when it simply did not know — see RolesFor.
type PolicyDescriber interface {
	// Grants returns every (role, grant) pair the policy declares. The order
	// is the policy's own declaration order, so a reader sees the policy as
	// its author wrote it.
	Grants() []RoleGrant
}
```

## Stage 2 — `RolesFor`

Same file, right after `AnyGrant` — that is where it belongs: `AnyGrant(grants, r, a) bool`
already answers the forward question over a grant slice, and `RolesFor` is its inverse over a
described policy. **Do not reimplement `Grant.Matches` or `AnyGrant`**; call what is there.

```go
// RolesFor returns the role codes p grants (r, a) to, in the policy's own
// declaration order and without repeats.
//
// An EMPTY result is the finding this function exists for: a permission no
// role holds. A route requiring it is a permanent 403 that looks correctly
// declared — the exact failure that motivated this API. Callers must render
// that case loudly rather than as an empty column.
//
// A nil p returns nil. That is "unknown", not "nobody": a caller that cannot
// tell the two apart must check p != nil itself before reporting.
func RolesFor(p PolicyDescriber, r Resource, a Action) []RoleCode {
	if p == nil {
		return nil
	}
	var out []RoleCode
	for _, rg := range p.Grants() {
		if !rg.Grant.Matches(r, a) {
			continue
		}
		seen := false
		for _, have := range out {
			if have == rg.Role {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, rg.Role)
		}
	}
	return out
}
```

`Grant.Matches` already handles the wildcard resource and the empty grant; do not
reimplement either.

Note the neighbours this API must stay consistent with, all already in `rbac.go`:
`Resource`, `RoleCode`, `Action`, `Grant`, `Matches`, `AnyGrant`, `Authorizer`, `Allowed`,
`Wildcard`, and `ResourceOf(ModuleNaming) Resource` — the function that ties a module's
identity to its RBAC resource. `RolesFor` completes that set; it does not start a new one.

## Stage 3 — Tests

File: **`rbac_test.go`** (append to the existing file, `package model`).

Required cases, each a `t.Run` subtest:

| Name | Setup | Assert |
|---|---|---|
| `RolesFor: exact match` | `admin → catalog:ru` | `RolesFor(p, "catalog", Read)` = `["admin"]` |
| `RolesFor: action not granted` | `admin → catalog:ru` | `RolesFor(p, "catalog", Delete)` = empty |
| `RolesFor: resource not granted` | `admin → catalog:ru` | `RolesFor(p, "invoice", Read)` = empty |
| `RolesFor: two roles, one permission` | `admin → catalog:r`, `editor → catalog:r` | result is `["admin","editor"]`, **in that order** |
| `RolesFor: same role twice` | `admin → catalog:r`, `admin → catalog:ru` | result is `["admin"]` — no repeat |
| `RolesFor: wildcard resource` | `root → *:crud` | `RolesFor(p, "anything", Create)` = `["root"]` |
| `RolesFor: nil describer` | `p = nil` | result is `nil` |
| `RolesFor: nobody holds it` | `admin → catalog:r` | `RolesFor(p, "site_asset", Create)` is **empty** — the case the API exists for |

Use a small in-test type as the fixture; do not export a test double from the
package:

```go
type testPolicy []RoleGrant

func (p testPolicy) Grants() []RoleGrant { return p }
```

## Stage 4 — Documentation

- **`docs/ARCHITECTURE.md`** — add a short subsection under the RBAC vocabulary
  explaining the asymmetry `Authorizer` (forward: may X do Y?) vs
  `PolicyDescriber` (inverse: who may do Y?), and why the inverse is optional.
  State the motivating failure in one sentence: a permission no role holds turns
  a correctly-declared route into a permanent 403, and only the inverse question
  finds it.
- **`docs/API_PERMITTED.md`** — add `RoleGrant`, `PolicyDescriber` and
  `RolesFor` to the reference with their signatures.
- Do **not** link `docs/PLAN.md` from any permanent document.

## Acceptance criteria

- [ ] `go build ./...` and `go vet ./...` clean.
- [ ] `gotest ./...` green.
- [ ] `grep -n "^require\|^	" go.mod` → the module still has **zero**
      dependencies.
- [ ] `grep -rn "map\[" rbac.go` → empty.
- [ ] `RolesFor` with a policy where nothing matches returns an empty slice, and
      a test asserts it by name.
- [ ] `docs/ARCHITECTURE.md` and `docs/API_PERMITTED.md` describe the new API.

## Out of scope

Enforcement, wire encoding of `RoleGrant` (the consumer that serializes it
declares its own shape — see `router.RouteInfo.EncodeFields` for the precedent),
and any change to `Authorizer`, `Grant`, `Action` or `Access`.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | `RoleGrant` + `PolicyDescriber` | `rbac.go` |
| 2 | `RolesFor` | `rbac.go` |
| 3 | Tests | `rbac_test.go` |
| 4 | Documentation | `docs/ARCHITECTURE.md`, `docs/API_PERMITTED.md` |
