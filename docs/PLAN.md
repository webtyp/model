---
PLAN: "feat: model.ExternalCallDefinition and Definition.ExternalCalls — an entity declares the external calls its rules may ask for, with mode and limits written out"
TAG: v0.2.19
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> A PR that does not compile, or whose tests could not run, is a failed delivery. Stuck for 30 minutes: open the PR anyway and list what fails.

# Plan: external call declarations in `model`

First slice of decision D36 (an entity declares its external calls inside its own `model.Definition`; `crud` and `externalcalls` run them later). This plan adds **only the declaration types and their validation**. It does not touch `crud`, `definitionc` or `appkit`: nothing reads the new field yet, so no consumer changes.

## Ecosystem search

Searched `grep -rn "ExternalCall\|WaitForReply\|RetryInBackground\|WaitForConfirmation" --include=*.go --include=*.md ~/Dev/Project/webtyp ~/Dev/Project/veltylabs`.

| Where | What exists | Decision |
|---|---|---|
| `webtyp/docs/EXTERNAL_CALLS_DRAFT.md`, `ARCHITECTURE_MASTER_PLAN.md` D36 | the design and the names, approved by the owner (names table of 2026-11-10) | **reuse** the names exactly |
| `model.RuleMode` (`rules.go`) | closed enum, zero = «undeclared», refused by definitioncheck | **reuse the pattern** for `ExternalCallMode` |
| `webtyp/outbox` | browser queue of mutations, writes outside crud's transaction | **not related**: another concern (D36) |

No Go code with these names exists: **new**.

## Design gate

**1. Prior art** (links **unverified**, not read in this session): Stripe API idempotency keys and async webhooks (https://docs.stripe.com/api/idempotent_requests); Temporal activity options `StartToCloseTimeout` and `RetryPolicy.MaximumAttempts` (https://docs.temporal.io/activities); Sidekiq retries with a maximum count (https://github.com/sidekiq/sidekiq/wiki/Error-Handling). They share: a call has a timeout or a maximum number of attempts, written by the caller, never hidden. This plan follows that and **departs** by closing the set of modes to three and rejecting a limit that does not belong to the mode (illegal states unrepresentable).

**2. Novice-name test.** `WaitForReply` («wait for the provider's answer, up to MaxWaitSeconds»), `RetryInBackground` («answer now, keep trying»), `WaitForConfirmation` («stay waiting until the provider confirms»), `MaxWaitSeconds`, `MaxAttempts`: each reads as a sentence. The unit is in the name (`Seconds`) so a literal `10` cannot be read as milliseconds.

**3. Complexity ledger.**
```
Concepts the developer must learn   +3 (ExternalCallDefinition, ExternalCallMode, ExternalCalls)
Files they must touch to do X       +0 (declared in the entity's own definition.go)
Lines at the call site              +1 per external call
Ways to do the same thing           +0 (no external-call declaration exists today; module hooks `Committed` are removed by later plans)
```

**4. Where it belongs.** `model` is the single schema: an entity's declarations live in `model.Definition` (same place as `Allow`, `Rules`, `MFA`). `model` imports only `webtyp.com/fmt`; this plan adds no import (the limit is `int64` seconds, not `time.Nano`, so `definitionc` can read a plain literal).

**5. What it deletes.** Nothing: new capability. The `Committed` hooks and hand-written outbox code of the five modules with external calls are deleted by their own plans, after `externalcalls` is published.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`; `go mod tidy` after any `go get`. This plan adds no dependency.

## APIs used

| symbol | exact signature | tag | file |
|---|---|---|---|
| `model.Invalid` | `type Invalid string` with `Error() string` and `Kind() ErrorKind` | this repo `v0.2.18` | `error.go` |
| `model.RuleMode` | `type RuleMode uint8`, `AddMyBusinessRule RuleMode = iota + 1` | this repo `v0.2.18` | `rules.go` |
| `model.Definition` | struct with `Name string`, `Fields Fields`, `Wire bool`, `Allow Action`, `Rules Rules`, `MFA Action` | this repo `v0.2.18` | `definition.go:15` |

If a signature in the repo differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — tests first (red) and `docs/TESTS.md`

File: `tests/external_call_test.go`, package `model_test`, dot-import `. "webtyp.com/model"` (same as `tests/assurance_test.go`). Each test is setup → call → expected:

| Test | setup → call → expected |
|---|---|
| `TestExternalCallMode_ZeroValueIsUndeclared` | `var m ExternalCallMode` → `m != 0` is false: the zero value is none of the three modes |
| `TestExternalCall_WaitForReplyNeedsMaxWait` | `{Name:"charge", Mode: WaitForReply, MaxWaitSeconds: 10}` → `Validate()` nil; with `MaxWaitSeconds: 0` → `ErrExternalCallMaxWaitRequired` |
| `TestExternalCall_RetryInBackgroundNeedsMaxAttempts` | `{Name:"receipt", Mode: RetryInBackground, MaxAttempts: 8}` → nil; with `MaxAttempts: 0` → `ErrExternalCallMaxAttemptsRequired` |
| `TestExternalCall_WaitForConfirmationHasNoLimits` | `{Name:"settle", Mode: WaitForConfirmation}` → nil; with `MaxWaitSeconds: 5` → `ErrExternalCallLimitNotAllowed`; with `MaxAttempts: 3` → `ErrExternalCallLimitNotAllowed` |
| `TestExternalCall_LimitOfOtherModeRejected` | `{Mode: WaitForReply, MaxWaitSeconds: 10, MaxAttempts: 3}` → `ErrExternalCallLimitNotAllowed`; `{Mode: RetryInBackground, MaxAttempts: 8, MaxWaitSeconds: 10}` → `ErrExternalCallLimitNotAllowed` |
| `TestExternalCall_NameRequired` | `{Mode: WaitForConfirmation}` (no name) → `ErrExternalCallNameRequired` |
| `TestExternalCall_ModeUndeclared` | `{Name:"charge"}` (mode 0) → `ErrExternalCallModeUndeclared`; `{Name:"charge", Mode: 9}` → `ErrExternalCallModeUndeclared` |
| `TestDefinition_ValidateExternalCalls` | a `Definition{Name:"payment", ExternalCalls: [charge WaitForReply 10s, receipt RetryInBackground 8]}` → nil; a `Definition` with no `ExternalCalls` → nil |
| `TestDefinition_DuplicateExternalCallName` | two entries named `"charge"` → `ErrExternalCallDuplicate` |
| `TestDefinition_ExternalCallErrorsAreInvalid` | every `ErrExternalCall…` constant answers `Kind() == ErrorInvalid` |

Test bodies compare errors with `==` against the constant (tests run on the stdlib toolchain). `docs/TESTS.md`: add one row per test: `Test | File | Rule it proves` (copy the table shape already in that file). Commit. The package does not compile with these tests until Stage 2: expected.

## Stage 2 — `external_call.go`

New file `external_call.go` (package `model`; imports nothing, so it stays WASM-safe). Write exactly:

```go
package model

// ExternalCallMode says what a rule waits for when it asks for an external call. The zero value
// means "undeclared" and Validate refuses it.
type ExternalCallMode uint8

const (
	// WaitForReply: the request waits for the provider's answer, up to MaxWaitSeconds. If the time
	// runs out the call stays queued and the background runner finishes it.
	WaitForReply ExternalCallMode = iota + 1
	// RetryInBackground: the request answers at once and the call is retried in the background, at most MaxAttempts times.
	RetryInBackground
	// WaitForConfirmation: the call stays queued until the provider confirms it; there is no time limit here.
	WaitForConfirmation
)

// ExternalCallDefinition declares one call to another system that a rule of the entity may ask for.
// Limits are written out, never defaulted; a limit that does not belong to the Mode is an error.
//
//	ExternalCalls: []model.ExternalCallDefinition{
//		{Name: "charge", Mode: model.WaitForReply, MaxWaitSeconds: 10},
//		{Name: "receipt", Mode: model.RetryInBackground, MaxAttempts: 8},
//		{Name: "settle", Mode: model.WaitForConfirmation},
//	}
type ExternalCallDefinition struct {
	Name           string // the call's own name, written once here: "charge"
	Mode           ExternalCallMode
	MaxWaitSeconds int64 // WaitForReply only, greater than zero
	MaxAttempts    int64 // RetryInBackground only, greater than zero
}

const (
	ErrExternalCallNameRequired        Invalid = "external call name required"
	ErrExternalCallModeUndeclared      Invalid = "external call mode undeclared: use WaitForReply, RetryInBackground or WaitForConfirmation"
	ErrExternalCallMaxWaitRequired     Invalid = "external call MaxWaitSeconds required: WaitForReply needs a maximum wait greater than zero"
	ErrExternalCallMaxAttemptsRequired Invalid = "external call MaxAttempts required: RetryInBackground needs a number of attempts greater than zero"
	ErrExternalCallLimitNotAllowed     Invalid = "external call limit not allowed: WaitForReply takes only MaxWaitSeconds, RetryInBackground only MaxAttempts, WaitForConfirmation none"
	ErrExternalCallDuplicate           Invalid = "external call name repeated in the same entity"
)

// Validate reports the first thing wrong with the declaration, in this order: name, mode, then the limits of the mode.
//
//	err := model.ExternalCallDefinition{Name: "charge", Mode: model.WaitForReply}.Validate() // ErrExternalCallMaxWaitRequired
func (c ExternalCallDefinition) Validate() error {
	if c.Name == "" {
		return ErrExternalCallNameRequired
	}
	switch c.Mode {
	case WaitForReply:
		if c.MaxAttempts != 0 {
			return ErrExternalCallLimitNotAllowed
		}
		if c.MaxWaitSeconds <= 0 {
			return ErrExternalCallMaxWaitRequired
		}
	case RetryInBackground:
		if c.MaxWaitSeconds != 0 {
			return ErrExternalCallLimitNotAllowed
		}
		if c.MaxAttempts <= 0 {
			return ErrExternalCallMaxAttemptsRequired
		}
	case WaitForConfirmation:
		if c.MaxWaitSeconds != 0 || c.MaxAttempts != 0 {
			return ErrExternalCallLimitNotAllowed
		}
	default:
		return ErrExternalCallModeUndeclared
	}
	return nil
}

// ValidateExternalCalls validates every declared call and refuses two with the same name.
func (d Definition) ValidateExternalCalls() error {
	for i := range d.ExternalCalls {
		if err := d.ExternalCalls[i].Validate(); err != nil {
			return err
		}
		for j := 0; j < i; j++ {
			if d.ExternalCalls[j].Name == d.ExternalCalls[i].Name {
				return ErrExternalCallDuplicate
			}
		}
	}
	return nil
}
```

No `map` (WASM): the duplicate check is the quadratic loop above on purpose (a handful of calls per entity). Do not change it.

In `definition.go`, add to `type Definition struct`, after the `MFA` field and following the style of the neighbouring fields (a comment with an example):

```go
	// ExternalCalls are the calls to other systems that a rule of this entity may ask for, each with its mode and limits written out.
	// Names are unique within the entity; ValidateExternalCalls checks them.
	//
	//	ExternalCalls: []model.ExternalCallDefinition{{Name: "charge", Mode: model.WaitForReply, MaxWaitSeconds: 10}}
	ExternalCalls []ExternalCallDefinition
```

`go.mod` is unchanged. Run `gotest`: every test of Stage 1 passes and nothing else changes. Commit.

## Stage 3 — documentation

- `docs/API_EXTERNAL_CALL.md` (new, same shape as `docs/API_FIELD.md`: read it first and copy its headings): what the three modes mean in plain words, one complete `Definition` literal with the three calls, the table of the six errors with what to do for each, and a link to `external_call.go`. State «nothing runs these calls yet: `crud` and `externalcalls` do, in later releases».
- `docs/ARCHITECTURE.md`: in the section that lists the pieces of the package, add one row for `external_call.go`.
- `README.md`: add one row for `docs/API_EXTERNAL_CALL.md` to its Documentation Map table.

Permanent docs never link `docs/PLAN.md`. English, no client names. Commit.

## Delivery check

1. `gotest` — builds, vets and runs every test; all pass.
2. `grep -n "import" external_call.go` → empty (no imports); `grep -n "map\[" external_call.go` → empty.
3. `grep -rn "PLAN.md" README.md docs/API_EXTERNAL_CALL.md docs/ARCHITECTURE.md` → empty.
4. `git diff --stat main` lists exactly: `external_call.go`, `definition.go`, `tests/external_call_test.go`, `docs/TESTS.md`, `docs/API_EXTERNAL_CALL.md`, `docs/ARCHITECTURE.md`, `README.md`, and badge files `gotest` rewrites.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | red tests, `TESTS.md` | `tests/external_call_test.go`, `docs/TESTS.md` |
| 2 | the declaration types and validation | `external_call.go`, `definition.go` |
| 3 | documentation | `docs/API_EXTERNAL_CALL.md`, `docs/ARCHITECTURE.md`, `README.md` |
