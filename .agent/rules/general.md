---
trigger: always_on
---

# Antigravity Agent Rules / System Prompt

You are assisting the user in a repository that builds a mini version of Temporal (a workflow engine) from scratch. 
You MUST adhere to the following strict guidelines during all interactions:

## Language and Architecture
1. **Language**: All code must be written in Go. You must adhere strictly to Go community best practices and idioms (e.g., standard layout, effective Go).
2. **Architecture**: The project uses **Hexagonal Architecture** (Ports and Adapters). Organize the directory structure to reflect this, keeping domain logic isolated and ensuring easy mocking of external dependencies.

## Coding Standards
1. **Comments**: Do NOT add any redundant or obvious comments. ONLY add comments if the user explicitly requests them. When requested, comments must read naturally and be human-like.
2. **Testing**: All methods and modules MUST be thoroughly tested. You must use `mockery` for generating mocks and writing unit tests.
3. **Dependencies**: When downloading or adding a new Go package/dependency, ensure it aligns with Go best practices (e.g., use `go get`, update `go.mod` appropriately, and avoid unnecessary or unmaintained dependencies).
4. **Planning**: Always try to planning first before execute anything, interact and ask me critical questions to make sure you clarify everything.

## Documentation
1. **README.md**: Do NOT create, modify, or summarize the `README.md` file unless the user explicitly commands you to do so.