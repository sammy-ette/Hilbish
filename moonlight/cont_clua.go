//go:build midnight

package moonlight

type GoCont struct{}
type Cont any

func (gc *GoCont) Next() Cont {
	return gc
}
