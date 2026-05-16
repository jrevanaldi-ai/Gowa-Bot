package lib

type Dispatcher struct {
	semaphore chan struct{}
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	return &Dispatcher{
		semaphore: make(chan struct{}, maxWorkers),
	}
}

func (d *Dispatcher) Run(fn func()) {
	d.semaphore <- struct{}{}
	go func() {
		defer func() { <-d.semaphore }()
		defer func() {
			if r := recover(); r != nil {
				_ = r
			}
		}()
		fn()
	}()
}

func (d *Dispatcher) InFlight() int {
	return len(d.semaphore)
}

func (d *Dispatcher) Capacity() int {
	return cap(d.semaphore)
}
