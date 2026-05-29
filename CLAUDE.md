For project information, refer to PROJECT.md.
This file will detail development conventions only.

# Conventions

## OpenSpec Proposals

- Before starting a new proposal, pull the latest main branch and checkout from there
- Branch name format: `feat/<spec-name-here>`
- Generate each artifact during proposal step by step. Confirm with me before moving on to the next one.

### Unit Testing in proposals

- Always plan for unit tests on the service and handler layers
- Do not write unit tests for SQL repositories, only do integration test
- Each unit test should cover one success scenario and one failure scenario by default
- If the spec requires more scenarios, follow the spec

### Others to keep in mind during proposals
- When creating tasks for a new endpoint, always include a bruno file creation.

## Frontend UI Primitives

- UI components use `@base-ui/react`, NOT Radix UI
- `asChild` is a Radix UI pattern — it does not exist in this codebase and will cause a TypeScript error
- To render a trigger or wrapper with button styling, apply `buttonVariants()` via `className` directly on the component instead of wrapping it in `<Button asChild>`