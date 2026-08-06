# Lessons

- **No comments in code.** Do not add comments, not even package doc comments or
  `ponytail:` markers. Name things so the code reads without them. (2026-08-06)
- **Settle the concept before writing code.** For spec-shaped requests ("extend
  the scheme so X"), present the grammar, its ambiguities and the trade-offs
  first; write the file only after the rules are agreed. (2026-08-06)
- **Never lowercase or otherwise normalise user data.** Env var names and values
  are passed through exactly as written. (2026-08-06)
