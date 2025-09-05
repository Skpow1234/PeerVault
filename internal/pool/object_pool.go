package pool

import (
	"sync"
)

// ObjectPool provides a generic pool of reusable objects
type ObjectPool[T any] struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool with the specified constructor
func NewObjectPool[T any](constructor func() T) *ObjectPool[T] {
	return &ObjectPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return constructor()
			},
		},
	}
}

// Get retrieves an object from the pool
func (op *ObjectPool[T]) Get() T {
	return op.pool.Get().(T)
}

// Put returns an object to the pool
func (op *ObjectPool[T]) Put(obj T) {
	op.pool.Put(obj)
}

// ResetFunc defines a function to reset an object to its initial state
type ResetFunc[T any] func(T)

// ResettableObjectPool provides a pool of objects that can be reset
type ResettableObjectPool[T any] struct {
	pool   sync.Pool
	reset  ResetFunc[T]
}

// NewResettableObjectPool creates a new resettable object pool
func NewResettableObjectPool[T any](constructor func() T, reset ResetFunc[T]) *ResettableObjectPool[T] {
	return &ResettableObjectPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return constructor()
			},
		},
		reset: reset,
	}
}

// Get retrieves an object from the pool
func (rop *ResettableObjectPool[T]) Get() T {
	return rop.pool.Get().(T)
}

// Put returns an object to the pool after resetting it
func (rop *ResettableObjectPool[T]) Put(obj T) {
	rop.reset(obj)
	rop.pool.Put(obj)
}
