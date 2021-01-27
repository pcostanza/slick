package list

func lset2le(list1, list2 *Pair) bool {
	return list1.Every(func(x interface{}) bool {
		return list2.Member(x) != nil
	})
}

// SetLessThanEqual returns true iff every list_i is a subset of list_i+1, using ==
// to compare elements.
//
// List A is a subset of list B if every element in A is equal to some element of B.
//
//   SetLessThanEqual(List(1), List(1, 2, 1), List(1, 2, 3, 3)) => true
//
//   // Trivial cases
//   SetLessThanEqual() => true
//   SetLessThanEqual(List(1)) => true
//
func SetLessThanEqual(lists ...*Pair) bool {
	if len(lists) < 2 {
		return true
	}
	for index, s1 := range lists[:len(lists)-1] {
		s2 := lists[index+1]
		if s1 != s2 && !lset2le(s1, s2) {
			return false
		}
	}
	return true
}

// SetEqual returns true iff every list_i is set-equal to list_i+1, using ==
// to compare elements.
//
// Set-equal simply means that list_i is a subset of list_i+1, and list_i+1 is a subset of list_i.
//
//   SetEqual(List("b", "e", "a"), List("a", "e", "b"), List("e", "e", "b", "a")) => true
//
//   // Trivial cases
//   SetEqual() => true
//   SetEqual(List(1)) => true
//
func SetEqual(lists ...*Pair) bool {
	if len(lists) < 2 {
		return true
	}
	for index, s1 := range lists[:len(lists)-1] {
		s2 := lists[index+1]
		if s1 != s2 && !(lset2le(s1, s2) && lset2le(s2, s1)) {
			return false
		}
	}
	return true
}

// Adjoin adds the elements not already in the list parameter to the result list. The result
// shares a common tail with the list parameter. The new elements are added to the front of
// the list, but no guarantees are made about their order. Elements are compared using ==.
//
// The list parameter is always a suffix of the result -- even if the list parameter contains
// repeated elements, these are not reduced.
//
//   Adjoin(List("a", "b", "c", "d", "c", "e"), "a", "e", "i", "o", "u")
//    => ("u" "o" "i" "a" "b" "c" "d" "c" "e")
//
func (list *Pair) Adjoin(elements ...interface{}) *Pair {
	for _, element := range elements {
		if list.Member(element) == nil {
			list = &Pair{Car: element, Cdr: list}
		}
	}
	return list
}

// SetUnion returns the union of the lists, using == to compare elements.
//
// The union of lists A and B is constructed as follows:
//
// * If A is the empty list, the answer is B (or a copy of B).
//
// * Otherwise, the result is initialised to be list A (or a copy of A).
//
// * Proceed through the elements of list B in a left-to-right order. If b is such an element of B,
// compare every element r of the current result list to b: r == b. If all comparisons fail, b is
// consed onto the front of the result.
//
// In the n-ary case, the two-argument list-union operation is simply folded across the argument lists.
//
//   SeUnion(List("a", "b", "c", "d", "e"), List("a", "e", "i", "o", "u"))
//    => ("u" "o" "i" "a" "b" "c" "d" "e")
//
//   // Repeated elements in the first list are preserved.
//   SetUnion(List("a", "a", "c"), List("x", "a", "x")) => ("x" "a" "a" "c")
//
//   // Trivial cases
//   SetUnion() => ()
//   SetUnion(List("a", "b", "c")) => ("a", "b", "c")
//
func SetUnion(lists ...*Pair) *Pair {
	return Tabulate(len(lists), func(i int) interface{} {
		return lists[i]
	}).Reduce(func(temp, list interface{}) interface{} {
		t := temp.(*Pair)
		l := list.(*Pair)
		if l == nil {
			return t
		}
		if t == nil {
			return l
		}
		if l == t {
			return t
		}
		return l.Fold(func(temp, element interface{}) interface{} {
			if temp.(*Pair).Any(func(x interface{}) bool { return x == element }) {
				return temp
			}
			return NewPair(element, temp)
		}, t)
	}, Nil()).(*Pair)
}

// NSetUnion is the linear-update variant of SetUnion.
func NSetUnion(lists ...*Pair) *Pair {
	return Tabulate(len(lists), func(i int) interface{} {
		return lists[i]
	}).Reduce(func(temp, list interface{}) interface{} {
		t := temp.(*Pair)
		l := list.(*Pair)
		if l == nil {
			return t
		}
		if t == nil {
			return l
		}
		if l == t {
			return t
		}
		return l.PairFold(func(temp interface{}, pair *Pair) interface{} {
			element := pair.Car
			if temp.(*Pair).Any(func(x interface{}) bool { return x == element }) {
				return temp
			}
			pair.Cdr = temp
			return pair
		}, t).(*Pair)
	}, Nil()).(*Pair)
}

// SetIntersection returns the intersection of the lists, using == to compare elements.
//
// The intersection of lists A and B is comprised of every element of A that is equal to some element of B:
// a == b, for a in A, and b in B. Note this implies that an element which appears in B and multiple
// times in list A will also appear multiple times in the result.
//
// The order in which elements appear in the result is the same as they appear in the first list -- that is,
// SetIntersection essentially filters the first list, without disarranging element order. The result may
// share a common tail with the first list.
//
// In the n-ary case, the two-argument list-intersection operation is simply folded across the argument
// lists.
//
//   SetIntersection(List("a", "b", "c", "d", "e"), List("a", "e", "i", "o", "u")) => ("a" "e")
//
//   // Repeated elements in the first list are preserved.
//   SetIntersection(List("a", "x", "y", "a"), List("x", "a", "x", "z")) => ("a" "x" "a")
//
//   SetIntersection(List("a", "b", "c")) => ("a" "b" "c")  // Trivial case
//
func SetIntersection(list *Pair, moreLists ...*Pair) *Pair {
	lists := NAppendTabulate(len(moreLists), func(i int) *Pair {
		l := moreLists[i]
		if l == list {
			return nil
		}
		return &Pair{Car: l, Cdr: Nil()}
	})
	if lists.Any(IsNilPair) {
		return nil
	}
	if lists == nil {
		return list
	}
	return list.Filter(func(x interface{}) bool {
		return lists.Every(func(list interface{}) bool {
			return list.(*Pair).Member(x) != nil
		})
	})
}

// NSetIntersection is the linear-update variant of SetIntersection. It is allowed, but not required,
// to use the cons cells in its first list parameter to construct its answer.
func NSetIntersection(list *Pair, moreLists ...*Pair) *Pair {
	lists := NAppendTabulate(len(moreLists), func(i int) *Pair {
		l := moreLists[i]
		if l == list {
			return nil
		}
		return &Pair{Car: l, Cdr: Nil()}
	})
	if lists.Any(IsNilPair) {
		return nil
	}
	if lists == nil {
		return list
	}
	return list.NFilter(func(x interface{}) bool {
		return lists.Every(func(list interface{}) bool {
			return list.(*Pair).Member(x) != nil
		})
	})
}

// SetDifference returns the difference of the lists, using == for comparing elements.
// That is, it returns all the elements of the first list that are not equal to any element
// from one of the other list parameters.
//
// Elements that are repeated multiple times in the first list will occur multiple
// times in the result. The order in which elements appear in the result is the same as they appear in
// the first list -- that is, SetDifference essentially filters the first list, without disarranging element order. The
// result may share a common tail with the first list.
//
//   SetDifference(List("a", "b", "c", "d", "e"), List("a", "e", "i", "o", "u")) => ("b" "c" "d")
//
//   SetDifference(List("a", "b", "c")) => ("a" "b" "c")  // Trivial case
//
func SetDifference(list *Pair, moreLists ...*Pair) *Pair {
	lists := NAppendTabulate(len(moreLists), func(i int) *Pair {
		l := moreLists[i]
		if l == nil {
			return nil
		}
		return &Pair{Car: l, Cdr: Nil()}
	})
	if lists == nil {
		return list
	}
	if lists.Member(list) != nil {
		return nil
	}
	return list.Filter(func(x interface{}) bool {
		return lists.Every(func(list interface{}) bool {
			return list.(*Pair).Member(x) == nil
		})
	})
}

// NSetDifference is the linear-update variant of SetDifference. It is allowed, but not required,
// to use the cons cells in its first list parameter to construct its answer.
func NSetDifference(list *Pair, moreLists ...*Pair) *Pair {
	lists := NAppendTabulate(len(moreLists), func(i int) *Pair {
		l := moreLists[i]
		if l == nil {
			return nil
		}
		return &Pair{Car: l, Cdr: Nil()}
	})
	if lists == nil {
		return list
	}
	if lists.Member(list) != nil {
		return nil
	}
	return list.NFilter(func(x interface{}) bool {
		return lists.Every(func(list interface{}) bool {
			return list.(*Pair).Member(x) == nil
		})
	})
}

// SetXor returns the exclusive-or of the sets, using == to compare elements.
// If there are exactly two lists, this is all the elements
// that appear in exactly one of the two lists. The operation is associative, and thus extends to the
// n-ary case -- the elements that appear in an odd number of the lists. The result may share a common
// tail with any of the list parameters.
//
// More precisely, for two lists A and B, A xor B is a list of
//
// * every element a of A such that there is no element b of B such that a == b, and
//
// * every element b of B such that there is no elmente a of A such that b == a.
//
// In the n-ary case, the binary-xor operation is simply folded across the lists.
//
//   SetXor(List("a", "b", "c", "d", "e"), List("a", "e", "i", "o", "u")) => ("d" "c" "b" "i" "o" "u")
//
//   // Trivial cases
//   SetXor() => ()
//   SetXor(List("a", "b", "c")) => ("a", "b", "c")
//
func SetXor(lists ...*Pair) *Pair {
	return Tabulate(len(lists), func(i int) interface{} {
		return lists[i]
	}).Reduce(func(ai, bi interface{}) interface{} {
		a, b := ai.(*Pair), bi.(*Pair)
		ab, aintb := SetDifferenceAndIntersection(a, b)
		if ab == nil {
			return SetDifference(b, a)
		}
		if aintb == nil {
			return Append(b, a)
		}
		return b.Fold(func(tmp, xb interface{}) interface{} {
			if aintb.Member(xb) != nil {
				return tmp
			}
			return NewPair(xb, tmp)
		}, ab)
	}, Nil()).(*Pair)
}

// NSetXor is the linear-update variant of SetXor. It is allowed, but not required,
// to use the cons cells in its first list parameter to construct its answer.
func NSetXor(lists ...*Pair) *Pair {
	return Tabulate(len(lists), func(i int) interface{} {
		return lists[i]
	}).Reduce(func(ai, bi interface{}) interface{} {
		a, b := ai.(*Pair), bi.(*Pair)
		ab, aintb := NSetDifferenceAndIntersection(a, b)
		if ab == nil {
			return NSetDifference(b, a)
		}
		if aintb == nil {
			return NAppend(b, a)
		}
		return b.PairFold(func(tmp interface{}, bpair *Pair) interface{} {
			if aintb.Member(bpair.Car) != nil {
				return tmp
			}
			bpair.Cdr = tmp
			return bpair
		}, ab)
	}, Nil()).(*Pair)
}

// SetDifferenceAndIntersection returns two values -- the difference (as if by SetDifference) and
// the intersection (as if by SetIntersection) of the lists. It can be implemented more efficiently
// than calling SetDifference and SetIntersection separately.
//
// Either of the answer lists may share a common tail with the first list. This operation essentially
// partitions the first list.
func SetDifferenceAndIntersection(list *Pair, moreLists ...*Pair) (difference, intersection *Pair) {
	everyNil := true
	for _, l := range moreLists {
		if l != nil {
			everyNil = false
			break
		}
	}
	if everyNil {
		return list, nil
	}
	for _, l := range moreLists {
		if list == l {
			return nil, list
		}
	}
	lists := Tabulate(len(moreLists), func(i int) interface{} { return moreLists[i] })
	return list.Partition(func(element interface{}) bool {
		return !lists.Any(func(list interface{}) bool {
			return list.(*Pair).Member(element) != nil
		})
	})
}

// NSetDifferenceAndIntersection is the linear-update variant of SetDifferenceAndIntersection. It is allowed, but not required,
// to use the cons cells in its first list parameter to construct its answer.
func NSetDifferenceAndIntersection(list *Pair, moreLists ...*Pair) (difference, intersection *Pair) {
	everyNil := true
	for _, l := range moreLists {
		if l != nil {
			everyNil = false
			break
		}
	}
	if everyNil {
		return list, nil
	}
	for _, l := range moreLists {
		if list == l {
			return nil, list
		}
	}
	lists := Tabulate(len(moreLists), func(i int) interface{} { return moreLists[i] })
	return list.NPartition(func(element interface{}) bool {
		return !lists.Any(func(list interface{}) bool {
			return list.(*Pair).Member(element) != nil
		})
	})
}
