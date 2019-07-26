# Assert

Assert is a minimal assertion library build on top of Go-lang builtin `testing`
package.

## Usage

The library exposes the following assertions:

- `Equal(t *testing.T, expected, actual interface{})`
- `NotEqual(t *testing.T, expected, actual interface{})`
- `True(t *testing.T, assertion bool)`
- `False(t *testing.T, assertion bool)`
- `Nil(t *testing.T, v interface{})`
- `NotNil(t *testing.T, v interface{})`
- `Error(t *testing.T, err error, message string...)`
- `Len(t *testing.T, length int, v interface{})`
- `Panic(t *testing.T, fn func())`

With the following aliases:

- `EQ = Equal`
- `NEQ = NotEqual`
- `OK = True`
- `Present = NotNil`
- `Err = Error`

### Example

This is how the assertions look in action:

```go
func TestFindObject(t *testing.T) {
	obj, err := factory.CreateObject(store, nil, nil)
	assert.Nil(t, err)

	object, err := store.FindObject(a.ID, nil)
	assert.Nil(t, err)

	assert.Equal(t, a.ID, object.ID)
}
```

## Extensibility

You can override failed equality assertions by overriding the `assert.Diff` function:

`Diff(t *testing.T, positive bool, expected, actual interface{})`

You can override it with a difference library like
https://github.com/go-test/deep or https://github.com/kr/pretty to show
prettier diffs.

```go
assert.Diff = func(t *testing.T, _ bool, expected, actual interface{}) {
	// Don't forget to mark the function, to hint go test to skip this
	// frame when reporting the error to the user.
	assert.Mark(t)

	if diff := deep.Equal(expected, actual); diff != nil {
		t.Fatal(diff)
	}
}
```
