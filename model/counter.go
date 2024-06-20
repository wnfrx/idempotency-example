package model

import "sync"

type Counter struct {
	sync.Mutex
	value int
}

func (c *Counter) Increment() (value int) {
	c.Lock()
	c.value++
	value = c.value
	c.Unlock()
	return
}

func (c *Counter) Value() int {
	return c.value
}
