# Map Internals Deep Dive

## Hash Table Structure (Go 1.24+: Swiss tables)

Since Go 1.24, maps are Swiss tables with open addressing — there are no overflow chains:

- **Group** — 8 slots, each holding one key-value pair, plus a 64-bit control word with one byte per slot (empty, deleted, or the low 7 bits of the key's hash)
- **Lookup** — the upper hash bits pick a starting group; the control word is compared against the 7-bit fragment for all 8 slots at once, so most misses never touch a key; probing moves group to group
- **Table** — an array of groups, capped at 1024 slots
- **Directory** — large maps split across many tables, indexed by the upper hash bits (extendible hashing)
- **Small maps** — up to 8 entries live in a single group with no table or directory

## Memory Growth and Capacity

- **Load factor**: a table grows when it passes 7/8 full — denser than the old design's 6.5 entries per 8-entry bucket
- **Bounded growth**: a table under 1024 slots is rehashed into a new table twice its size; a full-size table splits into two, so no single insert rehashes more than one table — this replaces the old incremental evacuation
- **Tables never shrink** — deletes free slots (or leave tombstones in full groups) but return no memory, so a map that once held 1M entries keeps that footprint; copy the survivors into a new map (or let the whole map be collected) to reclaim it
- **No `cap()` function**: capacity is internal. Preallocation (`make(map[string]int, expectedSize)`) is still worthwhile for large maps to avoid repeated growth

<details><summary>Old design (pre-Go 1.24, or built with <code>GOEXPERIMENT=noswissmap</code> on Go 1.24)</summary>

Buckets of 8 entries with overflow chains; growth at 6.5 entries per bucket (or too many overflow buckets) doubled the bucket count, and entries were evacuated incrementally from `oldbuckets` during later writes.

</details>

## Preallocation

```go
// Without preallocation — multiple growths as entries are added
m := map[string]int{}

// With preallocation — allocates enough tables upfront
m := make(map[string]int, expectedSize)
```

Preallocation avoids repeated growths. The hint is approximate — Go sizes its tables so `hint` entries fit under the 7/8 load factor; a hint of 8 or less uses a single group.

## Pointers vs Values

For large value types, storing pointers reduces copy overhead:

```go
// Large struct — copied on every read/write
m := map[string]BigStruct{}  // copies large struct

// Pointer — only pointer is copied
m := map[string]*BigStruct{} // copies 8-byte pointer
```

Trade-off: pointer maps add GC pressure. For small structs (< 128 bytes), value maps are typically faster.

## `maps` Package (Go 1.21+)

| Function | Description |
| --- | --- |
| `Clone`, `Equal`, `EqualFunc` | Shallow copy and equality comparison |
| `Keys`, `Values`, `All` (1.23+) | Iterators over keys, values, or pairs |
| `Collect`, `Insert` (1.23+) | Build maps from iterators or insert entries |

See `samber/cc-skills-golang@golang-safety` skill for `Clone`, `Equal`, and sorted iteration patterns.

## Map Key Requirements

Map keys must be comparable (`==` must work). This includes:

- All numeric types, `string`, `bool`
- Pointers, channels, interfaces (compared by identity)
- Arrays of comparable types
- Structs where all fields are comparable

Slices, maps, and functions **cannot** be map keys.
