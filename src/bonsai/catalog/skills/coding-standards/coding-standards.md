# Coding Standards

> [!important]
> These rules are non-negotiable. Every line of code must conform.

---

## General

- Follow the language's official style guide
- All code must pass the project's configured linter and formatter
- Type annotations on all public functions and methods
- Google-style docstrings on all public functions and classes

## File Creation Rules

- Every directory with source files must have proper module markers (e.g., `__init__.py`)
- All imports use the project's top-level package — never relative to the repo root
- Never hardcode configuration values — everything through environment/config
- Never put business logic in route handlers — handlers call services

## Testing

- All new code must have tests
- Tests must be async where the code under test is async
- Run the full test suite before reporting completion
