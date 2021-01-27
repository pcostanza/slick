package list

// Filter returns all the elements of list that satisfy the predicate. The list
// is not disordered -- elements that appear in the result list occur in the same
// order as they occur in the argument list. The returned list may share a common
// tail with the argument list. The dynamic order in which the various applications
// of predicate are made is not specified.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   list.List(0, 7, 8, 8, 43, -4).Filter(even) => (0 8 8 -4)
//
func (list *Pair) Filter(predicate func(x interface{}) bool) (result *Pair) {
	// does not share longest tail
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		car := pair.Car
		if predicate(car) {
			result = &Pair{Car: car}
			last := result
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				car = pair.Car
				if predicate(car) {
					last = last.ncdr(car)
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// Partition partitions the elements of list with predicate pred, and returns two
// values: the list of in-elements and the list of out-elements. The lists are not
// disordered -- elements occur in the result lists in the same order as they
// occur in the argument list. The dynamic order in which the various applications
// of predicate are made is not specified. One of the returned lists may share
// a common tail with the argument list.
//
//   func isString(x interface{}) bool {
//     _, ok := x.(string)
//     return ok
//   }
//
//   list.List("one", 2, 3, "four", "five", 6).Partition(isString) =>
//      ("one" "four" "five")
//      (2 3 6)
//
func (list *Pair) Partition(predicate func(x interface{}) bool) (in, out *Pair) {
	var lastIn, lastOut *Pair
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		car := pair.Car
		if predicate(car) {
			if in == nil {
				in = &Pair{Car: car}
				lastIn = in
			} else {
				lastIn = lastIn.ncdr(car)
			}
		} else {
			if out == nil {
				out = &Pair{Car: car}
				lastOut = out
			} else {
				lastOut = lastOut.ncdr(car)
			}
		}
	}
	if lastIn != nil {
		lastIn.Cdr = (*Pair)(nil)
	}
	if lastOut != nil {
		lastOut.Cdr = (*Pair)(nil)
	}
	return
}

// Remove returns list without the elements that satisfy predicate:
//
//   func (list *Pair) Remove(predicate func(x interface{}) bool) *Pair {
//     return list.Filter(func (x interface{}) bool {return !predicate(x)})
//   }
//
// The list is not disordered -- elements that appear in the result list occur
// in the same order as they occur in the argument list. The returned list may
// share a common tail with the argument list. The dynamic order in which the
// various applications of predicate are made is not specified.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   list.List(0, 7, 8, 8, 43, -4).Remove(even) => (7 43)
//
func (list *Pair) Remove(predicate func(x interface{}) bool) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		car := pair.Car
		if !predicate(car) {
			result = &Pair{Car: car}
			last := result
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				car = pair.Car
				if !predicate(car) {
					last = last.ncdr(car)
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// NFilter is the linear-update variant of Filter.
func (list *Pair) NFilter(predicate func(x interface{}) bool) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if predicate(pair.Car) {
			result = pair
			for next := pair.Cdr.(*Pair); next != nil; next = next.Cdr.(*Pair) {
				if predicate(next.Car) {
					pair.Cdr = next
					pair = next
				}
			}
			pair.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// NPartition is the linear-update variant of Partition.
func (list *Pair) NPartition(predicate func(x interface{}) bool) (in, out *Pair) {
	var lastIn, lastOut *Pair
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if predicate(pair.Car) {
			if in == nil {
				in = pair
				lastIn = in
			} else {
				lastIn.Cdr = pair
				lastIn = pair
			}
		} else {
			if out == nil {
				out = pair
				lastOut = out
			} else {
				lastOut.Cdr = pair
				lastOut = pair
			}
		}
	}
	if lastIn != nil {
		lastIn.Cdr = (*Pair)(nil)
	}
	if lastOut != nil {
		lastOut.Cdr = (*Pair)(nil)
	}
	return
}

// NRemove is the linear-update variant of Remove.
func (list *Pair) NRemove(predicate func(x interface{}) bool) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if !predicate(pair.Car) {
			result = pair
			for next := pair.Cdr.(*Pair); next != nil; next = next.Cdr.(*Pair) {
				if !predicate(next.Car) {
					pair.Cdr = next
					pair = next
				}
			}
			pair.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}
