# proposals

one directory per initiative worth planning before building - a PRD (problem,
requirements, scope) and a tech spec (design, interfaces, requirements
coverage) written and agreed before execution starts. tracked normally, same
as [docs/decisions](../decisions) - not a scratch/gitignored planning
directory.

format: `NNNN-slug/`, numbered in order made, own sequence separate from
`docs/decisions`. each directory holds `prd.md` and `tech-spec.md`. not every
change needs one - reserve it for work large enough that alignment before
execution is worth the pass (see senzu's planning phase). a proposal that
surfaces a deliberate decision worth not re-litigating still gets its own
`docs/decisions` entry - the two aren't a substitute for each other.
