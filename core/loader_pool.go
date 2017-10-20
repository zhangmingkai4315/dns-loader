// Package core provides core for dns load generator
package core

type LoaderTicketPool interface {
	Get()
	Return()
	Active() bool
	Total() uint32
	Remainder() uint32
}

type myLoaderTicketPool struct {
	total       uint32
	poolChannel chan struct{}
	active      bool
}

func (mlp *myLoaderTicketPool) Get() {
	<-mlp.poolChannel
}
func (mlp *myLoaderTicketPool) Return() {
	mlp.poolChannel <- struct{}{}
}

func (mlp *myLoaderTicketPool) Total() uint32 {
	return mlp.total
}

func (mlp *myLoaderTicketPool) Remainder() uint32 {
	return uint32(len(mlp.poolChannel))
}

func (mlp *myLoaderTicketPool) Active() bool {
	return mlp.active
}

func NewLoaderPool(total uint32) LoaderTicketPool {
	mlp := myLoaderTicketPool{}
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
