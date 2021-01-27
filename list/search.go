package list

// Find returns the first element of list that satisfies predicate.
// It returns a second value of true if such an element is found, and false otherwise.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(3, 1, 4, 1, 5, 9).Find(even) => 4, true
//
func (list *Pair) Find(predicate func(interface{}) bool) (result interface{}, ok bool) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		car := pair.Car
		if predicate(car) {
			return car, true
		}
	}
	return nil, false
}

// FindTail returns the first pair whose Car satisfies predicate. If no pair does, return Nil().
//
// FindTail can be viewed as a general-predicate variant of the Member function.
//
// Examples:
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(3, 1, 37, -8, -5, 0, 0).FindTail(even) => (-8 -5 0 0)
//   List(3, 1, 37, -5).FindTail(even) => ()
//
// In the circular-list case, this function "rotates" the list.
//
// FindTail is essentially DropWhile, where the sense of the predicate is inverted: FindTail
// searches until it finds an element satisfying the predicate; DropWhile searches until it finds an
// element that doesn't satisfy the predicate.
func (list *Pair) FindTail(predicate func(interface{}) bool) (result *Pair) {
	for result = list; result != nil; result = result.Cdr.(*Pair) {
		if predicate(result.Car) {
			return
		}
	}
	return
}

// TakeWhile returns the longest initial prefix of list whose elements all satisfy predicate.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(2, 18, 3, 10, 22, 9).TakeWhile(even) => (2 18)
//
func (list *Pair) TakeWhile(predicate func(interface{}) bool) (result *Pair) {
	result, _ = list.Span(predicate)
	return
}

// NTakeWhile is the linear-update variant of TakeWhile.
func (list *Pair) NTakeWhile(predicate func(interface{}) bool) (result *Pair) {
	result, _ = list.NSpan(predicate)
	return
}

// DropWhile drops the longest initial prefix of list whose elements all satisfy predicate,
// and returns the rest of the list.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(2, 18, 3, 10, 10, 22, 9).DropWhile(even) => (3 10 22 9)
//
// The circular-list case may be viewed as "rotating" the list.
func (list *Pair) DropWhile(predicate func(interface{}) bool) (result *Pair) {
	for result = list; result != nil; result = result.Cdr.(*Pair) {
		if !predicate(result.Car) {
			return
		}
	}
	return
}

// Span splits the list into the longest initial prefix whose elements all satisfy predicate,
// and the remaining tail.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(2, 18, 3, 10, 22, 9).Span(even) =>
//     (2 18)
//     (3 10 22 9)
//
func (list *Pair) Span(predicate func(interface{}) bool) (prefix *Pair, suffix interface{}) {
	if pair := list; pair != nil {
		if car := pair.Car; predicate(car) {
			prefix = &Pair{Car: car}
			last := prefix
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if car = pair.Car; predicate(car) {
					last = last.ncdr(car)
				} else {
					break
				}
			}
			last.Cdr = (*Pair)(nil)
			suffix = pair
			return
		}
	}
	suffix = list
	return
}

// NSpan is the linear-update variant of Span.
func (list *Pair) NSpan(predicate func(interface{}) bool) (prefix *Pair, suffix interface{}) {
	if pair := list; pair != nil {
		if predicate(pair.Car) {
			prefix = pair
			last := prefix
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if predicate(pair.Car) {
					last = pair
				} else {
					break
				}
			}
			last.Cdr = (*Pair)(nil)
			suffix = pair
			return
		}
	}
	suffix = list
	return
}

// Break splits the list into the longest initial prefix whose elements all do not satisfy predicate,
// and the remaining tail.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(3, 1, 4, 1, 5, 9).Break(even) =>
//     (3 1)
//     (4 1 5 9)
//
func (list *Pair) Break(predicate func(interface{}) bool) (prefix *Pair, suffix interface{}) {
	if pair := list; pair != nil {
		if car := pair.Car; !predicate(car) {
			prefix = &Pair{Car: car}
			last := prefix
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if car = pair.Car; !predicate(car) {
					last = last.ncdr(car)
				} else {
					break
				}
			}
			last.Cdr = (*Pair)(nil)
			suffix = pair
			return
		}
	}
	suffix = list
	return
}

// NBreak is the linear-update variant of Break.
func (list *Pair) NBreak(predicate func(interface{}) bool) (prefix *Pair, suffix interface{}) {
	if pair := list; pair != nil {
		if !predicate(pair.Car) {
			prefix = pair
			last := prefix
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if !predicate(pair.Car) {
					last = pair
				} else {
					break
				}
			}
			last.Cdr = (*Pair)(nil)
			suffix = pair
			return
		}
	}
	suffix = list
	return
}

// Any applies the predicate across the list, returning true if the predicate returns true on any application.
//
//   func isInteger(x interface{}) bool {
//     _, ok := x.(int)
//     return ok
//   }
//
//   List("a", 3, "b", 2.7).Any(isInteger)   => true
//   List("a", 3.1, "b", 2.7).Any(isInteger) => false
//
func (list *Pair) Any(predicate func(interface{}) bool) bool {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if predicate(pair.Car) {
			return true
		}
	}
	return false
}

// Any applies the predicate across the lists, returning true if the predicate returns true on any application.
//
// If there are n list arguments, then predicate must be a function taking n arguments and returning a single
// boolean value.
//
// Any applies predicate to the first elements of the list parameters. If this application returns true,
// Any immediately returns true. Otherwise, it iterates, applying predicate to the second elements of the
// list parameters, then the third, and so forth. The iteration stops when a true value is produced or one
// of the lists runs out of values; in the latter case, Any returns false.
//
//   func lessThan(xs ...interface{}) bool {
//     return xs[0].(int) < xs[1].(int)
//   }
//
//   Any(lessThan, List(3, 1, 4, 1, 5), List(2, 7, 1, 8, 2)) => true
//
func Any(predicate func(...interface{}) bool, lists ...*Pair) bool {
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		if predicate(a.args...) {
			return true
		}
	}
	return false
}

// Every applies the predicate across the list, returning true if the predicate returns true on every application.
func (list *Pair) Every(predicate func(interface{}) bool) bool {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if !predicate(pair.Car) {
			return false
		}
	}
	return true
}

// Every applies the predicate across the lists, returning true if the predicate returns true on every application.
//
// If there are n list arguments, then the predicate must be a function taking n arguments and returning a single
// boolean value.
//
// Every applies predicate to the first elements of the list parameters. If this application returns false,
// Every immediately returns fales. Otherwise, it iterates, applying predicate to the second elements of the
// list parameters, then the third, and so forth. The iteration stops when a false values is produced or one
// of the lists runs out of values; in the latter case, Every returns true.
func Every(predicate func(...interface{}) bool, lists ...*Pair) bool {
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		if !predicate(a.args...) {
			return false
		}
	}
	return true
}

// Index returns the index of the leftmost element that satisfies predicate.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(3, 1, 4, 1, 5, 9).Index(even) => 2
//
func (list *Pair) Index(predicate func(interface{}) bool) (result int) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if predicate(pair.Car) {
			return
		}
		result++
	}
	result = -1
	return
}

// Index returns the index of the leftmost elements that satisfy predicate.
//
// If there are n list arguments, then predicate must be a function taking n arguments and
// returning a single boolean value.
//
// Index applies predicate to the first elements of the list parameters. If this application returns
// true, Index immediately returns zero. Otherwise, it iterates, applying predicate to the second
// elements of the list parameters, then the third, and so forth. When it finds a tuple of list elements
// that cause predicate to return true, it stops and returns the zero-based index of that position in the lists.
//
// The iteration stops when one of the lists runs out of values; in this case, Index returns -1.
//
//   func lessThan(xs ...interface{}) bool {
//     return xs[0].(int) < xs[1].(int)
//   }
//
//   Index(lessThan, List(3, 1, 4, 1, 5, 9, 2, 5, 6), List(2, 7, 1, 8, 2)) => 1
//
//   func equal(xs ...interface{}) bool {
//     return xs[0].(int) == xs[1].(int)
//   }
//
//   Index(equal, List(3, 1, 4, 1, 5, 9, 2, 5, 6), List(2, 7, 1, 8, 2))    => -1
//
func Index(predicate func(...interface{}) bool, lists ...*Pair) (result int) {
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		if predicate(a.args...) {
			return
		}
		result++
	}
	result = -1
	return
}

// Member returns the first sublist of list whose Car is x, where the sublists of list
// ar the non-empty lists returned by list.Drop(i) for i less than the length of list.
// If x does not occur in list, then nil is returned. Member uses == to compare x with
// the elements of the list.
//
//   List(1, 2, 3).Member(1) => (1 2 3)
//   List(1, 2, 3).Member(2) => (2 3)
//   List(2, 3, 4).Member(1) => ()
//
// Note that fully general list searching may be performed with the Find and FindTail
// functions.
func (list *Pair) Member(x interface{}) (result *Pair) {
	for result = list; result != nil; result = result.Cdr.(*Pair) {
		if result.Car == x {
			return
		}
	}
	return
}
