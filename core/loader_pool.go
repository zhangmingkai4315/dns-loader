// Package core provides core for dns load generator
package core

type LoaderPool interface {
	Get()
	Return()
	Active() bool
	Total() uint32
	Remainder() uint32
}

type myLoaderPool struct {
	total       uint32
	poolChannel chan struct{}
	active      bool
}

func (mlp *myLoaderPool) Get() {
	<-mlp.poolChannel
}
func (mlp *myLoaderPool) Return() {
	mlp.poolChannel <- struct{}{}
}

func (mlp *myLoaderPool) Total() uint32 {
	return mlp.total
}

func (mlp *myLoaderPool) Remainder() uint32 {
	return uint32(len(mlp.poolChannel))
}

func (mlp *myLoaderPool) Active() bool {
	return mlp.active
}

func NewLoaderPool(total uint32) LoaderPool {
	mlp := myLoaderPool{}
	size := int(total)
	ch := make(chan struct{}, total)
	for i := 0; i < size; i++ {
		ch <- struct{}{}
	}
	mlp.poolChannel = ch
	mlp.total = total
	mlp.active = true
	return &mlp
}
