---
name: given-when-then-tests
description: Structure Go test bodies with given/when/then comment blocks. Use when writing a new test, refactoring a test, or reviewing test code.
---

# Given/when/then tests

A test body reads as three blocks, each opened by a lower-case comment:

- `// given`: the setup, the state before the call.
- `// when`: the call under test.
- `// then`: the assertions on the result and the new state.

The blocks are plain comments; the test stays ordinary Go with `testing`, `assert` and `require`.
Test names stay Go-style: `TestType_Method_Case`.

The structure is a recommendation for readability, not a rule a linter enforces.
Write only the blocks that carry content:

- A test without setup starts at `// when`.
- A one-line assertion on the call itself goes under `// when / then`.
- A test that alternates calls and checks repeats `// when` and `// then` pairs in sequence.

A block comment replaces any old comment that says the same thing.
Keep a comment that explains a non-obvious why, and put it below the block comment.

## Plain test

```go
func TestBiKeyMap_DeleteByKeyA(t *testing.T) {
	// given
	bm := New[string, int, string]()
	require.NoError(t, bm.Set("keyA1", 1, "value1"))

	// when
	bm.DeleteByKeyA("keyA1")

	// then
	_, exists := bm.GetByKeyB(1)
	assert.False(t, exists)
}
```

## Table-driven test

The table stays.
Name input fields `given…` and expected fields `want…` where that reads naturally.
The blocks go inside the loop body or the `t.Run` closure.

```go
func TestBiKeyMap_Set_Conflict(t *testing.T) {
	tests := []struct {
		name      string
		givenKeyA string
		givenKeyB int
		wantErr   error
	}{
		{"keyA taken", "keyA1", 2, ErrKeyAConflict},
		{"keyB taken", "keyA2", 1, ErrKeyBConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			bm := New[string, int, string]()
			require.NoError(t, bm.Set("keyA1", 1, "value1"))

			// when
			err := bm.Set(tt.givenKeyA, tt.givenKeyB, "value2")

			// then
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
```

## Scope

The structure applies to `Test` functions.
`Example` functions keep their `// Output:` layout, and benchmarks keep their own shape.
