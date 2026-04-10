# Database Conventions

---

## Naming

- Tables: `snake_case`, plural (e.g., `characters`, `chapter_events`)
- Columns: `snake_case` (e.g., `created_at`, `chapter_id`)
- Primary keys: `id` (UUID preferred)
- Foreign keys: `{referenced_table_singular}_id` (e.g., `chapter_id`)
- Indexes: `idx_{table}_{column(s)}`
- Constraints: `{type}_{table}_{column}` (e.g., `uq_characters_name`, `fk_events_chapter_id`)

## Migrations

- One migration per logical change — never bundle unrelated schema changes
- Migrations must be reversible (include both up and down)
- Never modify a migration that has been applied to any environment
- Test migrations against a copy of production data when possible

## Schema Rules

- Every table gets `created_at` and `updated_at` timestamps
- Use `UUID` for primary keys, not auto-increment integers
- Add `NOT NULL` constraints by default — only allow NULL when there's a reason
- Add indexes on foreign keys and frequently queried columns
