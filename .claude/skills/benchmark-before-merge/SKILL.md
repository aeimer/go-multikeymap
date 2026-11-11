---
name: benchmark-before-merge
description: Compare Go benchmarks of a branch against its base with benchstat. Use before merging or approving a pull request, and before declaring a performance-relevant change done.
---

# Benchmark before merge

Every pull request gets a benchstat comparison against its base before it merges.
A merge waits until every significant regression is explained or fixed.

## Steps

1. Run the benchmarks on the base, then on the branch.
   Keep the checkout on one branch for the whole run; switching mid-run mixes the results.
   ```sh
   go test -run=NO_TEST -bench=. -benchmem -count=6 ./... > <scratch>/bench-base.txt
   go test -run=NO_TEST -bench=. -benchmem -count=6 ./... > <scratch>/bench-branch.txt
   ```
   The runs take several minutes; start them in the background.
2. Compare:
   ```sh
   go run golang.org/x/perf/cmd/benchstat@latest <scratch>/bench-base.txt <scratch>/bench-branch.txt
   ```
3. Judge every row in `sec/op`, `B/op` and `allocs/op`.
   A row is a **regression** when benchstat shows a p-value below 0.05 and the change is worse by more than 5 %.
   Rows marked `~` are noise.
4. For each regression, find the cause in the diff and either fix it or state why the cost is accepted (for example, a correctness fix that must do more work).
   After a fix, re-run only the affected benchmarks with `-bench='<Name>'` and compare again.
5. When a change is notable (a regression you accept, or an improvement above 10 %), regenerate the benchmark block in `README.md` with `task tb`.
6. Done when every regression is fixed or explained, and the pull request description carries a short summary: the geomean per metric and every regression with its reason.
