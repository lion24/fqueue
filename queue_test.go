package fqueue

import (
	"errors"
	"math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

type queueTests struct {
	name string
	fun  func(q Queuer[int], t *testing.T)
}

var dummy any

func flushQueue(queue Queuer[int]) {
	for i := queue.Len(); i > 0; i-- { // Flush element from the queue
		_, _ = queue.Remove()
	}
}

func TestQueues(t *testing.T) {
	cases := []struct {
		name  string
		queue Queuer[int]
	}{
		{name: "BasicQueue", queue: NewBasicQueue[int](8)},
		{name: "LinkedQueue", queue: NewLinkedQueue[int](8)},
	}

	for _, tc := range cases {
		tc := tc // pin

		queue := newQueue(tc.queue)

		testslist := []queueTests{
			{name: "Add", fun: testAdd},
			{name: "Remove", fun: testRemove},
			{name: "Len", fun: testLen},
			{name: "QueueFull", fun: testQueueFull},
		}

		t.Run(tc.name, func(t *testing.T) {
			for _, test := range testslist {
				t.Run(test.name, func(t *testing.T) {
					test.fun(queue, t)
				})
			}
		})
	}
}

func testRemove(queue Queuer[int], t *testing.T) {
	defer flushQueue(queue)

	var (
		ok  bool
		err error
	)

	ok, err = queue.Add(1, 1, 2, 3, 5, 8)
	assert.NoError(t, err)
	assert.True(t, ok)

	dummy, err = queue.Remove()
	assert.Equal(t, 1, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 1, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 2, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 3, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 5, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 8, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Empty(t, dummy)
	assert.Error(t, err, "Expected error when dequeuing from an empty queue")
}

func testAdd(queue Queuer[int], t *testing.T) {
	defer flushQueue(queue)

	var (
		err error
		ok  bool
	)

	ok, err = queue.Add(1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = queue.Add(1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = queue.Add(2)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Dequeue some elements
	dummy, err = queue.Remove()
	assert.Equal(t, 1, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 1, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 2, dummy)
	assert.NoError(t, err)

	// Enqueue some more elements

	ok, err = queue.Add(3)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = queue.Add(5)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = queue.Add(8)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Remove remaining elements previously added

	dummy, err = queue.Remove()
	assert.Equal(t, 3, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 5, dummy)
	assert.NoError(t, err)

	dummy, err = queue.Remove()
	assert.Equal(t, 8, dummy)
	assert.NoError(t, err)

	// Assert queue is empty

	expectErr := ErrQueueIsEmpty

	dummy, err = queue.Remove()
	assert.Empty(t, dummy)
	assert.Error(t, err, "Expected error when dequeuing from an empty queue")

	if !errors.Is(err, expectErr) {
		t.Errorf("Expected err: %q, got %q", expectErr, err)
	}
}

func testLen(queue Queuer[int], t *testing.T) {
	defer flushQueue(queue)

	var (
		ok  bool
		err error
	)

	elems := []int{1, 1, 2, 3}

	// Enqueue some elements
	ok, err = queue.Add(elems...)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Assert the Len of the queue is valid
	expected := uint64(len(elems))
	got := queue.Len()

	if !assert.Equal(t, expected, got) {
		t.Errorf("Len() error, expected len to be %d, got %d", expected, got)
	}

	// Add some more elements
	_, _ = queue.Add(5)
	_, _ = queue.Add(8)
	_, _ = queue.Add(11)

	expected += 3 // We add 3 more elements
	got = queue.Len()

	if !assert.Equal(t, expected, got) {
		t.Errorf("Len() error, expected len to be %d, got %d", expected, got)
	}

	// Dequeue some elements
	_, err = queue.Remove()
	assert.NoError(t, err)

	_, err = queue.Remove()
	assert.NoError(t, err)

	_, err = queue.Remove()
	assert.NoError(t, err)

	_, err = queue.Remove()
	assert.NoError(t, err)

	expected -= 4 // We removed 4 elements
	got = queue.Len()

	if !assert.Equal(t, expected, got) {
		t.Errorf("Len() error, expected len to be %d, got %d", expected, got)
	}
}

func testQueueFull(queue Queuer[int], t *testing.T) {
	defer flushQueue(queue)

	var (
		err error
		ok  bool
	)

	for i := 0; i < queue.Size() && queue.Size() != int(queue.Len()); i++ {
		// Fill the queue
		ok, err = queue.Add(i)
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	// So basically the queue is full, adding an element should raise an error
	ok, err = queue.Add(22)
	assert.False(t, ok)

	expectErr := ErrQueueIsFull

	if !errors.Is(err, expectErr) {
		t.Errorf("Expected err: %q, got %q", expectErr, err)
	}

	// Dequeue some elements
	_, err = queue.Remove()
	assert.NoError(t, err)

	_, err = queue.Remove()
	assert.NoError(t, err)

	_, err = queue.Remove()
	assert.NoError(t, err)

	_, err = queue.Remove()
	assert.NoError(t, err)

	// And Ensure we can readd some elements
	ok, err = queue.Add(12)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = queue.Add(24)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = queue.Add(32)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func BenchmarkBasicQueueAdd(b *testing.B) {
	queue := NewBasicQueue[uint64](b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Add(rand.Uint64()) //nolint: gosec
	}
}

func BenchmarkLinkedQueueAdd(b *testing.B) {
	queue := NewLinkedQueue[uint64](b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Add(rand.Uint64()) //nolint: gosec
	}
}

func BenchmarkBasicQueueRemove(b *testing.B) {
	queue := NewBasicQueue[uint64](b.N)

	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Add(rand.Uint64()) //nolint: gosec
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Remove()
	}
}

func BenchmarkLinkedQueueRemove(b *testing.B) {
	queue := NewLinkedQueue[uint64](b.N)

	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Add(rand.Uint64()) //nolint: gosec
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Remove()
	}
}

func BenchmarkBasicQueueEnqueueDequeue(b *testing.B) {
	queue := NewBasicQueue[uint64](b.N)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Add(rand.Uint64()) //nolint: gosec
	}

	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Remove()
	}
}

func BenchmarkLinkedQueueEnqueuDeque(b *testing.B) {
	queue := NewLinkedQueue[uint64](b.N)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Add(rand.Uint64()) //nolint: gosec
	}

	for i := 0; i < b.N; i++ {
		dummy, _ = queue.Remove()
	}
}
