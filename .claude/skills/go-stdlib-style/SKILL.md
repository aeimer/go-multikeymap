---
name: go-stdlib-style
description: Go API design the way the Go team builds the standard library - naming, return values, errors, iterators, zero values, concurrency, docs. Use when designing or reviewing an exported Go API, naming methods or types, choosing between bool, error and panic, or planning breaking changes before a v1. Complements modern-go-guidelines, which covers language features per Go version.
---

# Go stdlib style

The yardstick is the standard library: shape every exported API so it reads like `maps`, `slices`, `sync`, `container/list`, `io`.
When a choice is open, find the closest stdlib precedent, name it, and follow it.
Older third-party conventions (Java-style collections such as gods) lose to the stdlib precedent.

Sources of truth, in order: the stdlib itself, Effective Go, Go Code Review Comments, the Google Go Style Guide.

## Naming

- Short, plain verbs from the stdlib vocabulary: `Get`/`Load`, `Set`/`Store`, `Delete`, `Len`, `Clear`, `Clone`, `All`, `Keys`, `Values`.
- `Len()`, not `Size()`; `Delete`, not `Remove`; an emptiness check is `Len() == 0`.
- Getters carry no `Get` prefix: `Owner()`, `Groups()`. `Get` stays for keyed lookup (`Get(key)`).
- The package name is part of the identifier: `multikeymap.New`, not `multikeymap.NewMultiKeyMap`.
- Initialisms keep their case: `ID`, `URL`, `JSON`.
- Receiver names are one or two letters, the same across all methods of a type.

## Return values

One rule, derived from the builtin map and `sync.Map`:

| Situation | Shape | Precedent |
| --- | --- | --- |
| Lookup | `(V, bool)` comma-ok | `v, ok := m[k]`, `sync.Map.Load` |
| Insert that cannot fail | no return | `m[k] = v`, `sync.Map.Store` |
| Delete of a missing key | no-op, no return | `delete(m, k)`, `sync.Map.Delete` |
| Delete that reports | `(V, bool)` | `sync.Map.LoadAndDelete` |
| Failed precondition, absent key | `bool` | `sync.Map.CompareAndSwap` |
| Several causes the caller must tell apart | `error` with `ErrX` sentinels | `io.EOF`, `fs.ErrNotExist` |
| Programmer error, broken invariant | `panic` | `slices.Insert` out of range |

On a `false` or an error the operation leaves the data unchanged.

## Errors

- Sentinels are exported `var ErrX = errors.New("pkg: x")`, matched with `errors.Is`.
- Typed errors end in `Error` and are matched with `errors.As`.
- Messages are lower case, without trailing punctuation, prefixed with the package name.
- Wrap with `fmt.Errorf("...: %w", err)` when adding context.

## Iteration

- Range-over-func iterators: `All() iter.Seq2[K, V]`, `Keys() iter.Seq[K]`, `Values() iter.Seq[V]`.
- Callers collect with `slices.Collect` / `maps.Collect`; the API returns iterators, not fresh slices.
- Document iteration order; "unspecified, like a Go map" is a valid statement.
- Concurrent types document whether an iterator sees a snapshot or live data, and whether it holds a lock.

## Zero value

- The zero value is ready to use, like `sync.Mutex`, `bytes.Buffer`, `strings.Builder`: maps are created lazily on first write.
- `New` exists for convenience or options, never as the only valid way to construct.

## Concurrency

- Plain types are not safe for concurrent use; say so in the type doc.
- A concurrent variant holds the plain type in an unexported field and delegates under the lock: one implementation, one place to fix.
- Exported fields and embedded types never expose state that bypasses the lock.
- Every method touching shared state takes the lock, reads included (`Len`).
- Values that hold a mutex are never copied: pointer receivers, and `go vet` copylocks stays clean.
- Atomic compound operations follow `sync.Map`: `LoadOrStore`, `LoadAndDelete`, `CompareAndSwap`.

## Interfaces and generics

- Accept interfaces, return concrete types.
- Define an interface at the consumer, small (one to three methods), when a second implementation exists. A speculative interface nobody asserts gets deleted.
- Type parameters use the stdlib constraints: `comparable`, `cmp.Ordered`, `any`.
- Assert intended interface satisfaction at compile time: `var _ fmt.Stringer = (*Map[int, int])(nil)`.

## Docs

- Every exported identifier has a doc comment that starts with its name and states behaviour, edge cases, and concurrency safety.
- Runnable `Example` functions with `// Output:` carry the usage docs on pkg.go.dev.

## Versioning

- Before v1, collect breaking changes and land them together; after v1, follow the Go module compatibility promise (new major = `/v2` import path).
- Mark breaking commits with `!` and a `BREAKING CHANGE:` footer.
