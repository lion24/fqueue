package fqueue

import "sync/atomic"

type BasicQueue[T any] struct {
	size  int
	count uint64 // Keep track about the number of element in the queue
	queue chan T
}

func NewBasicQueue[T any](size int, elems ...T) *BasicQueue[T] {
	queueChan := make(chan T, size)
	for _, item := range elems {
		queueChan <- item
	}

	return &BasicQueue[T]{
		queue: queueChan,
		size:  size,
	}
}

func (bq *BasicQueue[T]) Add(e ...T) (bool, error) {
	for _, item := range e {
		if len(bq.queue) == bq.size {
			return false, ErrQueueIsFull
		}

		bq.queue <- item
		atomic.AddUint64(&bq.count, uint64(1))
	}

	return true, nil
}

func (bq *BasicQueue[T]) Remove() (T, error) {
	if len(bq.queue) == 0 {
		return *new(T), ErrQueueIsEmpty
	}

	elem := <-bq.queue
	atomic.AddUint64(&bq.count, ^uint64(0))
	return elem, nil
}

func (bq *BasicQueue[T]) Len() uint64 {
	return atomic.LoadUint64(&bq.count)
}

func (bq *BasicQueue[T]) Size() int {
	return bq.size
}
