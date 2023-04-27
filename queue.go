package fqueue

import (
	"errors"
	"fmt"
)

var (
	ErrQueueIsEmpty = errors.New("queue is empty")
	ErrQueueIsFull  = errors.New("queue is full")
)

// A special queue which we can peek element at.
type QueuePeeker[T any] interface {
	// Retrieve, but does not remove the head of the queue, or return nil if the
	// queue is empty
	Peek() T
}

//go:generate mockery --name Queuer --case snake
type Queuer[T any] interface {
	// Insert one or more element of type T in the queue if it is possible to do so
	// immediately without violating capacity restrictions, return true uppon
	// success and return ErrQueueFull if not space is currently available.
	Add(...T) (bool, error)

	// Retrieve and remove the head of this queue
	Remove() (T, error)

	// Return the number of element sitting in the queue
	Len() uint64

	// Return the total capacity of the queue
	Size() int
}

type QueuingKind uint

const (
	Unknown QueuingKind = iota
	Basic
)

func (q QueuingKind) String() string {
	switch q {
	case Basic:
		return "BasicQueue"
	default:
		return fmt.Sprintf("Invalid Queue type: %d", q)
	}
}

type queue[T any] struct {
	q Queuer[T]
}

func newQueue[T any](size int, kind QueuingKind) *queue[T] {
	switch kind {
	case Basic:
		return &queue[T]{
			q: NewBasicQueue[T](size),
		}
	}

	return nil
}

func (q *queue[T]) Add(e ...T) (bool, error) { return q.q.Add(e...) }
func (q *queue[T]) Remove() (T, error)       { return q.q.Remove() }
func (q *queue[T]) Len() uint64              { return q.q.Len() }
func (q *queue[T]) Size() int                { return q.q.Size() }
