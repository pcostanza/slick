package list

// Assoc finds the first pair in alist whose Car field is key, and returns that pair and true. If no
// pair in alist has key as its Car, then nil and false are returned. Assoc uses ==
// for comparing key against the cars in alist.
func (alist *Pair) Assoc(key interface{}) (result interface{}, ok bool) {
	return alist.Find(func(x interface{}) bool { return key == x.(*Pair).Car })
}

// ACons conses a new alist entry mapping key -> value onto alist.
func (alist *Pair) ACons(key, value interface{}) *Pair {
	return NewPair(NewPair(key, value), alist)
}

// ACopy makes a fresh copy of alist. This means copying each pair that forms an assocation
// as well as the spine of the list.
func (alist *Pair) ACopy() *Pair {
	return alist.Map(func(x interface{}) interface{} {
		pair := x.(*Pair)
		return NewPair(pair.Car, pair.Cdr)
	})
}

// ADelete deletes all assocations from alist with the given key, using ==.
//
// The return value may share common tails with the alist argument. The alist is not
// disordered -- elements that appear in the result alist occur in the same order as
// they occur in the argument list.
func (alist *Pair) ADelete(key interface{}) *Pair {
	return alist.Remove(func(x interface{}) bool { return key == x.(*Pair).Car })
}

// NADelete is the linear-update variant of ADelete.
func (alist *Pair) NADelete(key interface{}) *Pair {
	return alist.NRemove(func(x interface{}) bool { return key == x.(*Pair).Car })
}
