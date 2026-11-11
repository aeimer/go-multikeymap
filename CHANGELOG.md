# Changelog

## [0.5.0](https://github.com/aeimer/go-multikeymap/compare/v0.4.1...v0.5.0) (2026-11-11)


### ⚠ BREAKING CHANGES

* Go 1.26 is required (was 1.23).
* the package github.com/aeimer/go-multikeymap/container is gone. Define the interface you need in your own package instead.
* the exported embedded fields MultiKeyMap and BiKeyMap of the concurrent types are gone. Use the methods of the concurrent type instead.
* methods of all four map types are renamed to the names of the Go standard library, see the migration table below.
* `Values` returns an `iter.Seq` instead of a slice. Use `slices.Collect(m.Values())` if you need a slice.
* Delete methods return nothing. A delete of a missing key is a no-op.
* **multikeymap:** `AddSecondaryKeys` (was `PutSecondaryKeys`) returns false and stores nothing if the primary key is absent. A secondary key that belongs to another primary key moves to this one.
* **multikeymap:** `KeyGroups` (was `GetAllKeyGroups`) is removed. Use the iterators `Groups` and `GroupKeys` instead.
* **bikeymap:** `Set` (was `Put`) returns no error any more, it overwrites like a Go map. If keyA is paired with another keyB, that keyB is removed. If keyB is paired with another keyA, the entry of that keyA is removed. `Len` shrinks by at most one.

#### Migration

| v0.4.1 | v0.5.0 |
|---|---|
| `multikeymap` `Put(k, v)` | `Set(k, v)` |
| `multikeymap` `PutSecondaryKeys(k, g, keys...)` | `AddSecondaryKeys(k, g, keys...) bool` |
| `multikeymap` `HasPrimaryKey(k)` | `Has(k)` |
| `multikeymap` `Remove(k)` | `Delete(k)` |
| `multikeymap` `GetAllKeyGroups()` | `Groups()`, `GroupKeys(g)` (iterators) |
| `bikeymap` `Put(a, b, v) error` | `Set(a, b, v)` (overwrites, no error) |
| `bikeymap` `RemoveByKeyA(a) error` | `DeleteByKeyA(a)` |
| `bikeymap` `RemoveByKeyB(b) error` | `DeleteByKeyB(b)` |
| both: `Size()` | `Len()` |
| both: `Empty()` | `Len() == 0` |
| both: `Values() []V` | `Values() iter.Seq[V]` |
| both: `Get…`, `HasSecondaryKey`, `Clear`, `String` | unchanged |

New in both packages: `All`, key iterators, `Clone`, `DeleteFunc`, `LoadAndDelete…`, key lookups between keys; `multikeymap` also `LoadOrStore`, `Update`, `DeleteSecondaryKey`, `DeleteGroup`; `bikeymap` also `Pairs`.
The zero value of every type is ready to use.

### Features

* add All and key iterators ([acbcabf](https://github.com/aeimer/go-multikeymap/commit/acbcabf4088c39adb874b10f9b87865b351d2ce1))
* add atomic load-or-store, load-and-delete and update ([63c3950](https://github.com/aeimer/go-multikeymap/commit/63c39506ddef9082beef69540b5e24702b7d694b))
* add Clone and DeleteFunc ([0a670f6](https://github.com/aeimer/go-multikeymap/commit/0a670f60477989901aee6f49c7eec7f9ccf41711))
* add key lookups between primary and secondary keys ([1d7ab4d](https://github.com/aeimer/go-multikeymap/commit/1d7ab4d97f7a7c816ecbf7a517d39cd7a15ceb88))
* **bikeymap:** add Pairs iterator ([b791b93](https://github.com/aeimer/go-multikeymap/commit/b791b93534bb61b8e035f5f0d370170fd86f31dc))
* **bikeymap:** let Set overwrite conflicting pairs ([37c5ee5](https://github.com/aeimer/go-multikeymap/commit/37c5ee59e029f4c402665c609bb75b88473e4999))
* make the zero value of all map types ready to use ([aaf6729](https://github.com/aeimer/go-multikeymap/commit/aaf67292479e876df55674d0e9299a35b976ccfe))
* **multikeymap:** add LoadAndDeleteBySecondaryKey ([4193a06](https://github.com/aeimer/go-multikeymap/commit/4193a067884ecb4a4d26acd886e5ef8d3b9ffffb))
* **multikeymap:** delete secondary keys and groups ([9f2c756](https://github.com/aeimer/go-multikeymap/commit/9f2c756194b2a24283c3967ea4d694464159bd20))
* **multikeymap:** replace KeyGroups with Groups and GroupKeys iterators ([bd37ec6](https://github.com/aeimer/go-multikeymap/commit/bd37ec6be33882ae21afcafa4b845b32d01199f4))


### Bug Fixes

* **bikeymap:** lock Len in ConcurrentBiKeyMap ([a688451](https://github.com/aeimer/go-multikeymap/commit/a6884518d9d5ea3f3f91087b80ecdfe6e0e7edc1))
* **deps:** bump actions/checkout from 4 to 5 ([#13](https://github.com/aeimer/go-multikeymap/issues/13)) ([cd876bb](https://github.com/aeimer/go-multikeymap/commit/cd876bb6707488d16288dc3420f95a27f29de6d1))
* **deps:** bump actions/checkout from 5 to 6 ([#15](https://github.com/aeimer/go-multikeymap/issues/15)) ([e83de7a](https://github.com/aeimer/go-multikeymap/commit/e83de7ab763d1825ff030cd8d86d16671faf7912))
* **deps:** bump actions/checkout from 6 to 7 ([#18](https://github.com/aeimer/go-multikeymap/issues/18)) ([e61dc3e](https://github.com/aeimer/go-multikeymap/commit/e61dc3ea56749bb8fa31b5f1253b7013c56ec508))
* **deps:** bump actions/setup-go from 5 to 6 ([#14](https://github.com/aeimer/go-multikeymap/issues/14)) ([ae265f8](https://github.com/aeimer/go-multikeymap/commit/ae265f81afe685cac297d04b1a0f5b4f92367773))
* **deps:** bump actions/setup-go from 6 to 7 ([#19](https://github.com/aeimer/go-multikeymap/issues/19)) ([5691fed](https://github.com/aeimer/go-multikeymap/commit/5691fed89ba13c308928d99187bf279515b50e90))
* **deps:** bump github.com/stretchr/testify from 1.10.0 to 1.11.0 ([#10](https://github.com/aeimer/go-multikeymap/issues/10)) ([f980415](https://github.com/aeimer/go-multikeymap/commit/f980415912932b1b720a326f6056e6f8f851c0ec))
* **deps:** bump github.com/stretchr/testify from 1.11.0 to 1.11.1 ([#12](https://github.com/aeimer/go-multikeymap/issues/12)) ([bff8d91](https://github.com/aeimer/go-multikeymap/commit/bff8d91a922cf9b2811be3cff2d943fabaa350c5))
* **deps:** bump github.com/stretchr/testify from 1.11.1 to 1.12.0 ([#20](https://github.com/aeimer/go-multikeymap/issues/20)) ([6fef2b7](https://github.com/aeimer/go-multikeymap/commit/6fef2b7063ebb7c1be54fbdc25497fe0aa252e88))
* **deps:** bump github.com/stretchr/testify from 1.12.0 to 1.12.1 ([#21](https://github.com/aeimer/go-multikeymap/issues/21)) ([c69bc23](https://github.com/aeimer/go-multikeymap/commit/c69bc23cc30324297a584d8f6484b23f9a2977b8))
* **deps:** bump golangci/golangci-lint-action from 7 to 8 ([#9](https://github.com/aeimer/go-multikeymap/issues/9)) ([e955944](https://github.com/aeimer/go-multikeymap/commit/e955944da975a8547b2213ef9a0f9178bffc7db2))
* **deps:** bump golangci/golangci-lint-action from 8 to 9 ([#16](https://github.com/aeimer/go-multikeymap/issues/16)) ([40c0eea](https://github.com/aeimer/go-multikeymap/commit/40c0eeaf77d76976f80f20dce05e0784dc6baeb3))
* **deps:** bump googleapis/release-please-action from 4 to 5 ([#17](https://github.com/aeimer/go-multikeymap/issues/17)) ([e93c5f9](https://github.com/aeimer/go-multikeymap/commit/e93c5f962cefc5eb9f0b418dfd500c67094dfdc7))
* format String of concurrent maps outside the lock ([db2a446](https://github.com/aeimer/go-multikeymap/commit/db2a44687763df1e718a6508bee89e1676a52a29))
* hide embedded map in concurrent types ([86633d5](https://github.com/aeimer/go-multikeymap/commit/86633d529ada4eaa0c4701b25d244d4e89f116d6))
* keep iteration consistent when Clear runs during a range ([a2cf40a](https://github.com/aeimer/go-multikeymap/commit/a2cf40a423c2ccb7753f4db3364fcc833f9d8fb6))
* **multikeymap:** report missing primary key in AddSecondaryKeys ([3b7cc8a](https://github.com/aeimer/go-multikeymap/commit/3b7cc8a7cf039793ea602591264a9e9e928eb343))
* **multikeymap:** track all secondary keys per group in reverse index ([2635368](https://github.com/aeimer/go-multikeymap/commit/2635368d7c69d2b18c0bd48f878badbcdba5fa7b))


### Performance Improvements

* **bikeymap:** store keyB next to the value ([8a57a68](https://github.com/aeimer/go-multikeymap/commit/8a57a687713e19b86c2df7a95d52ca62faa4f71a))


### Code Refactoring

* **bikeymap:** align method names with the standard library ([8cfb0f4](https://github.com/aeimer/go-multikeymap/commit/8cfb0f439cdbbf5d0fd96b1c4a222fa69c389313))
* **multikeymap:** align method names with the standard library ([9e27b26](https://github.com/aeimer/go-multikeymap/commit/9e27b26df414adbee393726e4829b3ad191e87f0))
* remove container package ([c5550c2](https://github.com/aeimer/go-multikeymap/commit/c5550c20f7c5b3322975bdc147e4f356ee32e62b))

## [0.4.1](https://github.com/aeimer/go-multikeymap/compare/v0.4.0...v0.4.1) (2025-04-08)


### Bug Fixes

* **deps:** bump golangci/golangci-lint-action from 6 to 7 ([#7](https://github.com/aeimer/go-multikeymap/issues/7)) ([5134765](https://github.com/aeimer/go-multikeymap/commit/5134765fe614ebf5ddc198c049e8cc376aa7819e))

## [0.4.0](https://github.com/aeimer/go-multikeymap/compare/v0.3.0...v0.4.0) (2025-01-28)


### Features

* supply concurrent an non-concurrent verions ([#5](https://github.com/aeimer/go-multikeymap/issues/5)) ([5178093](https://github.com/aeimer/go-multikeymap/commit/517809354a6f882e354c42bd4219ed49bdf31097))

## [0.3.0](https://github.com/aeimer/go-multikeymap/compare/v0.2.0...v0.3.0) (2024-11-23)


### Features

* adds benchmarks ([0098d28](https://github.com/aeimer/go-multikeymap/commit/0098d28992171bc5102c09b62c9bd66be347068b))
* uses common function naming Put and Remove like GoDS ([1c67161](https://github.com/aeimer/go-multikeymap/commit/1c67161e543b97316d15c2b964e8583492bf74a1))


### Code Refactoring

* renames Add to Set ([b741f31](https://github.com/aeimer/go-multikeymap/commit/b741f31b9e2bfaa9eea5988498331b1867acad7c))

## [0.2.0](https://github.com/aeimer/go-multikeymap/compare/v0.1.0...v0.2.0) (2024-11-18)


### Features

* removes HasKey and GetKeyGroups ([6a9b946](https://github.com/aeimer/go-multikeymap/commit/6a9b946b180a29e97739529443dc813777624c34))

## 0.1.0 (2024-11-18)


### Features

* initial code ([4fbec6a](https://github.com/aeimer/go-multikeymap/commit/4fbec6a452938ae31cf2cf22734b406950c92826))


### Miscellaneous Chores

* release 0.1.0 ([4dd064a](https://github.com/aeimer/go-multikeymap/commit/4dd064a84b3b3e1073ab812f9cf105d10fe05e40))
