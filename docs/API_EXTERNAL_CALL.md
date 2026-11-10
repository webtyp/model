# ExternalCallDefinition API

## Definitions

`ExternalCallDefinition` declares one call to another system that a rule of the entity may ask for. Limits are written out, never defaulted; a limit that does not belong to the Mode is an error.

nothing runs these calls yet: `crud` and `externalcalls` do, in later releases

### ExternalCallMode

`ExternalCallMode` says what a rule waits for when it asks for an external call.

- **WaitForReply**: the request waits for the provider's answer, up to `MaxWaitSeconds`. If the time runs out the call fails: it is never retried in the background (D60) and the record is kept.
- **RetryInBackground**: the request answers at once and the call is retried in the background, at most `MaxAttempts` times.
- **WaitForConfirmation**: the call stays queued until the provider confirms or rejects it; there is no time limit here and it is never retried in the background (D60).

### Example Usage

```go
ExternalCalls: []model.ExternalCallDefinition{
	{Name: "charge", Mode: model.WaitForReply, MaxWaitSeconds: 10},
	{Name: "receipt", Mode: model.RetryInBackground, MaxAttempts: 8},
	{Name: "settle", Mode: model.WaitForConfirmation},
}
```

## Validation Errors

| Error Constant | Resolution |
|----------------|------------|
| `ErrExternalCallNameRequired` | Provide a valid `Name` string for the call. |
| `ErrExternalCallModeUndeclared` | Use `WaitForReply`, `RetryInBackground` or `WaitForConfirmation`. |
| `ErrExternalCallMaxWaitRequired` | Provide a `MaxWaitSeconds` value greater than zero for `WaitForReply`. |
| `ErrExternalCallMaxAttemptsRequired` | Provide a `MaxAttempts` value greater than zero for `RetryInBackground`. |
| `ErrExternalCallLimitNotAllowed` | Provide only `MaxWaitSeconds` for `WaitForReply`, only `MaxAttempts` for `RetryInBackground`, and neither for `WaitForConfirmation`. |
| `ErrExternalCallDuplicate` | Ensure names are unique within the entity. |

See [external_call.go](../external_call.go) for details.
