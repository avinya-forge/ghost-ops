---
trigger: always_on
glob:
description: Vishwa-Vani project rules — loaded every session
---

# Vishwa-Vani: Agent Rules & Code Style Guide

## 🌌 Project Identity
**Vishwa-Vani** is a Next.js 16 / React 19 / TypeScript application — the "Vedic Wikipedia". It serves Bhagavad Gita, Mahabharata, Upanishads, and other Sanskrit scriptures with an AI-enhanced reading experience. Deployed on Vercel.

**Stack**: Next.js 16 · React 19 · TypeScript 5 (strict) · Tailwind CSS 4 · next-intl · sql.js (SQLite WASM) · Jest 30 · ESLint 9

**Repo root**: `D:\Code\avinya-forge\vishwa-vani`
**VM path** (when mounted): `/sessions/.../mnt/vishwa-vani`

---

## 📋 SDLC Process — Read This Every Session

### Startup Protocol (every session)
1. **SYNC GIT**: Pull latest `main` from remote.
2. Read `docs/backlog.md` — find the **top unchecked task** in the current sprint.
3. Read `docs/standards.md` — recall the lean template and quality rules.
3. Pick ONE task, mark it `[IN PROGRESS]` in the backlog comment (do not edit the file yet).
4. Implement incrementally. After EACH task:
   - Run `npm run lint` → fix all errors.
   - Run `npm run build` → fix all errors.
   - Run `npm test` → all tests must pass.
   - Mark task `[x]` in `docs/backlog.md`.
   - Add a brief entry under the current version in `docs/release-notes.md`.
5. After an entire **Epic** is done:
   - Run `npm run test -- --coverage` → check coverage.
   - Fix any regressions before starting the next epic.
   - Update the release notes with an epic-complete entry.
   - If creating a PR or pushing, **check remote main, rebase logically, and automatically resolve conflicts** without prompting the user.

### Golden Rules
- **Flat Documentation**: The `docs/` folder is flat. NEVER create subfolders (like `docs/planning`). This is a hard rule. AI-generated artifacts or scratchpads MUST NOT be pushed to git.
- **One task at a time.** Never start task N+1 before task N is lint-clean, build-clean, and test-passing.
- **Stability Gate (STAB-608) is active.** No new feature epics until EPIC 7 (STABILITY) is fully complete.
- **Gold Standard Only.** Only 100% complete, verified books go to `data/3-gold/`. No placeholders in production. No partial books in UI.
- **Incremental only.** No big-bang rewrites. Every commit must leave the codebase in a working state.
- **Zero broken windows.** No `TODO` comments or placeholders (e.g. `[PLACEHOLDER_...]`) in production code paths.

---

## 🏗️ Architecture Summary

```
app/                  Next.js App Router pages & API routes
  api/synthesize/     AI synthesis endpoint (POST)
  [text]/[chapter]/   Dynamic scripture reader pages
  lab/                Interactive Vedic lab micro-apps
  search/             Search page
components/
  shloka/             Core reading UI (study-client.tsx, shloka-mask.tsx, ...)
  lab/                Lab app components (chhanda-analyzer, grammar-tokenizer, ...)
  ui/                 Shared UI primitives (hierarchical-nav, ...)
lib/
  texts.ts            SINGLE SOURCE OF TRUTH for all registered texts (VEDIC_LIBRARY)
  data-service.ts     VedicDataService — unified data access layer
  lake.ts             SQLite WASM client-side lake reader
  server-lake.ts      Server-side SQLite reader
  nvf.ts              NVF schema types
data/                 Sharded JSON and SQLite lake files (GOLD tier data)
docs/
  vision.md           Strategic direction (read-only reference)
  planning/backlog.md The authoritative task list
  release-notes.md    Completed work history
  architecture/blueprint.md  Architecture principles
  rules/standards.md  Coding standards and UI template rules
__tests__/            Unit tests (Jest)
```

---

## 🎯 Coding Standards (Ultra-Lean)

### TypeScript
- **`strict: true` is mandatory** — already in `tsconfig.json`. Never disable it.
- **No `any`** without an explicit `// eslint-disable-next-line @typescript-eslint/no-explicit-any` comment AND a documented reason in the same line.
- Prefer `unknown` over `any` for external data. Narrow with type guards.
- Use `interface` for object shapes, `type` for unions/intersections.
- No unused variables or imports — lint will catch these.

### Naming Conventions
- Files: `lowercase-kebab-case.tsx` / `.ts`
- Components: `PascalCase`
- Hooks: `use-` prefix, `camelCase` (e.g., `useDebounce`)
- Constants: `SCREAMING_SNAKE_CASE` for module-level, `camelCase` for local
- Data hooks: `use-` prefix

### File Size & Complexity
- Max ~300 lines per component file. Extract sub-components when exceeded.
- Keep functions under 50 lines. Extract helpers.
- One component per file. Named exports for utilities, default export for pages/components.
- No barrel `index.ts` files unless they simplify imports significantly.

### React / Next.js
- `'use client'` only when genuinely needed (event handlers, state, browser APIs).
- Prefer Server Components by default (SSG).
- No `useEffect` for data that can be fetched at build time.
- Avoid `localStorage` on first render — wrap in `useEffect` to prevent hydration mismatch.
- `key` props must be stable and unique (never array indices for dynamic lists).

### CSS / Tailwind
- Use Tailwind utility classes only. No inline styles except for dynamic values.
- Dark mode: use `dark:` prefix consistently.
- Responsive: mobile-first (`sm:`, `md:`, `lg:`).

### API Routes
- Validate all inputs before processing.
- Return consistent `{ success: boolean, data?: T, message?: string }` shape.
- No unhandled `Promise` rejections — always try/catch.
- Log errors with `console.error` and return structured error responses.

---

## 🧪 Testing Standards

### Coverage Target
- **Floor**: 80% statements/branches/functions/lines (enforced by Jest `coverageThreshold`).
- **Goal**: 95% for core product paths (`StudyClient`, `SearchClient`, `VedicDataService`, API synthesis).

### Test Requirements
- Every new component gets at minimum: render test, happy-path interaction test, edge-case test.
- Every API route gets: valid input test, missing field test, invalid field test.
- Every utility function gets: unit tests for all branches.
- Tests live in `__tests__/` mirroring the source path (e.g., `__tests__/lib-texts.test.ts`).
- Use `@testing-library/react` for component tests. Mock `next/navigation` and `next-intl`.
- No snapshot tests — they are brittle. Use explicit assertions.

### Mocks Policy
- Mock only what is truly external (network, filesystem, WASM).
- Do not mock the module under test.
- PoC/placeholder code in production must be gated with a feature flag or clearly labeled in the UI.

---

## 🔍 Lint & Build Rules

### ESLint (eslint.config.mjs)
- Next.js recommended + core-web-vitals rules are active.
- `@typescript-eslint` rules are active (no-explicit-any, no-unused-vars, etc.).
- `npm run lint` must exit 0 before any commit.

### Build
- `npm run build` must exit 0 with zero TypeScript errors before any commit.
- Do not disable `strict` or add `// @ts-ignore` to make builds pass — fix the root cause.

### Git Hygiene
- Commit message format: `type(scope): description` (e.g., `fix(study-client): correct author toggle logic`).
- Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `perf`.
- No stray `console.log` in committed code (use `console.error` for error logging only).
- No `TODO` comments — convert to backlog tasks instead.

---

## 📐 UI Template Standard (Lean Template)

The Lean UI Template governs ALL scripture reading interfaces:

| Rule | Detail |
|------|--------|
| Base layer | Sanskrit + English meaning always visible |
| Commentary | Hidden by default (opt-in) |
| Author limit | Max 2 authors selected simultaneously |
| Author toggle | Selecting a 3rd replaces the oldest |
| AI synthesis | Always available; uses meaning + up to 2 commentaries |
| Language filter | Applies to commentary only |
| Initial state | `scholarSelection = []`, `languageSelection = 'all'` |
| Persistence | `localStorage` for scholar pref and reading position |
| Responsive | Inline toolbar on desktop, stacked on mobile |

---

## 📚 Data Architecture

- **NVF 1.0**: The frozen JSON schema for all scripture data. Fields: `id`, `chapter`, `verse`, `original` (Sanskrit), `transliteration`, `meaning` (English), `layers` (commentary by author).
- **Data tiers**: BRONZE (raw) → SILVER (structured) → GOLD (audited, in `data/`).
- **Storage**: `'json'` (sharded chapter files) or `'lake'` (SQLite WASM via `lib/lake.ts`).
- **Registration**: Add entries to `VEDIC_LIBRARY` in `lib/texts.ts` — the framework auto-discovers them.
- **Audit**: Run `python scripts/vishwa.py audit` before promoting any data to GOLD.

---

## 🚦 Current Sprint Priority (STABILITY GATE — EPIC 7)

**All of EPIC 7 must be complete before starting any new feature work.**

In order:
1. `STAB-601` Verification Audit
2. `STAB-602` Placeholder Removal
3. `STAB-603` Endpoint Hardening
4. `STAB-604` UI Behavior Audit
5. `STAB-605` Coverage Audit (baseline: 56.69%)
6. `STAB-606` Coverage Remediation (target: 80% floor → 95% goal)
7. `INFRA-001` ESLint TypeScript Enhancement
8. `INFRA-002` Jest Coverage Threshold Enforcement
9. `STAB-607` Documentation Verification
10. `STAB-608` Stability Gate cleared → unlock EPIC 10 (Mahabharata) and EPIC 6 (Labs)

After stability gate:
- EPIC 10 (MBH-203, MBH-204, MBH-205, MBH-303, MBH-304, MBH-305)
- EPIC 9 (TMPL-903, TMPL-904)
- EPIC 6 (LAB-802, LAB-803, APP-702, APP-703, ...)

---

## 📱 Phone / Remote Session Notes

When starting from a phone or new session:
1. Always mount `D:\Code\avinya-forge\vishwa-vani` first.
2. Read this file (`/.agents/rules/code-style-guide.md`) — it auto-loads on every session.
3. Read `docs/backlog.md` for the current task.
4. Follow the Startup Protocol above.
5. Never start two tasks in parallel — one thing at a time.
