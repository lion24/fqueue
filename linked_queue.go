package fqueue

import (
	"sync/atomic"
	"unsafe"
)

type LinkedQueue[T any] struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	size  int
	count uint64
}

type Node[T any] struct {
	next  unsafe.Pointer
	value T
}

func NewLinkedQueue[T any](size int) *LinkedQueue[T] {
	n := unsafe.Pointer(NewNode[T]())
	return &LinkedQueue[T]{
		head: n,
		tail: n,
		size: size,
	}
}

func NewNode[T any]() *Node[T] {
	return &Node[T]{}
}

func (q *LinkedQueue[T]) Add(e ...T) (bool, error) {
	for _, elem := range e {
		node := NewNode[T]()
		node.value = elem

		for {
			tail := load[T](&q.tail)
			next := load[T](&tail.next)

			if tail == load[T](&q.tail) {
				if next == nil { // Tail was point to last node?
					if q.Len() == uint64(q.size) {
						return false, ErrQueueIsFull
					}

					// try to link node at the end of the linked list
					if cas(&tail.next, next, node) {
						cas(&q.tail, tail, node) // Enqueue done try to swing tail to inserted node
						atomic.AddUint64(&q.count, uint64(1))
						break
					}
				} else {
					cas(&q.tail, tail, next)
					break
				}
			}
		}
	}

	return true, nil
}

// Retrieve and remove the head of this queue
func (q *LinkedQueue[T]) Remove() (T, error) {
	for {
		head := load[T](&q.head)
		tail := load[T](&q.tail)
		next := load[T](&head.next)

		if head == load[T](&q.head) {
			if head == tail { // Is queue empty of head fally behind?
				if next == nil {
					return *new(T), ErrQueueIsEmpty
				}

				// Tail is falling behing try to advance it
				cas(&q.tail, tail, next)
			} else {
				// Read value before CAS, otherwise another dequeue might free the next Node
				v := next.value

				if cas(&q.head, head, next) {
					atomic.AddUint64(&q.count, ^uint64(0))
					return v, nil // Dequeue is done return the element
				}
			}
		}
	}
}

// Return the number of element sitting in the queue
func (q *LinkedQueue[T]) Len() uint64 {
	return atomic.LoadUint64(&q.count)
}

// Return the total capacity of the queue
func (q *LinkedQueue[T]) Size() int {
	return q.size
}

func load[T any](p *unsafe.Pointer) *Node[T] {
	return (*Node[T])(atomic.LoadPointer(p))
}

func cas[T any](p *unsafe.Pointer, old, _new *Node[T]) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(_new))
}
