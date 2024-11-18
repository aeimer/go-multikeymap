package container

// Container is the interface that all multikeymaps must implement.
type Container[T any] interface {
	Empty() bool
	Size() int
	Values() []T
	Clear()
	String() string
}
