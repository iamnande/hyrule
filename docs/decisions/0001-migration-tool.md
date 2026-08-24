# 0001: migration tool - golang-migrate over atlas

the postgres data layer needs a migration tool. two real candidates:

- **golang-migrate** (or goose) - hand-written SQL up/down files. boring,
  matches the rest of this repo's ethos (sqlc: SQL is the source of truth,
  not a DSL generating it).
- **atlas** - declarative schema-as-code with diffing. the actual analog to
  Alembic's autogenerate, and it models RLS policies as schema objects
  natively, which matters here since RLS-readiness is a stated goal (see
  conventions.md#data-layer).

going with **golang-migrate**. atlas's RLS modeling is genuinely nice, but
it's a bigger conceptual shift (declarative schema, diffed) than this repo
needs for its current size, and staying boring here means one less thing to
learn to touch a migration. revisit if the migration workflow itself starts
hurting.
