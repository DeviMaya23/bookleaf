# Project Information

> This is the main entry point for agents working in this repo.
> Read this first. Then read only the reference file(s) for your work type.

## What is Bookleaf

Web-based image moodboarding app (inspired by Raindrop.io). Users organise images into folders; optional AI labelling surfaces suggested organisation.

## Booklet

Booklet is Bookleaf's sister app. Both apps share a single Kinde user pool.

- Bookleaf exposes a set of internal endpoints that Booklet calls (e.g. to validate and resolve public folder data for character pages).
- Account deletion is orchestrated by Bookleaf: when a user's account is deleted, Bookleaf coordinates cleanup across both apps.
- Booklet's code and data are not in this repo. If a task seems to require a Booklet-side change, stop and flag it.

## Key Domain Concepts

- **User** — authenticated user; holds a `vision_enabled` flag to opt into AI organising.
- **Image** — uploaded asset stored in Cloudflare R2. Carries metadata: R2 path, thumbnail path, folder assignment, MIME type, AI labels.
- **Folder** — user-managed grouping of images; supports manual hierarchy.

## Commands

Refer to `Makefile` for commands to run BE/FE/build extensions.  
For local development and DB migrations, refer to `Makefile.local`.  
Refer to `.env.example` for backend/frontend/extension env vars.

## Reference Files

Read these before the work type they cover — not all at once, only the relevant one:

| When you're about to… | Read |
|---|---|
| Orient to the project stack or pick a library | `STACK.md` |
| Place a new file, create a feature, or touch an existing structure | `ARCHITECTURE.md` |
| Write any frontend, backend, or extension code | `CONVENTIONS.md` |
| Write or plan any tests | `TESTING.md` |

## Decision Boundaries

Before introducing a new pattern, new abstraction, new dependency,
or new layer boundary not already present in the codebase, stop and
propose it to me first. Do not proceed until confirmed.

## Mid-Implementation Adaptations

If, while implementing a task, you discover that something not covered by the design is needed — a structural workaround, a bridging type, a scope change, an unanticipated dependency — stop and inform me before proceeding. Describe what you found, why the adaptation is needed, and what you're proposing. Do not implement it silently.

## OpenSpec Proposals

- Before starting a new proposal, pull the latest main branch and checkout from there
- Branch name format: `feat/<spec-name-here>`
- Generate each artifact during proposal step by step. Confirm with me before moving on to the next one.

### Unit Testing in proposals

- Always plan for unit tests on the usecase and handler layers
- Do not write unit tests for SQL repositories, only do integration tests
- Follow the unit testing rules in `TESTING.md` — they define what scenarios are worth writing and what test doubles to use
- Do not default to one success + one failure per function; write only the scenarios that have a reason to exist per the conventions

#### Assertion quality
- If a function returns a result, assert the result — not just the error
- Failure scenarios must assert the specific error type or message, not just that an error occurred

### Others to keep in mind during proposals
- When a change modifies a shared contract (a function signature, API endpoint, hook, etc.), grep for every call site of that symbol across all layers (backend, frontend, extension) before finalizing Impact/Capabilities/tasks — do not rely on a named flow (e.g. "the upload flow") to be exhaustive, as parallel entry points (e.g. drag-and-drop vs. modal vs. batch upload) commonly funnel into the same shared function and are easy to miss.
- When creating tasks for a new endpoint, always include a bruno file creation.
- On any BE development, include a task to run golang-ci lint at the end, and fix whatever issue arises.
- On any FE development, include a task to run npm run build and npm run lint at the end, and fix whatever issue arises.
