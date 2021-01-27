package list

// Car returns nil if x is Nil(), x.Car if x is a non-nil *Pair, and otherwise nil.
func Car(x interface{}) interface{} {
	pair, _ := x.(*Pair)
	if pair == nil {
		return nil
	}
	return pair.Car
}

// Cdr returns nil if x is Nil(), x.Cdr if x is a non-nil *Pair, and otherwise nil.
func Cdr(x interface{}) interface{} {
	pair, _ := x.(*Pair)
	if pair == nil {
		return nil
	}
	return pair.Cdr
}

// Caar returns Car(Car(x))
func Caar(x interface{}) interface{} {
	return Car(Car(x))
}

// Cadr returns Car(Cdr(x))
func Cadr(x interface{}) interface{} {
	return Car(Cdr(x))
}

// Cdar returns Cdr(Car(x))
func Cdar(x interface{}) interface{} {
	return Cdr(Car(x))
}

// Cddr returns Cdr(Cdr(x))
func Cddr(x interface{}) interface{} {
	return Cdr(Cdr(x))
}

// Caaar returns Car(Car(Car(x)))
func Caaar(x interface{}) interface{} {
	return Caar(Car(x))
}

// Caadr returns Car(Car(Cdr(x)))
func Caadr(x interface{}) interface{} {
	return Caar(Cdr(x))
}

// Cadar returns Car(Cdr(Car(x)))
func Cadar(x interface{}) interface{} {
	return Cadr(Car(x))
}

// Caddr returns Car(Cdr(Cdr(x)))
func Caddr(x interface{}) interface{} {
	return Cadr(Cdr(x))
}

// Cdaar returns Cdr(Car(Car(x)))
func Cdaar(x interface{}) interface{} {
	return Cdar(Car(x))
}

// Cdadr returns Cdr(Car(Cdr(x)))
func Cdadr(x interface{}) interface{} {
	return Cdar(Cdr(x))
}

// Cddar returns Cdr(Cdr(Car(x)))
func Cddar(x interface{}) interface{} {
	return Cddr(Car(x))
}

// Cdddr returns Cdr(Cdr(Cdr(x)))
func Cdddr(x interface{}) interface{} {
	return Cddr(Cdr(x))
}

// Caaaar returns Car(Car(Car(Car(x))))
func Caaaar(x interface{}) interface{} {
	return Caaar(Car(x))
}

// Caaadr returns Car(Car(Car(Cdr(x))))
func Caaadr(x interface{}) interface{} {
	return Caaar(Cdr(x))
}

// Caadar returns Car(Car(Cdr(Car(x))))
func Caadar(x interface{}) interface{} {
	return Caadr(Car(x))
}

// Caaddr returns Car(Car(Cdr(Cdr(x))))
func Caaddr(x interface{}) interface{} {
	return Caadr(Cdr(x))
}

// Cadaar returns Car(Cdr(Car(Car(x))))
func Cadaar(x interface{}) interface{} {
	return Cadar(Car(x))
}

// Cadadr returns Car(Cdr(Car(Cdr(x))))
func Cadadr(x interface{}) interface{} {
	return Cadar(Cdr(x))
}

// Caddar returns Car(Cdr(Cdr(Car(x))))
func Caddar(x interface{}) interface{} {
	return Caddr(Car(x))
}

// Cadddr returns Car(Cdr(Cdr(Cdr(x))))
func Cadddr(x interface{}) interface{} {
	return Caddr(Cdr(x))
}

// Cdaaar returns Cdr(Car(Car(Car(x))))
func Cdaaar(x interface{}) interface{} {
	return Cdaar(Car(x))
}

// Cdaadr returns Cdr(Car(Car(Cdr(x))))
func Cdaadr(x interface{}) interface{} {
	return Cdaar(Cdr(x))
}

// Cdadar returns Cdr(Car(Cdr(Car(x))))
func Cdadar(x interface{}) interface{} {
	return Cdadr(Car(x))
}

// Cdaddr returns Cdr(Car(Cdr(Cdr(x))))
func Cdaddr(x interface{}) interface{} {
	return Cdadr(Cdr(x))
}

// Cddaar returns Cdr(Cdr(Car(Car(x))))
func Cddaar(x interface{}) interface{} {
	return Cddar(Car(x))
}

// Cddadr returns Cdr(Cdr(Car(Cdr(x))))
func Cddadr(x interface{}) interface{} {
	return Cddar(Cdr(x))
}

// Cdddar returns Cdr(Cdr(Cdr(Car(x))))
func Cdddar(x interface{}) interface{} {
	return Cdddr(Car(x))
}

// Cddddr returns Cdr(Cdr(Cdr(Cdr(x))))
func Cddddr(x interface{}) interface{} {
	return Cdddr(Cdr(x))
}

// Ref returns the nth element of list.
//
// (This is the same as the Car of list.Drop(n).)
// Ref panics if n >= l, where l is the length of list.
func (list *Pair) Ref(n int) (result interface{}) {
	if n >= 0 {
		for l, i := list, 0; l != nil; i++ {
			if i == n {
				return l.Car
			}
			l, _ = l.Cdr.(*Pair)
		}
	}
	panic(outOfBounds(n, list))
}

// Take returns the first k elements of the list.
//
//   List(1, 2, 3, 4, 5).Take(2) => (1 2)
//
// x may be any value -- a proper, circular, or dotted list.
//
//   Cons(1, 2, 3, "d").Take(2) => (1 2)
//   Cons(1, 2, 3, "d").Take(3) => (1 2 3)
//
// For a legal k, Take and Drop partition the list in a manner which can be inverted with Append:
//
//   Append(x.Take(k), x.Drop(k)) => x
//
// If the argument list is a list of non-zero length, Take is guaranteed to return a
// freshly-allocated list, even in the case where the entire list is taken.
func (list *Pair) Take(k int) (result *Pair) {
	if k == 0 {
		return
	}
	if k < 0 || list == nil {
		panic(outOfBounds(k, list))
	}
	result = &Pair{Car: list.Car}
	pair := list
	last := result
	for i := k - 1; i > 0; i-- {
		if pair, _ = pair.Cdr.(*Pair); pair == nil {
			panic(outOfBounds(k, list))
		}
		last = last.ncdr(pair.Car)
	}
	last.Cdr = (*Pair)(nil)
	return
}

// Drop returns all but the first k elements of the list.
//
//   List(1, 2, 3, 4, 5).Drop(2) => (3 4 5)
//
// x may be any value -- a proper, circular, or dotted list.
//
//   Cons(1, 2, 3, "d").Drop(2) => (3 . "d")
//   Cons(1, 2, 3, "d").Drop(3) => "d"
//
// For a legal k, Take and Drop partition the list in a manner which can be inverted with Append:
//
//   Append(x.Take(k), x.Drop(k)) => x
//
// Drop is exactly equivalent to performing k Cdr operations on x; the returned value shares a
// common tail with x.
func (list *Pair) Drop(k int) (result interface{}) {
	if k < 0 {
		panic(outOfBounds(k, list))
	}
	result = list
	for i := k; i > 0; i-- {
		pair, _ := result.(*Pair)
		if pair == nil {
			panic(outOfBounds(k, list))
		}
		result = pair.Cdr
	}
	return
}

// TakeRight returns the last k elements of list.
//
//   List(1, 2, 3, 4, 5).TakeRight(2) => (4 5)
//
// list may be any finite list, either proper or dotted.
//
//   Cons(1, 2, 3, "d").TakeRight(2) => (2 3 . "d")
//   Cons(1, 2, 3, "d").TakeRight(0) => "d"
//
// For a legal k, TakeRight and DropRight partition the list in a manner which can be inverted with Append:
//
//   Append(x.DropRight(k), x.TakeRight(k)) => x
//
// TakeRight's return value is guaranteed to share a common tail with list.
func (list *Pair) TakeRight(k int) (result interface{}) {
	result = list
	lead, _ := list.Drop(k).(*Pair)
	for lead != nil {
		result = result.(*Pair).Cdr
		lead, _ = lead.Cdr.(*Pair)
	}
	return
}

// DropRight returns all but the last k elements of list.
//
//   List(1, 2, 3, 4, 5).DropRight(2) => (1 2 3)
//
// list may be any finite list, either proper or dotted.
//
//   Cons(1, 2, 3, "d").DropRight(2) => (1)
//   Cons(1, 2, 3, "d").DropRight(0) => (1 2 3)
//
// For a legal k, TakeRight and DropRight partition the list in a manner which can be inverted with Append:
//
//   Append(x.DropRight(k), x.TakeRight(k)) => x
//
// If the argument list is a list of non-zero length, DropRight is guaranteed to return a freshly-allocated
// list, even in the case where nothing is dropped.
func (list *Pair) DropRight(k int) (result *Pair) {
	lead, _ := list.Drop(k).(*Pair)
	if lead == nil {
		return
	}
	result = &Pair{Car: list.Car}
	last := result
	for {
		lead, _ = lead.Cdr.(*Pair)
		if lead == nil {
			last.Cdr = (*Pair)(nil)
			return
		}
		list = list.Cdr.(*Pair)
		last = last.ncdr(list.Car)
	}
}

// NTake is the linear-update variant of Take.
//
// If x is circular, NTake may return a shorter-than-expected list.
func (list *Pair) NTake(k int) (result *Pair) {
	if k == 0 {
		return
	}
	pair, _ := list.Drop(k - 1).(*Pair)
	if pair == nil {
		panic(outOfBounds(k, list))
	}
	pair.Cdr = (*Pair)(nil)
	return list
}

// NDropRight is the linear-update variant of DropRight.
func (list *Pair) NDropRight(k int) (result *Pair) {
	lead, _ := list.Drop(k).(*Pair)
	if lead == nil {
		return
	}
	lag := list
	lead, _ = lead.Cdr.(*Pair)
	for lead != nil {
		lag = lag.Cdr.(*Pair)
		lead, _ = lead.Cdr.(*Pair)
	}
	lag.Cdr = (*Pair)(nil)
	return list
}

// SplitAt splits the list at index k, returning a list of the first k elements, and the remaining tail.
func (list *Pair) SplitAt(k int) (prefix *Pair, suffix interface{}) {
	if k < 0 {
		panic(outOfBounds(k, list))
	}
	if k == 0 {
		suffix = list
		return
	}
	if list == nil {
		panic(outOfBounds(k, list))
	}
	prefix = &Pair{Car: list.Car}
	last := prefix
	for i := k - 1; i > 0; i-- {
		if list, _ = list.Cdr.(*Pair); list == nil {
			panic(outOfBounds(k, list))
		}
		last = last.ncdr(list.Car)
	}
	last.Cdr = (*Pair)(nil)
	suffix = list.Cdr
	return
}

// NSplitAt is the linear-update variant of SplitAt.
func (list *Pair) NSplitAt(k int) (prefix *Pair, suffix interface{}) {
	if k < 0 {
		panic(outOfBounds(k, list))
	}
	if k == 0 {
		suffix = list
		return
	}
	prev, _ := list.Drop(k - 1).(*Pair)
	if prev == nil {
		panic(outOfBounds(k, list))
	}
	prefix = list
	suffix = prev.Cdr
	prev.Cdr = (*Pair)(nil)
	return
}

// Last returns the last element of the finite list. If list is nil or dotted, Last returns Nil().
func (list *Pair) Last() (result interface{}) {
	return Car(list.LastPair())
}

// LastPair returns the last pair in the finite list. If list is nil or dotted, LastPair returns Nil().
func (list *Pair) LastPair() (result *Pair) {
	if list == nil {
		return
	}
	result = list
	for {
		cdr, _ := result.Cdr.(*Pair)
		if cdr == nil {
			return
		}
		result = cdr
	}
}
