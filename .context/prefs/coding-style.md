# Coding Style Guide

> 此文件定义团队编码规范，所有 LLM 工具在修改代码时必须遵守。
> 提交到 Git，团队共享。

## General
- Prefer small, reviewable changes; avoid unrelated refactors.
- Keep functions short (<50 lines); avoid deep nesting (≤3 levels).
- Name things explicitly; no single-letter variables except loop counters.
- Handle errors explicitly; never swallow errors silently.

## Language-Specific

### Go (Backend)
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting, `goimports` for import ordering
- Error handling: return errors explicitly, wrap with context using `fmt.Errorf`
- Avoid global state; prefer dependency injection
- Use interfaces for testability (e.g., `MessageStore`, `DeliveryReportSender`)
- Context propagation: pass `context.Context` as first parameter to I/O operations
- Naming: mixedCaps for unexported, MixedCaps for exported

### TypeScript/Vue (Frontend)
- Use TypeScript strict mode (`strict: true` in tsconfig)
- Prefer `interface` over `type` for object shapes
- Use Composition API with `<script setup lang="ts">`
- Component naming: PascalCase for components, camelCase for composables
- Use Pinia stores for state management, avoid reactive() for complex state
- Follow Element Plus component conventions

## Git Commits
- Conventional Commits, imperative mood (feat/fix/refactor/test/docs/chore)
- Atomic commits: one logical change per commit
- Keep commits independent: each commit should pass tests
- Use `ccg:commit` skill for intelligent commit generation

## Testing
- Every feat/fix MUST include corresponding tests
- Coverage must not decrease
- Fix flow: write failing test FIRST, then fix code
- Backend: Go testing with `*_test.go` files
- Frontend: Vitest for unit tests, Playwright for E2E
- E2E tests: `scripts/test-e2e.sh` (API), `scripts/test-smpp.sh` (SMPP protocol)

## Security
- Never log secrets (tokens/keys/cookies/JWT/passwords)
- Validate inputs at trust boundaries (HTTP handlers, SMPP PDU parsing)
- Use `crypto/rand` for random ID generation (not `math/rand`)
- JWT tokens: never store in localStorage, use httpOnly cookies or memory
- SMPP: validate PDU length fields to prevent buffer overflow
- CORS: restrict origins in production (never use `*` in prod)
