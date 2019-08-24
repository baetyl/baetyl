package router

import "container/list"

type deque struct {
	elements *list.List
}

func newDeque() *deque {
	return &deque{elements: list.New()}
}

// Push add an element in the last
func (d *deque) Push(v interface{}) {
	d.elements.PushBack(v)
}

// Pop remove the last element and return it's value
func (d *deque) Pop() interface{} {
	e := d.elements.Back()
	if e != nil {
		d.elements.Remove(e)
		return e.Value
	}
	return nil
}

// Peek return the value of the last element
func (d *deque) Peek() interface{} {
	e := d.elements.Back()
	if e != nil {
		return e.Value
	}
	return nil
}

// Offer add an element in the last
func (d *deque) Offer(v interface{}) {
	d.elements.PushBack(v)
}

// Poll remove the first element and return it's value
func (d *deque) Poll() interface{} {
	e := d.elements.Front()
	if e != nil {
		d.elements.Remove(e)
		return e.Value
	}
	return nil
}

// Len return the deque's length
func (d *deque) Len() int {
	return d.elements.Len()
}

// Empty validate the deque is empty or not
func (d *deque) Empty() bool {
	return d.elements.Len() == 0
}
