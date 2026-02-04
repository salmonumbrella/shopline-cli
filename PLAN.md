## Shopline CLI Agent-Friendly Improvements Plan

### Goals
- Align CLI behavior with agent-friendly patterns (desire paths, copy/pasteable IDs).
- Make persistent flags actually influence list/query behavior.
- Improve discovery for automation (schema/help JSON).
- Update documentation and add an agent guide.

### Steps
1. Add core helper utilities
   - ID normalization for args/flags (`[resource:$id]` -> `id`).
   - Pagination/limit helper for list commands.
   - Sort and date parsing helpers.
   - Desire-path alias application across commands.
   - JSON help/introspection output.

2. Update command behavior
   - Apply ID formatting in list tables.
   - Accept formatted IDs in args and `*-id` flags.
   - Use `--limit`, `--sort-by`, `--desc` where supported.
   - Add `--from` / `--to` date filters where API supports.
   - Add `schema list` / `schema get` desire paths.

3. Documentation + agent guide
   - Update `README.md` to match actual behavior and new affordances.
   - Add `AGENTS.md` with concise workflows and conventions.

4. Tests
   - Update expected outputs where ID formatting changed.
   - Add/adjust tests for new helpers and desire-path aliases as needed.

### Acceptance
- Commands accept `[resource:$id]` in args and ID flags.
- `--limit`, `--sort-by`, `--desc`, `--from`, `--to` behave consistently.
- `shopline schema list/get` and `shopline help --json` work.
- README and agent guide reflect the new behavior.
