package list

// Delete finds all elements of list that are equal (==) to x, and deletes them from the list.
//
// The list is not disordered -- elements that appear in the result list occur in the same
// order as they occur in the argument list. The result may share a common tail with the
// argument list.
//
// Note that fully general element deletion can be performed with the Remove and NRemove
// methods, for example:
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   // Delete all the even elements from list:
//   list.Remove(even)
//
func (list *Pair) Delete(x interface{}) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if car := pair.Car; car != x {
			result = &Pair{Car: car}
			last := result
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if car = pair.Car; car != x {
					last = last.ncdr(car)
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// NDelete is the linear-update variant of Delete.
func (list *Pair) NDelete(x interface{}) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if car := pair.Car; car != x {
			result = pair
			last := result
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if car = pair.Car; car != x {
					last.Cdr = pair
					last = pair
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// DeleteDuplicates removes duplicate elements from the list argument. If there are
// multiple equal (==) elements in the argument list, the result list contains the first
// or leftmost of these elements in the result. The order of these surviving elements
// is the same as in the original list -- DeleteDuplicates does not disorder the list
// (hence it is useful for "cleaning up" assocation lists).
//
// The result of DeleteDuplicates may share common tails between argument and result
// lists -- for example, if the list argument contains only unique elements, it may
// simply return exactly this list.
//
// Be aware that, in general, DeleteDuplicates runs in time O(n^2) for n-element lists.
// Uniquifying long lists can be accomplished in O(n lg n) time by sorting the list to
// bring equal elements together, then using a linear-time algorithm to remove equal
// elements. Alternatively, one can use algorithms based on element-marking, with
// linear-time results.
func (list *Pair) DeleteDuplicates() (result *Pair) {
	var recur func(*Pair) *Pair
	recur = func(list *Pair) *Pair {
		if list == nil {
			return nil
		}
		car, cdr := list.Car, list.Cdr.(*Pair)
		newTail := recur(cdr.Delete(car))
		if cdr == newTail {
			return list
		}
		return &Pair{Car: car, Cdr: newTail}
	}
	return recur(list)
}

// NDeleteDuplicates is the linear-update variant of DeleteDuplicates.
func (list *Pair) NDeleteDuplicates() (result *Pair) {
	result = list
	for pair := list; pair != nil; {
		cdr := pair.Cdr.(*Pair).NDelete(pair.Car)
		pair.Cdr = cdr
		pair = cdr
	}
	return
}
