package list

// IsProper returns true iff x is a proper list -- a finite, Nil()-terminated list.
//
// More carefully: The empty list (that is, (*Pair)(nil)) is a proper list.
// A pair whose Cdr is a proper list is also a proper list.
//
// Note that this definition rules out circular lists. This function detects circular lists
// and returns false in this case.
//
// Lists that end in nil (of type interface{}) are not proper, but dotted lists.
func IsProper(x interface{}) bool {
	pair, ok := x.(*Pair)
	if pair == nil {
		return ok
	}
	lag := pair
	for {
		if pair, ok = pair.Cdr.(*Pair); pair == nil {
			return ok
		}
		if pair, ok = pair.Cdr.(*Pair); pair == nil { // intentionally a second time
			return ok
		}
		if lag = lag.Cdr.(*Pair); pair == lag {
			return false
		}
	}
}

// IsCircular returns true if x is a circular list.
//
// A circular list is a value such that for every n >= 0, Cdr^n(x) is a pair.
//
// Terminology: The opposite of circular is finite.
func IsCircular(x interface{}) bool {
	pair, _ := x.(*Pair)
	if pair == nil {
		return false
	}
	lag := pair
	for {
		if pair, _ = pair.Cdr.(*Pair); pair == nil {
			return false
		}
		if pair, _ = pair.Cdr.(*Pair); pair == nil { // intentionally a second time
			return false
		}
		if lag = lag.Cdr.(*Pair); pair == lag {
			return true
		}
	}
}

// IsDotted returns true if x is a finite, non-Nil()-terminated list.
//
// That is, there exists an n >= 0 such that Cdr^n(x) is not of type *Pair.
// This includes non-pair values which are considered to be dotted lists
// of length 0.
func IsDotted(x interface{}) bool {
	pair, ok := x.(*Pair)
	if pair == nil {
		return !ok
	}
	lag := pair
	for {
		if pair, ok = pair.Cdr.(*Pair); pair == nil {
			return !ok
		}
		if pair, ok = pair.Cdr.(*Pair); pair == nil { // intentionally a second time
			return !ok
		}
		if lag = lag.Cdr.(*Pair); pair == lag {
			return false
		}
	}
}

// IsEnd returns true if x is not of type *Pair, or if x is Nil().
func IsEnd(x interface{}) bool {
	pair, _ := x.(*Pair)
	return pair == nil
}

// IsProperEnd returns true if x is of type *Pair and Nil().
func IsProperEnd(x interface{}) bool {
	pair, ok := x.(*Pair)
	return ok && pair == nil
}

// IsDottedEnd returns true if x is not of type *Pair.
func IsDottedEnd(x interface{}) bool {
	_, ok := x.(*Pair)
	return !ok
}

// IsNilPair returns true if x is of type *Pair and Nil(). If x is not of type *Pair, IsNilPair panics.
func IsNilPair(x interface{}) bool {
	return x.(*Pair) == nil
}

// IsPair returns true if x is of type *Pair.
func IsPair(x interface{}) bool {
	_, ok := x.(*Pair)
	return ok
}

// Equal determines list equality.
//
// Proper list A equals proper list B if they are of the same length,
// and their corresponding elements are ==.
//
// It is an error to apply Equal to circular lists.
func Equal(x, y interface{}) bool {
	for {
		pair1, ok := x.(*Pair)
		if !ok {
			return x == y
		}
		pair2, ok := y.(*Pair)
		if !ok {
			return false
		}
		if pair1 == pair2 {
			return true
		}
		if pair1 == nil || pair2 == nil {
			return false
		}
		if pair1.Car != pair2.Car {
			return false
		}
		x = pair1.Cdr
		y = pair2.Cdr
	}
}
