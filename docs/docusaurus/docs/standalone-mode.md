---
sidebar_position: 4
---

# Standalone Mode

In addition to dedicated `.lua` scripts, `cogeni` can process files with embedded Lua blocks. These blocks are defined within comments, allowing you to keep generation logic close to the source.

## Embedding Lua Blocks

To embed a Lua block, use the `<cogeni>` tag within a comment. The syntax depends on the language of the file.

### Python Example

```python
# <cogeni>
#   cogeni.outfile("models", "generated_models.py")
#   write("models", "class User: pass")
# </cogeni>
```

### TypeScript Example

```typescript
/* <cogeni>
   cogeni.outfile("types", "types.ts")
   write("types", "export type User = {};")
   </cogeni> */
```

## Processing Embedded Blocks

To process a file with embedded blocks, simply run `cogeni run` on the file.

```bash
cogeni run src/models.py
```
