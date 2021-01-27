package list

// Fold is the fundamental list iterator.
//
// If list == (e_1 e_2 ... e_n), then this method returns
//
//   f(f(f(f(init, e1), e2), ...), e_n)
//
// That is, it obeys the recursion
//
//   list.Fold(f, init)  == list.Cdr.Fold(f, f(init, list.Car))
//   Nil().Fold(f, init) == init
//
// The list argument must be finite.
//
// Examples:
//
//   func add(x, y interface{}) interface{} {
//     return x.(int) + y.(int)
//   }
//
//   list.Fold(add, 0)          // Add up the elements of list.
//
//   func cons(tail, head interface{}) interface{} {
//     return NewPair(head, tail)
//   }
//
//   list.Fold(cons, Nil())     // Reverse list.
//
//   // How many strings in list?
//   list.Fold(func(count, x interface{}) interface{} {
//     if _, ok := x.(string); ok {
//       return count.(int) + 1
//     }
//     return count.(int)
//   })
//
//   // Length of the longest string in list:
//   list.Fold(func(maxLen, x interface{}) interface{} {
//     if s, ok := x.(string); ok && len(s) > maxLen.(int) {
//       return len(s)
//     }
//     return maxLen
//   })
//
func (list *Pair) Fold(f func(intermediate, element interface{}) interface{}, init interface{}) (result interface{}) {
	result = init
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		result = f(result, pair.Car)
	}
	return
}

// Fold is the fundamental list iterator.
//
// If n list arguments are provided, then the f
// function must take n+1 parameters: the "seed" or fold state, which is initially init, and then
// one element from each list. The fold operation terminates when the shortest list runs out of values.
// At least one of the list arguments must be finite.
func Fold(f func(intermediate interface{}, elements ...interface{}) interface{}, init interface{}, lists ...*Pair) (result interface{}) {
	result = init
	if len(lists) == 0 {
		return
	}
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		result = f(result, a.args...)
	}
	return
}

// FoldRight is the fundamental list recursion operator.
//
// If list == (e_1 e_2 ... e_n), then this method returns
//
//   f(f(f(f(init, e_n), ...), e_2), e_1)
//
// That is, it obeys the recursion
//
//   list.FoldRight(f, init)  == f(list.Cdr.FoldRight(f, init), list.Car)
//   Nil().FoldRight(f, init) == init
//
// The list argument must be finite.
//
// Examples:
//
//   func cons(tail, head interface{}) interface{} {
//     return NewPair(head, tail)
//   }
//
//   list.FoldRight(cons, Nil())     // Copy list.
//
//   // Filter the even numbers out of list.
//   list.FoldRight(func(l, x interface{}) interface{} {
//     if x.(int)%2 == 0 {
//       return NewPair(x, l)
//     }
//     return l
//   }, Nil())
//
func (list *Pair) FoldRight(f func(intermediate, element interface{}) interface{}, init interface{}) (result interface{}) {
	var recur func(*Pair) interface{}
	recur = func(list *Pair) interface{} {
		if list == nil {
			return init
		}
		return f(recur(list.Cdr.(*Pair)), list.Car)
	}
	return recur(list)
}

// FoldRight is the fundamental list recursion operator.
//
// If n list arguments are provided, then the f
// function must take n+1 parameters: the "seed" or fold state, which is initially init, and then one
// element from each list. The fold operation terminates when the shortest list runs out of values.
// At least one of the list arguments must be finite.
func FoldRight(f func(intermediate interface{}, elements ...interface{}) interface{}, init interface{}, lists ...*Pair) (result interface{}) {
	var recur func(lists ...*Pair) interface{}
	recur = func(lists ...*Pair) interface{} {
		var cars []interface{}
		var cdrs []*Pair
		for _, l := range lists {
			if l == nil {
				return init
			}
			cars = append(cars, l.Car)
			cdrs = append(cdrs, l.Cdr.(*Pair))
		}
		return f(recur(cdrs...), cars...)
	}
	return recur(lists...)
}

// PairFold is analogous to Fold, but f is applied to successive sublists of the list, rather than
// successive elements -- that is, f is applied to the pairs making up the list, giving this recursion:
//
//   list.PairFold(f, init)  == { tail := list; list.Cdr.PairFold(f, f(init, tail)) }
//   Nil().PairFold(f, init) == init
//
// The list argument must be finite. The f function may reliably assign to the Cdr of each pair it is given without
// altering the sequence of execution.
//
// Example:
//
//   // Destructively reverse a list.
//   list.PairFold(func(tail interface{}, pair *Pair) interface{} {
//     pair.Cdr = tail
//     return pair
//   }, Nil())
//
func (list *Pair) PairFold(f func(intermediate interface{}, sublist *Pair) interface{}, init interface{}) (result interface{}) {
	result = init
	for pair := list; pair != nil; {
		cdr := pair.Cdr.(*Pair)
		result = f(result, pair)
		pair = cdr
	}
	return
}

// PairFold is analogous to Fold, but f is applied to successive sublists of the lists, rather than
// successive elements -- that is, f is applied to the pairs making up the lists, giving this recursion.
//
// For a finite list, the f function may reliably assign to the Cdr of each pair it is given without
// altering the sequence of execution.
//
// At least one of the list arguments must be finite.
func PairFold(f func(intermediate interface{}, sublists ...*Pair) interface{}, init interface{}, lists ...*Pair) (result interface{}) {
	result = init
	if len(lists) == 0 {
		return
	}
	for a, ok := initPairArgs(lists); ok; ok = a.next() {
		result = f(result, a.args...)
	}
	return
}

// PairFoldRight holds the same relationship with FoldRight that PairFold holds with Fold.
//
// PairFoldRight obeys the recursion:
//
//   list.PairFoldRight(f, init)  == f(list.Cdr.PairFoldRight(f, init), list)
//   Nil().PairFoldRight(f, init) == init
//
// Example:
//
//   func cons(tail interface{}, pair *Pair) interface{} {
//     return NewPair(pair, tail)
//   }
//
//   List(1, 2, 3).PairFoldRight(cons, Nil()) => ((1 2 3) (2 3) (3))
//
// The list argument must be finite.
func (list *Pair) PairFoldRight(f func(intermediate interface{}, sublist *Pair) interface{}, init interface{}) (result interface{}) {
	var recur func(*Pair) interface{}
	recur = func(list *Pair) interface{} {
		if list == nil {
			return init
		}
		return f(recur(list.Cdr.(*Pair)), list)
	}
	return recur(list)
}

// PairFoldRight holds the same relationship with FoldRight that PairFold holds with Fold.
//
// At least one of the list arguments must be finite.
func PairFoldRight(f func(intermediate interface{}, sublists ...*Pair) interface{}, init interface{}, lists ...*Pair) (result interface{}) {
	var recur func(...*Pair) interface{}
	recur = func(lists ...*Pair) interface{} {
		var cdrs []*Pair
		for _, l := range lists {
			if l == nil {
				return init
			}
			cdrs = append(cdrs, l.Cdr.(*Pair))
		}
		return f(recur(cdrs...), lists...)
	}
	return recur(lists...)
}

// Reduce is a variant of Fold.
//
// init should be a "right identity" of the function f -- that is, for any value x acceptable
// to f, f(init, x) == x.
//
// Reduce has the following definition:
//   If list == Nil(), return init;
//   Otherwise, return list.Cdr.Fold(f, list.Car)
// ...in other words, we compute list.Fold(f, init)
//
// Note that init is used only in the empy-list case. You typically use Reduce when applying f is
// expensive and you'd like to avoid the extra application incurred when Fold applies f to the head of
// list and the identity value, redundantly producing the same value passed in to f. For example, if f
// involves searching a file directory or performing a database query, this can be significant. In
// general, however, Fold is useful in many contexts where Reduce is not (consider the examples
// given in the Fold documentation -- only one of them uses a function with a right identity. The others
// may not be performed with Reduce).
//
//   func max(x, y interface{}) interface{} {
//     if x.(int) > y.(int) {
//       return x
//     }
//     return y
//   }
//
//   // Take the max of a list of non-negative integers.
//   nums.Reduce(max, 0)
//
func (list *Pair) Reduce(f func(intermediate, element interface{}) interface{}, init interface{}) (result interface{}) {
	if list == nil {
		return init
	}
	return list.Cdr.(*Pair).Fold(f, list.Car)
}

// ReduceRight is the fold-right variant of Reduce.
//
// It obeys the following definition:
//   Nil().ReduceRight(f, init) == init
//   List(e_1).ReduceRight(f, init) == f(init, e_1) == e_1
//   List(e_1, e_2, ...).ReduceRight(f, init) == f(List(e_2, ...).Reduce(f, init), e_1)
// ...in other words, we compute list.FoldRight(f, init)
//
//   func append(intermediate, element interface{}) interface{} {
//     return Append(intermediate.(*Pair), element.(*Pair))
//   }
//
//   listOfLists.ReduceRight(append, Nil())
//
func (list *Pair) ReduceRight(f func(intermediate, element interface{}) interface{}, init interface{}) (result interface{}) {
	if list == nil {
		return init
	}
	var recur func(*Pair) interface{}
	recur = func(list *Pair) interface{} {
		cdr := list.Cdr.(*Pair)
		if cdr == nil {
			return list.Car
		}
		return f(recur(cdr), list.Car)
	}
	return recur(list)
}

// Unfold is the fundamental recursive list constructor, just as FoldRight is the fundamental recursive list consumer.
//
// Unfold is best described by its basic recursion:
//
//   Unfold(predicate, element, nextSeed, seed, tailGen) ==
//     if predicate(seed) {
//       return tailGen(seed)
//     }
//     return NewPair(element(seed), Unfold(predicate, element, nextSeed, nextSeed(seed), tailGen))
//
//   predicate   Determines when to stop unfolding.
//   element     Maps each seed value to the corresponding list element.
//   nextSeed    Maps each seed value to the next seed value.
//   seed        The "state" value for the Unfold.
//   tailGen     Creates the tail of the list; defaults to func(_ interface{}) interface{} { return Nil() }
//
// In other words, we use nextSeed to generate a sequence of seed values
//
//   seed, nextSeed(seed), nextSeed^2(seed), nextSeed^3(seed), ...
//
// These seed values are mapped to list elements by element, producing the elements of the result list
// in a left-to-right order. predicate says when to stop.
//
// While Unfold may seem a bit abstract to novice functional programmers, it can be used in a number of ways:
//
//   // List of squares: 1^2, ..., 10^2
//   Unfold(func(x interface{}) bool {return x.(int) > 10},
//          func(x interface{}) interface{} {return x.(int) * x.(int)},
//          func(x interface{}) interface{} {return x.(int)+1},
//          1, nil)
//
//   // Copy a proper list.
//   Unfold(IsNilPair, Car, Cdr, list, nil)
//
//   // Copy a possibly non-proper list.
//   Unfold(IsEnd, Car, Cdr, list, nil)
//
//   // Append head onto tail.
//   Unfold(IsNilPair, Car, Cdr, head, func(_ interface{}) interface{} { return tail })
//
func Unfold(
	predicate func(interface{}) bool,
	element func(interface{}) interface{},
	nextSeed func(interface{}) interface{},
	seed interface{},
	tailGen func(interface{}) interface{}) (result interface{}) {
	if predicate(seed) {
		if tailGen == nil {
			return (*Pair)(nil)
		}
		return tailGen(seed)
	}
	res := &Pair{Car: element(seed)}
	last := res
	for {
		seed = nextSeed(seed)
		if predicate(seed) {
			if tailGen != nil {
				last.Cdr = tailGen(seed)
			} else {
				last.Cdr = (*Pair)(nil)
			}
			return res
		}
		last = last.ncdr(element(seed))
	}
}

// UnfoldRight is the fundamental iterative list constructor, just as Fold is the fundamental iterative list consumer.
//
// UnfoldRight constructs a list with the following loop:
//
//   func loop(seed, list interface{}) interface{} {
//     if predicate(seed) {
//       return list
//     }
//     return loop(nextSeed(seed), NewPair(element(seed), list))
//   }
//   loop(seed, tail)
//
//   predicate   Determines when to stop unfolding.
//   element     Maps each seed value to the corresponding list element.
//   nextSeed    Maps each seed value to the next seed value.
//   seed        The "state" value for the UnfoldRight.
//   tail        list terminator
//
// In other words, we use nextSeed to generate a sequence of seed values
//
//   seed, nextSeed(seed), nextSeed^2(seed), nextSeed^3(seed), ...
//
// These seed values are mapped to list elements by element, producing the elements of the result list in a
// right-to-left order. predicate says when to stop.
//
// While UnfoldRight may seem a bit abstract to novice functional programmers, it can be used in a number of ways:
//
//   // List of squares: 1^2, ..., 10^2
//   UnfoldRight(func (x interface{}) bool {return x.(int) == 0},
//               func (x interface{}) interface{} {return x.(int) * x.(int)},
//               func (x interface{}) interface{} {return x.(int) - 1},
//               10, Nil())
//
//   // Reverse a proper list.
//   UnfoldRight(IsNilPair, Car, Cdr, list, Nil())
//
//   // AppendReverse(revHead, tail)
//   UnfoldRight(IsNilPair, Car, Cdr, revHead, tail)
//
func UnfoldRight(
	predicate func(interface{}) bool,
	element func(interface{}) interface{},
	nextSeed func(interface{}) interface{},
	seed interface{},
	tail interface{}) (result interface{}) {
	if predicate(seed) {
		return tail
	}
	res := &Pair{Car: element(seed), Cdr: tail}
	for {
		seed = nextSeed(seed)
		if predicate(seed) {
			return res
		}
		res = &Pair{Car: element(seed), Cdr: res}
	}
}

// Map applies f element-wise to the elements of list and returns a list of the results, in order.
// Map is guaranteed to call f on the elements of the list in order from left to right.
// The list argument must be finite.
//
//   List(List(1, 2), List(3, 4), List(5, 6)).Map(Cadr) => (2 4 6)
//
//   List(1, 2, 3, 4, 5).Map(func (n interface{}) interface{} {return x.(int)+1}) => (2 3 4 5 6)
//
//   count := 0
//   List("a", "b").Map(func(_ interface{}) interface{} {
//      count++
//      return count
//   })                  => (1 2)
//
func (list *Pair) Map(f func(element interface{}) interface{}) (result *Pair) {
	if list == nil {
		return
	}
	result = &Pair{Car: f(list.Car)}
	last := result
	for pair := list.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
		last = last.ncdr(f(pair.Car))
	}
	last.Cdr = (*Pair)(nil)
	return
}

// Map applies f element-wise to the elements of the lists and returns a list of the results, in order.
// f is a function taking as many arguments as there are list arguments and returning a single value.
// Map is guaranteed to call f on the elements of the lists in order from left to right.
//
//   func sum(xs ...interface{}) interface{} {
//     result := 0
//     for x := range xs {
//       result += x.(int)
//     }
//     return result
//   }
//
//   Map(sum, List(1, 2, 3), List(4, 5, 6)) => (5 7 9)
//
// Map terminates when the shortest list runs out. At least one of the arguments must be finite:
//
//   Map(sum, List(3, 1, 4, 1), Circular(1, 0)) => (4 1 5 1)
//
func Map(f func(elements ...interface{}) interface{}, lists ...*Pair) (result *Pair) {
	if len(lists) == 0 {
		return
	}
	a, ok := initCarArgs(lists)
	if !ok {
		return
	}
	result = &Pair{Car: f(a.args...)}
	last := result
	for ok = a.next(); ok; ok = a.next() {
		last = last.ncdr(f(a.args...))
	}
	last.Cdr = (*Pair)(nil)
	return
}

// ForEach is like Map, but ForEach calls f for its side effects rather than for its values.
// ForEach is guaranteed to call f on the elements of the list in order from left to right.
// The list argument must be finite.
//
//   v := make([]int, 5)
//   List(0, 1, 2, 3, 4).ForEach(func (e interface{}) {
//     i := e.(int)
//     v[i] = i
//   })
//   v => [0, 1, 2, 3, 4]
//
func (list *Pair) ForEach(f func(element interface{})) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		f(pair.Car)
	}
}

// ForEach is like Map, but ForEach calls f for its side effects rather than for its values.
// ForEach is guaranteed to call f on the elements of the lists in order from left to right.
// At least one of the argument lists must be finite.
func ForEach(f func(elements ...interface{}), lists ...*Pair) {
	if len(lists) == 0 {
		return
	}
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		f(a.args...)
	}
}

// AppendMap maps f over the elements of the list, just as in Map. However, the results of the
// applications are appended together to make the final result. AppendMap uses Append to append
// the results together. The list argument must be finite.
func (list *Pair) AppendMap(f func(element interface{}) *Pair) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if res := f(pair.Car); res != nil {
			if pair = pair.Cdr.(*Pair); pair == nil {
				result = res
				return
			}
			var last *Pair
			result, last = copyList(res)
			for {
				if cdr := f(pair.Car); cdr != nil {
					if pair = pair.Cdr.(*Pair); pair == nil {
						last.Cdr = cdr
						return
					}
					last.Cdr, last = copyList(cdr)
				} else if pair = pair.Cdr.(*Pair); pair == nil {
					return
				}
			}
		}
	}
	return
}

// AppendMap maps f over the elements of the lists, just as in Map. However, the results of the
// applications are appended together to make the final result. AppendMap uses Append to append
// the results together. At least one of the list arguments must be finite.
func AppendMap(f func(elements ...interface{}) *Pair, lists ...*Pair) (result *Pair) {
	if len(lists) == 0 {
		return
	}
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		if res := f(a.args...); res != nil {
			if !a.next() {
				result = res
				return
			}
			var last *Pair
			result, last = copyList(res)
			for {
				if cdr := f(a.args...); cdr != nil {
					if !a.next() {
						last.Cdr = cdr
						return
					}
					last.Cdr, last = copyList(cdr)
				} else if !a.next() {
					return
				}
			}
		}
	}
	return
}

// NAppendMap maps f over the elements of the list, just as in Map. However, the results of the
// applications are appended together to make the final result. NAppendMap uses NAppend to append
// the results together. The list argument must be finite.
//
// Example:
//
//   List(1, 3, 8).NAppendMap(func(x interface{}) *Pair {
//     return List(x, -(x.(int)))
//   })                              => (1 -1 3 -3 8 -8)
//
func (list *Pair) NAppendMap(f func(element interface{}) *Pair) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if result = f(pair.Car); result != nil {
			if pair = pair.Cdr.(*Pair); pair == nil {
				return
			}
			last := result.LastPair()
			for {
				if cdr := f(pair.Car); cdr != nil {
					last.Cdr = cdr
					if pair = pair.Cdr.(*Pair); pair == nil {
						return
					}
					last = cdr.LastPair()
				} else if pair = pair.Cdr.(*Pair); pair == nil {
					return
				}
			}
		}
	}
	return
}

// NAppendMap maps f over the elements of the lists, just as in Map. However, the results of the
// applications are appended together to make the final result. NAppendMap uses NAppend to append
// the results together. At least one of the list arguments must be finite.
func NAppendMap(f func(elements ...interface{}) *Pair, lists ...*Pair) (result *Pair) {
	if len(lists) == 0 {
		return
	}
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		if result = f(a.args...); result != nil {
			if !a.next() {
				return
			}
			last := result.LastPair()
			for {
				if cdr := f(a.args...); cdr != nil {
					last.Cdr = cdr
					if !a.next() {
						return
					}
					last = cdr.LastPair()
				} else if !a.next() {
					return
				}
			}
		}
	}
	return
}

// NMap is the linear-update variant of Map.
func (list *Pair) NMap(f func(element interface{}) interface{}) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		pair.Car = f(pair.Car)
	}
	return list
}

// NMap is the linear-update variant of Map. NMap is allowed, but not required, to alter the cons cells of
// the first list to construct the result. The remaining lists must have at least as many elements as the
// first list.
func NMap(f func(elements ...interface{}) interface{}, lists ...*Pair) (result *Pair) {
	if len(lists) == 0 {
		return
	}
	pair := lists[0]
	if pair == nil {
		return
	}
	result = pair
	a, ok := initCarArgs(lists)
	for {
		if !ok {
			panic("not enough elements in remaining lists")
		}
		pair.Car = f(a.args...)
		if pair = a.cdrSlice[0]; pair == nil {
			return
		}
		ok = a.next()
	}
}

// PairForEach is like ForEach, but f is applied to successive sublists of the argument list. That is, f is applied to
// the cons cells of the list, rather than the list's elements. These applications occur in left-to-right order.
//
// The f function may reliably assign to the Cdr of the pairs it is given without altering the sequence of execution.
//
//   List(1, 2, 3).PairForEach(func (x interface{}) { fmt.Println(x) }) =>
//      (1 2 3)
//      (2 3)
//      (3)
//
// The list argument must be finite.
func (list *Pair) PairForEach(f func(sublist *Pair)) {
	for pair, cdr := list, (*Pair)(nil); pair != nil; pair = cdr {
		cdr = pair.Cdr.(*Pair)
		f(pair)
	}
}

// PairForEach is like ForEach, but f is applied to successive sublists of the argument lists. That is, f is applied to
// the cons cells of the lists, rather than the lists' elements. These applications occur in left-to-right order.
//
// The f function may reliably assign to the Cdr of the pairs it is given without altering the sequence of execution.
//
// At least one of the list arguments must be finite.
func PairForEach(f func(sublists ...*Pair), lists ...*Pair) {
	if len(lists) == 0 {
		return
	}
	for a, ok := initPairArgs(lists); ok; ok = a.next() {
		f(a.args...)
	}
}

// FilterMap is like Map, but only when f returns true as a second value, the first value is
// saved.
//
//   List("a", 1, "b", 3, "c", 7).FilterMap(func(x interface{}) (interface{}, bool) {
//     if n, ok := x.(int); ok {
//       return n*n, true
//     }
//     return nil, false
//   })                    => (1 9 49)
//
// The list argument must be finite.
func (list *Pair) FilterMap(f func(element interface{}) (interface{}, bool)) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if res, ok := f(pair.Car); ok {
			result = &Pair{Car: res}
			last := result
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if res, ok = f(pair.Car); ok {
					last = last.ncdr(res)
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// FilterMap is like Map, but only when f returns true as a second value, the first value is
// saved. At least one of the list arguments must be finite.
func FilterMap(f func(elements ...interface{}) (interface{}, bool), lists ...*Pair) (result *Pair) {
	if len(lists) == 0 {
		return
	}
	for a, aok := initCarArgs(lists); aok; aok = a.next() {
		if res, ok := f(a.args...); ok {
			result = &Pair{Car: res}
			last := result
			for aok = a.next(); aok; aok = a.next() {
				if res, ok = f(a.args...); ok {
					last = last.ncdr(res)
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// NFilterMap is the linear-update variant of FilterMap.
func (list *Pair) NFilterMap(f func(element interface{}) (interface{}, bool)) (result *Pair) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if res, ok := f(pair.Car); ok {
			result = list
			last := result
			last.Car = res
			for pair = pair.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
				if res, ok = f(pair.Car); ok {
					last = last.Cdr.(*Pair)
					last.Car = res
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}

// NFilterMap is the linear-update variant of FilterMap.
func NFilterMap(f func(elements ...interface{}) (interface{}, bool), lists ...*Pair) (result *Pair) {
	if len(lists) == 0 {
		return
	}
	for a, aok := initCarArgs(lists); aok; aok = a.next() {
		if res, ok := f(a.args...); ok {
			result = lists[0]
			last := result
			last.Car = res
			for aok = a.next(); aok; aok = a.next() {
				if res, ok = f(a.args...); ok {
					last = last.Cdr.(*Pair)
					last.Car = res
				}
			}
			last.Cdr = (*Pair)(nil)
			return
		}
	}
	return
}
