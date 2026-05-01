For project information, refer to PROJECT.md.
This file will detail development conventions only.

# Conventions

## OpenSpec Proposals

- Before starting a new proposal, pull the latest main branch and checkout from there
- Branch name format: `feat/<spec-name-here>`

## Unit Testing

- Always plan for unit tests on the service and handler layers
- Do not write unit tests for SQL repositories (integration tests will be addressed later)
- Each unit test should cover one success scenario and one failure scenario by default
- If the spec requires more scenarios, follow the spec
