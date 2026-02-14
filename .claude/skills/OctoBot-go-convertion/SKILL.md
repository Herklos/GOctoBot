---
name: octobot-go-convertion
description: Convert the OctoBot Python codebase into a Go codebase that mirrors the original structure, classes, functions, variables, and tests. Use for architectural planning, file-by-file translation rules, language choices, bindings, and strict parity guidance for the OctoBot to GOctoBot migration.
---

# OctoBot Go Conversion

## Goal

Create a Go codebase that is a near-mirror of the OctoBot Python codebase, preserving:
- File and folder structure
- Class and function names
- Public APIs and behaviors
- Variables and configuration keys
- Test coverage and semantics

## Non-Goals

- Do not redesign architecture or refactor for style.
- Do not change external behavior or API contracts.
- Do not introduce new features or remove existing ones.

## High-Level Architecture Plan

- Mirror the Python package tree under a Go module root.
- Map each Python module to a Go package with the same relative path.
- Map each Python class to a Go `struct` with methods.
- Map each Python function to a Go function in the corresponding package.
- Preserve initialization and side-effects order from module import time by explicit `init()` or bootstrap functions.
- Preserve configuration layout and defaults with a single source of truth.
- Defer conversion of complex Services tentacles for now and focus on core libraries and framework layers first.

## Language Choices and Bindings

- Use Go 1.22+ with a single module for the entire repo.
- Use standard library where possible.
- Use minimal external dependencies and document all of them in `go.mod`.
- For HTTP, use `net/http` and `context` for timeouts and cancellation.
- For JSON, use `encoding/json` with strict struct tags mirroring Python dict keys.
- For logging, choose one logging package and apply it uniformly.
- For concurrency, use goroutines and channels only where Python uses concurrency or async.

## Naming and Parity Rules

- Preserve file names. Convert `.py` to `.go` with the same stem.
- Preserve package names by folder path. Use lowercase package names.
- Preserve class names and method names. Use PascalCase for exported symbols and camelCase for private symbols.
- Preserve function signatures as closely as possible.
- Preserve variable names in struct fields and JSON tags to match Python keys.
- Preserve constant names and values. Use `const` in Go.
- Preserve enum-like classes by using typed constants in Go.

## Structural Mapping Rules

- Python package `OctoBot/packages/tentacles/...` maps to `GOctoBot/packages/tentacles/...`.
- Ignore `OctoBot/tentacles` if it mirrors `OctoBot/packages/tentacles`.
- Python module file `foo.py` maps to `foo.go` in the same relative folder.
- Python classes map to Go structs with methods in the same file unless file size forces split.
- Python module-level state maps to package-level variables.
- Python `__init__.py` behavior maps to `init()` or explicit `Init()` in the package.

## Type System Mapping

- Use explicit Go types for all values. Avoid `interface{}` unless the Python code is truly dynamic.
- Map Python `dict` to Go `map[string]T` where key set is dynamic.
- Map Python `list` to Go slices.
- Map Python `None` to Go zero values or pointers.
- Map Python classes with dynamic attributes to Go `map[string]any` only when unavoidable.

## Error Handling Rules

- Convert Python exceptions to Go `error` returns.
- Preserve error messages and error types when possible.
- Use wrapped errors (`fmt.Errorf("%w", err)`) to preserve context.
- Propagate errors up the stack unless Python explicitly swallows them.

## Concurrency and Timing

- Translate Python async and threading into Go goroutines and `context`.
- Preserve polling intervals, timeouts, and backoff behavior.
- Preserve scheduling order where tests depend on it.
- OctoBot is primarily async; mirror this behavior in Go.
- It is critical to use async everywhere the Python codebase is async to preserve concurrency behavior.
- Use goroutines as the primary equivalent of `asyncio.Task`.
- Implement a lightweight task-runner or scheduler to preserve task lifecycle semantics.
- Prefer async-style APIs across the codebase; avoid introducing synchronous blocking paths.

## Configuration and Environment

- Preserve all config file formats and keys.
- Keep default values identical.
- Provide a Go config loader that mirrors Python behavior.
- Preserve environment variable names and precedence rules.

## Logging and Observability

- Preserve log messages verbatim when possible.
- Preserve log levels and when logs are emitted.
- Add structured fields only if they do not change output semantics.

## IO and External Services

- Preserve request URLs, query params, headers, and payloads.
- Preserve retry logic and rate-limiting behavior.
- Preserve cache behavior, file locations, and serialization formats.

## Tests and Parity Validation

- Mirror Python tests into Go tests with the same file names and test names.
- Keep test inputs and expected outputs identical.
- Create parity tests that compare Go outputs to Python outputs for critical paths.
- Add golden-file tests for serialized outputs.
- Ensure deterministic ordering in maps and JSON to avoid flaky tests.
- Do not drop or skip tests; all Python tests must be mirrored in Go even if they appear redundant.

## Translation Workflow

1. Inventory the Python package tree and create a Go package map.
2. For each Python module, create a Go file with stub types and functions.
3. Port constants and configuration first to establish stable types.
4. Port data models and serialization mappings next.
5. Port service clients and IO next.
6. Port logic and orchestration last.
7. Port tests in lockstep with module conversion.
8. Record any seemingly unused functions in the unused-function log while still converting them.

## Exploration Checklist (If Not Already Covered)

- Identify dynamic imports and plugin loading paths; replicate with explicit registries.
- Track any runtime class discovery (`__subclasses__`, metaclasses, decorators) and map to registries.
- Detect any monkeypatching or global state mutation and mirror ordering and side effects.
- Note any reliance on module import order; make it explicit in Go `init()` or bootstrap.
- Find any implicit singletons or cached instances and mirror lifecycle and reset behavior for tests.
- Record any serialization quirks or custom encoders and port them verbatim.
- Check for background threads or timers and mirror scheduling and shutdown paths.
- Identify any test-only hooks, fixtures, or test helpers and port them to Go test utilities.

## Module-Specific Guidance

- Services feeds (e.g., `coingecko_service_feed`, `alternative_me_service_feed`) should preserve:
- Endpoint URLs and request parameters.
- Rate limiting and backoff behavior.
- Response parsing and error handling.

## Go Project Structure Expectations

- Keep a monorepo layout that mirrors the Python tree.
- Prefer isolated libraries with their own `go.mod` where Python has distinct libs.
- Use `go.work` at `GOctoBot/` to tie multiple modules together.
- Keep cross-library dependencies explicit and minimal.
- Use `internal/` only if Python had internal-only modules.
- Use `cmd/` if Python has entrypoints or CLIs.
- Ensure package import paths match the folder hierarchy.

## Scope For Now

- Skip Services tentacles in early passes.
- Stub Services packages with minimal types and TODOs to keep builds green.
- Convert core framework and shared libraries first.

## OctoBot-Channels Detailed Conversion Checklist

1. Inventory and Map
1.1. List every module and public symbol (classes, functions, constants).
1.2. Produce a module-to-package map with file-level parity targets.
1.3. Document all public APIs and their expected behavior.
1.4. Identify any dynamically registered classes or handlers.

2. Semantics Capture
2.1. Define message envelope structure, metadata fields, and defaults.
2.2. Define routing rules (topic names, filters, wildcard rules).
2.3. Define delivery guarantees (at-most-once, at-least-once, exactly-once).
2.4. Capture ordering constraints (per-channel, per-subscriber, global).
2.5. Capture backpressure behavior (drop, block, queue limits).
2.6. Capture error handling and retry behavior.

3. Concurrency Model
3.1. Enumerate all concurrency boundaries in Python (threads, async, locks).
3.2. Identify shared mutable state and required locking.
3.3. Decide on Go implementation:
- If Go channels can match semantics, use them with explicit buffering and shutdown.
- Otherwise implement a custom dispatcher with queues and locks.
3.4. Define shutdown and drain behavior explicitly.

4. API and Type Parity
4.1. Create Go interfaces for core abstractions (Channel, Subscriber, Publisher).
4.2. Create Go structs for concrete types with identical names and fields.
4.3. Mirror all constants and defaults.
4.4. Preserve any extension or plugin APIs.

5. Lifecycle and Registry
5.1. Implement explicit registration of channel types and handlers.
5.2. Preserve initialization order and side effects.
5.3. Provide deterministic enumeration of registered types.

6. Tests and Parity
6.1. Port all tests with identical names and inputs.
6.2. Add parity tests for ordering, backpressure, and error semantics.
6.3. Add stress tests for throughput and race safety if present in Python.
6.4. Add golden tests for serialized message formats.

7. Validation
7.1. Run Go tests and compare outcomes to Python tests.
7.2. Document any behavior differences and resolve them.
7.3. Do not finalize until parity is proven for core channel flows.

## Introspection and Class Registry Strategy

- Acknowledge that Go cannot enumerate all types at runtime without explicit registration.
- Implement explicit registries that mirror Python class discovery patterns.
- Register each concrete type in `init()` with a stable string key matching the Python class name.
- Maintain parent-to-children indexes to emulate `__subclasses__` and parent introspection.
- Use interfaces to represent Python base classes and embedding for shared behavior.
- Use `reflect.Type` only for metadata and debugging, not primary logic.
- Consider `go:generate` for codegen to keep registries complete and deterministic.
- Ensure registry iteration order is deterministic for tests.

## Registry Pattern Template (Go Skeleton)

```go
package registry

import (
    "fmt"
    "sort"
    "sync"
)

type Base interface {
    BaseName() string
}

type Factory func() Base

var (
    mu             sync.RWMutex
    byName         = map[string]Factory{}
    childrenByBase = map[string][]string{}
)

func Register(baseName, typeName string, factory Factory) {
    mu.Lock()
    defer mu.Unlock()
    if _, exists := byName[typeName]; exists {
        panic(fmt.Sprintf("duplicate type registration: %s", typeName))
    }
    byName[typeName] = factory
    childrenByBase[baseName] = append(childrenByBase[baseName], typeName)
    sort.Strings(childrenByBase[baseName])
}

func Create(typeName string) (Base, error) {
    mu.RLock()
    factory, ok := byName[typeName]
    mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("unknown type: %s", typeName)
    }
    return factory(), nil
}

func ListChildren(baseName string) []string {
    mu.RLock()
    defer mu.RUnlock()
    children := childrenByBase[baseName]
    out := make([]string, len(children))
    copy(out, children)
    return out
}
```

## Documentation Outputs

- Maintain a conversion map document listing Python module to Go package mapping.
- Maintain a parity checklist per module.
- Keep a running list of behavior differences and address them.
- Maintain an unused-function log listing functions ported but apparently unused.

## Coding Rules

- Prefer clarity over cleverness.
- Keep functions short and aligned with Python logical blocks.
- Do not change control flow unless required by Go semantics.
- Avoid global state unless it exists in Python.
- Preserve ordering and defaults to keep tests green.

## Done Criteria

- All Python modules have Go equivalents in the same relative paths.
- All Python classes and functions have Go equivalents.
- All tests are ported and passing.
- Behavior is validated by parity tests for key modules.

## When Blocked

- Document the exact Python behavior that cannot be mirrored.
- Add a temporary shim with a TODO and a test that locks the intended behavior.
- Escalate for design decisions only when strict parity is impossible.
