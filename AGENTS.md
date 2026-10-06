# AGENTS.md — webtyp/model

Working notes for AI agents operating in this repository. End-user docs: [README.md](README.md).

## What this repo is

The single schema of the webtyp ecosystem: `model.Definition`, `model.Field`, `Kind`, `Permitted`,
the typed codec interfaces. One `Field` serves DDL (`orm`, `postgres`), validation, wire codecs and UI
(`form`). `ormc` reads `Definition` literals statically (AST) to generate code.

## This package compiles to WASM

- Do NOT import `strings`, `strconv`, `errors` or stdlib `fmt`; use `webtyp.com/fmt`.
- No `map`, no `reflect` in code that reaches WASM.
- A new `Field` key must stay a plain literal-friendly value (string, bool, pointer to a literal
  struct): `ormc` parses `Definition` literals from source and ignores keys it does not know.

## The build that defines "done"

```bash
go install webtyp.com/devflow/cmd/gotest@latest   # once
gotest
```

## Rules

- Tests live in `tests/` (`package model_test`, public API only). A root-level test is allowed only
  with a top-of-file `// Root-level test (justified): …` comment. **Never export a symbol so a test
  can reach it.**
- Every repeated string is a named constant.
- Display text in the schema (e.g. `Field.Label`) is written in **English**; it is a translation key
  for `webtyp.com/lang`, never a human language chosen by a library.
