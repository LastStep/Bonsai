# Testing Conventions

> [!important]
> All new code must include tests. Coverage must stay at or above 80%.

---

## Rules

- Write tests for every new function, endpoint, and component
- Test the happy path AND edge cases
- Use the project's configured test framework — do not introduce alternatives
- Mock external services, not internal ones
- Never skip or disable tests without documenting why

## Structure

- Tests mirror the source tree: `app/services/foo.py` → `tests/test_foo.py`
- Use fixtures for shared setup
- One assertion focus per test — test one behavior at a time

## Before Reporting

- Run `pytest -v --cov --cov-report=term-missing` (or equivalent)
- Confirm no regressions in existing tests
- Include coverage summary in your report
