# Repository-Level Work Breakdown

This section maps work to likely repository areas. Exact paths may change, but responsibilities should remain.

The purpose is to prevent architecture tasks from becoming vague cross-repository work. Each area must have concrete ownership, deliverables, tests, and acceptance gates.

## 32.1 `proto/`

Tasks:

- define protobuf messages for new modules;
- define query services;
- define tx services;
- define genesis messages;
- define params messages;
- run code generation;
- add proto breaking-change checks if available.

Tests:

- generated code compiles;
- proto lint passes if configured;
- query/tx service registration tested.

The `proto/` tree owns public wire contracts. Query, tx, genesis, and params definitions are not internal implementation details; they are API commitments for CLI, gRPC, REST gateway, wallets, explorers, dashboards, indexers, validators, and automation.

Code generation must be reproducible. Generated Go must compile, must match the selected code generation workflow, and must not be manually edited. If proto breaking-change checks are available, they must run before public testnet and on every protocol-facing proto change.

Service registration tests must prove that query and tx services are reachable through the app module wiring, not only that generated files exist. If a module has proto definitions but no registered query/tx service where one is required, the module is not production-ready.

## Acceptance Gate

The implementation gate is `DefaultAetraRepoProtoWorkEvidence` in `app/params/repository_work_breakdown.go`.

Required behavior:

- missing protobuf message definitions fail readiness;
- missing query service definitions fail readiness;
- missing tx service definitions fail readiness;
- missing genesis or params messages fail readiness;
- missing code generation step fails readiness;
- missing breaking-change checks fail readiness where available;
- missing generated-code compile, proto lint, or service-registration tests fail readiness.
