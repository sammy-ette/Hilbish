//go:build midnight

package moonlight

type Callable interface {
	isLuaFunction() bool
}
