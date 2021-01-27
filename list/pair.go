package list

import (
	"bytes"
	"fmt"
)

// Pair is the core tuple type from which list- and tree-like data structures can be created.
type Pair struct {
	Car, Cdr interface{}
}

// Nil returns (*Pair)(nil), but should be easier on the eyes.
func Nil() *Pair {
	return nil
}

func (list *Pair) String() string {
	if list == nil {
		return "()"
	}
	var buf bytes.Buffer
	fmt.Fprint(&buf, "(", list.Car)
	for {
		nextPair, ok := list.Cdr.(*Pair)
		if !ok {
			fmt.Fprint(&buf, " . ", list.Cdr)
			break
		}
		if nextPair == nil {
			break
		}
		fmt.Fprint(&buf, " ", nextPair.Car)
		list = nextPair
	}
	fmt.Fprint(&buf, ")")
	return buf.String()
}

// NewPair returns &Pair{Car: car, Cdr: cdr}
func NewPair(car, cdr interface{}) *Pair {
	return &Pair{Car: car, Cdr: cdr}
}

func list(elements ...interface{}) (result *Pair, last *Pair) {
	if len(elements) == 0 {
		return
	}
	result = &Pair{Car: elements[0]}
	last = result
	for _, e := range elements[1:] {
		last = last.ncdr(e)
	}
	last.Cdr = (*Pair)(nil)
	return
}

// List returns a newly allocated list of its arguments.
func List(elements ...interface{}) (result *Pair) {
	result, _ = list(elements...)
	return
}

// Cons is like List, but the last argument provides the tail of the constructed list.
//
//   Cons(1, 2, 3, 4) => (1 2 3 . 4)
//
func Cons(element1, element2 interface{}, moreElements ...interface{}) (result *Pair) {
	if len(moreElements) == 0 {
		return &Pair{Car: element1, Cdr: element2}
	}
	last := &Pair{Car: element2}
	result = &Pair{Car: element1, Cdr: last}
	sz := len(moreElements) - 1
	for i := 0; i < sz; i++ {
		last = last.ncdr(moreElements[i])
	}
	last.Cdr = moreElements[len(moreElements)-1]
	return
}

// NewList returns a list of the given length, whose elements are all the value element.
func NewList(length int, element interface{}) (result *Pair) {
	if length < 0 {
		panic(negativeLength(length))
	}
	for i := 0; i < length; i++ {
		result = &Pair{Car: element, Cdr: result}
	}
	return
}

// Tabulate returns a list of the given length. Element i of the list, where 0 <= i < lenght,
// is produced by init(i). No guarantee is made about the dynamic order in which init is applied
// to these indices.
//
//   Tabulate(4, func(x interface{}) interface{} {return x.(int)+1}) => (1 2 3 4)
//
func Tabulate(length int, init func(int) interface{}) (result *Pair) {
	if length < 0 {
		panic(negativeLength(length))
	}
	if length == 0 {
		return
	}
	result = &Pair{Car: init(0)}
	last := result
	for i := 1; i < length; i++ {
		last = last.ncdr(init(i))
	}
	last.Cdr = (*Pair)(nil)
	return
}

func copyList(list *Pair) (result *Pair, last *Pair) {
	if list == nil {
		return
	}
	result = &Pair{Car: list.Car}
	last = result
	for {
		pair, _ := list.Cdr.(*Pair)
		if pair == nil {
			last.Cdr = list.Cdr
			return
		}
		last = last.ncdr(pair.Car)
		list = pair
	}
}

// Copy copies the spine of the argument.
func (list *Pair) Copy() (result *Pair) {
	result, _ = copyList(list)
	return
}

// Circular constructs a circular list of the elements.
//
//   Circular(1, 2) => (1 2 1 2 1 2 ...)
//
func Circular(element interface{}, moreElements ...interface{}) (result *Pair) {
	if len(moreElements) == 0 {
		result = &Pair{Car: element}
		result.Cdr = result
		return
	}
	more, last := list(moreElements...)
	result = &Pair{Car: element, Cdr: more}
	last.Cdr = result
	return
}
