//go:build midnight

package moonlight

type GoCont struct{}
type Cont interface{}

func (gc *GoCont) Next() Cont {
	return gc
}
