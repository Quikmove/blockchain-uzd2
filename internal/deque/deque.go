package deque

type Deque[T any] struct {
	front, back *Node[T]
}
type Node[T any] struct {
	value T
	prev  *Node[T]
	next  *Node[T]
}

func New[T any]() *Deque[T] {
	return &Deque[T]{front: nil, back: nil}
}

func (d *Deque[T]) PushFront(value T) {
	newNode := &Node[T]{value: value, prev: nil, next: d.front}
	if d.front != nil {
		d.front.prev = newNode
	} else {
		d.back = newNode
	}
	d.front = newNode
}

func (d *Deque[T]) PushBack(value T) {
	newNode := &Node[T]{value: value, prev: d.back, next: nil}
	if d.back != nil {
		d.back.next = newNode

	} else {
		d.front = newNode
	}
	d.back = newNode
}
