package list

import (
	"reflect"
)

// Miscellaneous

// AppendToSlice uses Go's reflect package to append each element of the list to the given slice.
//
//   List(1, 2, 3).AppendToSlice([]int(nil)) => [1, 2, 3]
//
func (list *Pair) AppendToSlice(slice interface{}) (result interface{}) {
	rslice := reflect.ValueOf(slice)
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		rslice = reflect.Append(rslice, reflect.ValueOf(pair.Car))
	}
	return rslice.Interface()
}

// ToSlice uses Go's reflect package to convert the list to a slice.
//
// If you need a slice of a particular type, use AppendToSlice to
// append to a nil value of that slice type instead.
func (list *Pair) ToSlice() (result []interface{}) {
	return list.AppendToSlice([]interface{}(nil)).([]interface{})
}

// FromSlice uses Go's reflect package to convert the slice to a list.
func FromSlice(slice interface{}) (result *Pair) {
	rslice := reflect.ValueOf(slice)
	length := rslice.Len()
	if length == 0 {
		return
	}
	result = &Pair{Car: rslice.Index(0).Interface()}
	last := result
	for i := 1; i < length; i++ {
		last = last.ncdr(rslice.Index(i).Interface())
	}
	last.Cdr = (*Pair)(nil)
	return
}

// AppendTabulate applies init to each integer i, where 0 <= i < length, and uses Append to append together the results.
// No guarantee is made about the dynamic order in which init is applied to these integers.
func AppendTabulate(length int, init func(int) *Pair) (result *Pair) {
	if length < 0 {
		panic(negativeLength(length))
	}
	for i := 0; i < length; i++ {
		if res := init(i); res != nil {
			if i++; i == length {
				result = res
				return
			}
			var last *Pair
			result, last = copyList(res)
			for {
				if cdr := init(i); cdr != nil {
					if i++; i == length {
						last.Cdr = cdr
						return
					}
					last.Cdr, last = copyList(cdr)
				} else if i++; i == length {
					return
				}
			}
		}
	}
	return
}

// NAppendTabulate applies init to each integer i, where 0 <= i < length, and uses NAppend to append together the results.
// No guarantee is made about the dynamic order in which init is applied to these integers.
func NAppendTabulate(length int, init func(int) *Pair) (result *Pair) {
	if length < 0 {
		panic(negativeLength(length))
	}
	for i := 0; i < length; i++ {
		if result = init(i); result != nil {
			if i++; i == length {
				return
			}
			last := result.LastPair()
			for {
				if cdr := init(i); cdr != nil {
					last.Cdr = cdr
					if i++; i == length {
						return
					}
					last = cdr.LastPair()
				} else if i++; i == length {
					return
				}
			}
		}
	}
	return
}

// Length returns the length of the argument. It is an error to pass a value which is not a proper list
// (finite and Nil()-terminated). In particular, this means Length may diverge or panic when Length
// is applied to a circular list.
//
// The length of a proper list is a non-negative integer n such that Cdr applied n times to the list
// produces the empty list.
func (list *Pair) Length() (result int) {
	if list == nil {
		return
	}
	x := list.Cdr
	result = 1
	for {
		pair, _ := x.(*Pair)
		if pair == nil {
			return
		}
		result++
		x = pair.Cdr
	}
}

// NonCircularLength returns the length of the argument and true if list is a proper list.
// If list is circular, though, NonCircularLength returns -1 and false.
//
// The length of a proper list is a non-negative integer n such that Cdr applied n times to the list
// produces the empty list.
func (list *Pair) NonCircularLength() (result int, nonCircular bool) {
	if list == nil {
		return 0, true
	}
	result = 1
	lag := list
	for {
		if list, _ = list.Cdr.(*Pair); list == nil {
			return result, true
		}
		result++
		if list, _ = list.Cdr.(*Pair); list == nil { // intentionally a second time
			return result, true
		}
		result++
		if lag = lag.Cdr.(*Pair); list == lag {
			return -1, false
		}
	}
}

// Append returns a list consisting of the elements of the first list followed by the elements of the other lists.
//
//   List(1).Append(List(2))          => (1 2)
//   List(1).Append(List(2, 3, 4))    => (1 2 3 4)
//   List(1, List(2)).Append(List(3)) => (1 (2) 3)
//
// The resulting list is always newly allocated, except that it shares structure with the final list.
func (list *Pair) Append(lists ...*Pair) (result *Pair) {
	return Append(list, Append(lists...))
}

// Append returns a list consisting of the elements of the first list followed by the elements of the other lists.
//
//   Append(List(1), List(2))          => (1 2)
//   Append(List(1), List(2, 3, 4))    => (1 2 3 4)
//   Append(List(1, List(2)), List(3)) => (1 (2) 3)
//
// The resulting list is always newly allocated, except that it shares structure with the final list.
func Append(lists ...*Pair) (result *Pair) {
	lastIndex := len(lists) - 1
	for index, pair := range lists {
		if pair != nil {
			if index == lastIndex {
				return pair
			}
			res, last := copyList(pair)
			result = res
			for index, pair = range lists[index+1:] {
				if pair != nil {
					if index == lastIndex {
						last.Cdr = pair
						return
					}
					last.Cdr, last = copyList(pair)
				}
			}
			return
		}
	}
	return
}

// NAppend is the linear-update variant of Append -- it is allowed, but not required, to alter cons
// cells in the argument lists to construct the result list. The last list is never altered; the result
// list shares structure with this parameter.
func (list *Pair) NAppend(lists ...*Pair) (result *Pair) {
	return NAppend(list, NAppend(lists...))
}

// NAppend is the linear-update variant of Append -- it is allowed, but not required, to alter cons
// cells in the argument lists to construct the result list. The last list is never altered; the result
// list shares structure with this parameter.
func NAppend(lists ...*Pair) (result *Pair) {
	lastIndex := len(lists) - 1
	for index, pair := range lists {
		if pair != nil {
			if index == lastIndex {
				return pair
			}
			result = pair
			last := pair.LastPair()
			for index, pair = range lists[index+1:] {
				if pair != nil {
					last.Cdr = pair
					if index == lastIndex {
						return
					}
					last = pair.LastPair()
				}
			}
			return
		}
	}
	return
}

// AppendLast is like Append, except that the whole result is newly allocated and does not share any structure with
// any of its arguments. AppendLast returns both the resulting list as well as the last pair of the resulting list.
// This enables setting the Cdr of the last pair to a value of a type other than *Pair.
func AppendLast(lists ...*Pair) (result *Pair, last *Pair) {
	for index, pair := range lists {
		if pair != nil {
			result, last = copyList(pair)
			for _, pair = range lists[index+1:] {
				if pair != nil {
					last.Cdr, last = copyList(pair)
				}
			}
			return
		}
	}
	return
}

// NAppendLast is the linear-update variant of AppendLast -- it is allowed, but not required, to alter cons
// cells in the argument lists to construct the result list, including the last list.
func NAppendLast(lists ...*Pair) (result *Pair, last *Pair) {
	for index, pair := range lists {
		if pair != nil {
			result, last = pair, pair.LastPair()
			for _, pair = range lists[index+1:] {
				if pair != nil {
					last.Cdr, last = pair, pair.LastPair()
				}
			}
			return
		}
	}
	return
}

// Reverse returns a newly allocated list consisting of the elements of list in reverse order.
// The list must be a proper list.
func (list *Pair) Reverse() (result *Pair) {
	return list.AppendReverse(nil)
}

// NReverse is the linear-update variant of Reverse.
// The list must be a proper list.
func (list *Pair) NReverse() (result *Pair) {
	return list.NAppendReverse(nil)
}

// ReverseLast is like Reverse, except that both the resulting list as well as the last pair of
// the resulting list is returned. This enables setting the Cdr of the last pair to a value other than Nil().
func (list *Pair) ReverseLast() (result, last *Pair) {
	if list == nil {
		return
	}
	last = &Pair{Car: list.Car, Cdr: (*Pair)(nil)}
	result = list.Cdr.(*Pair).AppendReverse(last)
	return
}

// NReverseLast is the linear-update variant of ReverseLast.
func (list *Pair) NReverseLast() (result, last *Pair) {
	return list.NAppendReverse(nil), list
}

// AppendReverse returns Append(list.Reverse(), tail).
//
// It is provided because it is a common operation -- a common list-processing style calls for this
// exact operation to transfer values accumulated in reverse order onto the front of another list, and
// because the implementation is significantly more efficient than the simple composition it replaces.
// (But note that this pattern of iterative computation followed by a reverse can frequently be rewritten
// as a recursion, disposing with the Reverse and AppendReverse steps, and shifting temporary, intermediate
// storage from the heap to the stack.)
func (list *Pair) AppendReverse(tail *Pair) (result *Pair) {
	result = tail
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		result = &Pair{Car: pair.Car, Cdr: result}
	}
	return
}

// NAppendReverse is the linear-update variant of AppendReverse.
func (list *Pair) NAppendReverse(tail *Pair) (result *Pair) {
	result = tail
	for pair := list; pair != nil; {
		pair, pair.Cdr, result = pair.Cdr.(*Pair), result, pair
	}
	return
}

// Zip returns a list of the same length, each element of which is a one-element
// list comprised of the corresponding elements from list. The list must be finite.
//
//   List(1, 2, 3).Zip() => ((1) (2) (3))
//
func (list *Pair) Zip() (result *Pair) {
	if list == nil {
		return
	}
	result = &Pair{Car: List(list.Car)}
	last := result
	for pair := list.Cdr.(*Pair); pair != nil; pair = pair.Cdr.(*Pair) {
		last = last.ncdr(List(pair.Car))
	}
	last.Cdr = (*Pair)(nil)
	return
}

func carList(lists ...*Pair) (result *Pair) {
	result = &Pair{Car: lists[0].Car}
	last := result
	for _, p := range lists[1:] {
		last = last.ncdr(p.Car)
	}
	last.Cdr = (*Pair)(nil)
	return
}

// Zip returns a list as long as the shortest of the argument lists, each
// element of which is an n-element list comprised of the corresponding elements
// from the parameter lists, where n is the number of lists passed to Zip.
//
//   Zip(List("one", "two", "three"),
//       List(1, 2, 3),
//       List("odd", "even", "odd", "even", "odd", "even", "odd", "even"))
//    => (("one" 1 "odd") ("two" 2 "even") ("three" 3 "odd"))
//
// At least one of the argument lists must be finite:
//
//   Zip(List(3, 1, 4, 1), Circular(false, true))
//    => ((3 false) (1 true) (4 false) (1 true))
//
func Zip(lists ...*Pair) (result *Pair) {
	switch len(lists) {
	case 0:
		return
	case 1:
		return lists[0].Zip()
	}
	a, ok := initCdrSlice(lists)
	if !ok {
		return
	}
	result = &Pair{Car: carList(a...)}
	last := result
	for ok = a.next(); ok; ok = a.next() {
		last = last.ncdr(carList(a...))
	}
	last.Cdr = (*Pair)(nil)
	return
}

// Unzip takes a list, which must contain at least n elements,
// and returns a slice of length n of lists. The first result list contains
// the first element of the list, the second result list
// contains the second element the list, and so on.
//
//   List(1, "one").Unzip(2) => [(1), ("one")]
//
func (list *Pair) Unzip(n int) (result []*Pair) {
	if n < 0 {
		panic(negativeLength(n))
	}
	for pair, i := list, 0; i < n; pair, i = pair.Cdr.(*Pair), i+1 {
		result = append(result, &Pair{Car: pair.Car, Cdr: (*Pair)(nil)})
	}
	return
}

// Unzip takes several lists, where every list must contain at least n elements,
// and returns a slice of length n of lists. The first result list contains
// the first element of each list, the second result list contains the second
// element of each list, and so on.
//
//   Unzip(2, List(1, "one"), List(2, "two"), List(3, "three"))
//    => [(1 2 3), ("one" "two" "three")]
//
func Unzip(n int, lists ...*Pair) (result []*Pair) {
	switch len(lists) {
	case 0:
		return
	case 1:
		return lists[0].Unzip(n)
	}
	if n == 0 {
		return
	}
	if n < 0 {
		panic(negativeLength(n))
	}
	i := 0
	a, ok := initListArgs(lists)
	for {
		if !ok {
			panic("not enough elements in lists")
		}
		result = append(result, a.args)
		if i++; i == n {
			return
		}
		ok = a.next()
	}
}

// Count applies predicate element-wise to the elements of list, and a count
// is tallied of the number of elements that produce a true value. This count
// is returned. Count is guaranteed to apply predicate to the list elements
// in a left-to-right order. The list must be finite.
//
//   func even(x interface{}) bool {
//     return x.(int)%2 == 0
//   }
//
//   List(3, 1, 4, 1, 5, 9, 2, 5, 6).Count(even) => 3
//
func (list *Pair) Count(predicate func(interface{}) bool) (result int) {
	for pair := list; pair != nil; pair = pair.Cdr.(*Pair) {
		if predicate(pair.Car) {
			result++
		}
	}
	return
}

// Count applies predicate element-wise to the elements of the lists, and a count
// is tallied of the number of elements that produce a true value. This count
// is returned. Count is guaranteed to apply predicate to the list elements
// in a left-to-right order. The counting stops when the shortest list expires.
//
//   func lessThan(xs ...interface{}) bool {
//     return x[0].(int) < x[1].(int)
//   }
//
//   Count(lessThan, List(1, 2, 4, 8), List(2, 4, 6, 8, 10, 12, 14, 16)) => 3
//
// At least one of the argument lists must be finite:
//
//   Count(lessThan, List(3, 1, 4, 1), Circular(1, 10)) => 2
//
func Count(predicate func(...interface{}) bool, lists ...*Pair) (result int) {
	if len(lists) == 0 {
		return
	}
	for a, ok := initCarArgs(lists); ok; ok = a.next() {
		if predicate(a.args...) {
			result++
		}
	}
	return
}
