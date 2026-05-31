For project information, refer to PROJECT.md.
For code conventions, refer to CONVENTIONS.md.
This file details Claude Code specific process and behavior only.

## Decision Boundaries

Before introducing a new pattern, new abstraction, new dependency, 
or new layer boundary not already present in the codebase, stop and 
propose it to me first. Do not proceed until confirmed.


## OpenSpec Proposals

- Before starting a new proposal, pull the latest main branch and checkout from there
- Branch name format: `feat/<spec-name-here>`
- Generate each artifact during proposal step by step. Confirm with me before moving on to the next one.

### Unit Testing in proposals

- Always plan for unit tests on the service and handler layers
- Do not write unit tests for SQL repositories, only do integration tests
- Each unit test should cover one success scenario and one failure scenario by default
- If the spec requires more scenarios, follow the spec

#### Assertion quality
- If a function returns a result, assert the result — not just the error
- Failure scenarios must assert the specific error type or message, not just that an error occurred

### Others to keep in mind during proposals
- When creating tasks for a new endpoint, always include a bruno file creation.
