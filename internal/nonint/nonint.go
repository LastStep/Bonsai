package nonint

// This file is intentionally minimal — the package surface is split across:
//
//   - config.go : LoadConfig + applyDefaults
//   - result.go : Result (the structured headless return value) + Counts
//   - events.go : EmitJSONL (Result → JSONL) + EmitFile / EmitSummary
//                 helpers + private fileEvent / summaryEvent JSON shapes
//   - runner.go : RunInit + RunAdd orchestrators (+ ExitOK / ExitInvalidConfig
//                 / ExitRuntime / ExitWrongCWDForInit / ExitConflict codes)
//
// Keeping nonint.go around as a stable anchor for godoc and as the obvious
// "start reading here" file when the package surface grows.
