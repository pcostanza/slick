package list_test

import (
	"testing"

	"github.com/exascience/slick/list"
)

func TestString(t *testing.T) {
	t.Run("Empty list String", func(t *testing.T) {
		if list.List().String() != "()" {
			t.Fail()
		}
	})
	t.Run("Proper list String", func(t *testing.T) {
		if list.List(1, 2, 3).String() != "(1 2 3)" {
			t.Fail()
		}
	})
	t.Run("Dotted list String", func(t *testing.T) {
		if list.Cons(1, 2, 3).String() != "(1 2 . 3)" {
			t.Fail()
		}
	})
}

func TestConstructors(t *testing.T) {
	t.Run("NewList", func(t *testing.T) {
		if list.NewList(0, 42) != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.NewList(5, 42), list.List(42, 42, 42, 42, 42)) {
			t.Fail()
		}
	})
	t.Run("Tabulate", func(t *testing.T) {
		if list.Tabulate(0, nil) != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.Tabulate(5, func(i int) interface{} { return i * i }), list.List(0, 1, 4, 9, 16)) {
			t.Fail()
		}
	})
	t.Run("Copy", func(t *testing.T) {
		l1 := list.List(1, 2, 3, 4, 5)
		l2 := l1.Copy()
		if !list.Equal(l1, l2) {
			t.Fail()
		}
		list.PairForEach(func(ls ...*list.Pair) {
			if ls[0] == ls[1] {
				t.Fail()
			}
		}, l1, l2)
		l1 = list.Nil()
		l2 = l1.Copy()
		if !list.Equal(l1, l2) {
			t.Fail()
		}
		if l2 != list.Nil() {
			t.Fail()
		}
	})
	t.Run("Circular", func(t *testing.T) {
		l := list.Circular(1, 2, 3)
		if !list.Equal(l.Take(7), list.List(1, 2, 3, 1, 2, 3, 1)) {
			t.Fail()
		}
		l = list.Circular(42)
		if !list.Equal(l.Take(3), list.List(42, 42, 42)) {
			t.Fail()
		}
	})
}

func TestPredicates(t *testing.T) {
	proper := list.List(1, 2, 3)
	dotted := list.Cons(1, 2, 3)
	circular := list.Circular(1, 2, 3)
	t.Run("IsProper", func(t *testing.T) {
		if !list.IsProper(list.Nil()) {
			t.Fail()
		}
		if !list.IsProper(proper) {
			t.Fail()
		}
		if list.IsProper(dotted) {
			t.Fail()
		}
		if list.IsProper(circular) {
			t.Fail()
		}
	})
	t.Run("IsDotted", func(t *testing.T) {
		if list.IsDotted(list.Nil()) {
			t.Fail()
		}
		if list.IsDotted(proper) {
			t.Fail()
		}
		if !list.IsDotted(dotted) {
			t.Fail()
		}
		if list.IsDotted(circular) {
			t.Fail()
		}
	})
	t.Run("IsCircular", func(t *testing.T) {
		if list.IsCircular(list.Nil()) {
			t.Fail()
		}
		if list.IsCircular(proper) {
			t.Fail()
		}
		if list.IsCircular(dotted) {
			t.Fail()
		}
		if !list.IsCircular(circular) {
			t.Fail()
		}
	})
	t.Run("IsEnd", func(t *testing.T) {
		if !list.IsEnd(list.Nil()) {
			t.Fail()
		}
		if !list.IsEnd(nil) {
			t.Fail()
		}
		if !list.IsEnd(42) {
			t.Fail()
		}
		if list.IsEnd(proper) {
			t.Fail()
		}
		if list.IsEnd(dotted) {
			t.Fail()
		}
		if list.IsEnd(circular) {
			t.Fail()
		}
	})
	t.Run("IsProperEnd", func(t *testing.T) {
		if !list.IsProperEnd(list.Nil()) {
			t.Fail()
		}
		if list.IsProperEnd(nil) {
			t.Fail()
		}
		if list.IsProperEnd(42) {
			t.Fail()
		}
		if list.IsProperEnd(proper) {
			t.Fail()
		}
		if list.IsProperEnd(dotted) {
			t.Fail()
		}
		if list.IsProperEnd(circular) {
			t.Fail()
		}
	})
	t.Run("IsDottedEnd", func(t *testing.T) {
		if list.IsDottedEnd(list.Nil()) {
			t.Fail()
		}
		if !list.IsDottedEnd(nil) {
			t.Fail()
		}
		if !list.IsDottedEnd(42) {
			t.Fail()
		}
		if list.IsDottedEnd(proper) {
			t.Fail()
		}
		if list.IsDottedEnd(dotted) {
			t.Fail()
		}
		if list.IsDottedEnd(circular) {
			t.Fail()
		}
	})
	t.Run("IsNilPair", func(t *testing.T) {
		if !list.IsNilPair(list.Nil()) {
			t.Fail()
		}
		testNotNilPair := func(x interface{}) (result bool) {
			defer func() {
				if p := recover(); p != nil {
					result = true
					return
				}
				panic("expected panic")
			}()
			return !list.IsNilPair(x)
		}
		if !testNotNilPair(nil) {
			t.Fail()
		}
		if !testNotNilPair(42) {
			t.Fail()
		}
		if list.IsNilPair(proper) {
			t.Fail()
		}
		if list.IsNilPair(dotted) {
			t.Fail()
		}
		if list.IsNilPair(circular) {
			t.Fail()
		}
	})
}

func TestSelectors(t *testing.T) {
	t.Run("Cars and Cdrs", func(t *testing.T) {
		p := list.NewPair(1, 2)
		if list.Car(p) != 1 {
			t.Fail()
		}
		if list.Cdr(p) != 2 {
			t.Fail()
		}
		if list.Car(list.Nil()) != nil {
			t.Fail()
		}
		if list.Cdr(list.Nil()) != nil {
			t.Fail()
		}
		if list.Car(42) != nil {
			t.Fail()
		}
		if list.Cdr(42) != nil {
			t.Fail()
		}
		if list.Caaaar(p) != nil {
			t.Fail()
		}
		if list.Cddddr(p) != nil {
			t.Fail()
		}
	})
	t.Run("Ref", func(t *testing.T) {
		l := list.List(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
		for i := 0; i < 12; i++ {
			if l.Ref(i) != i+1 {
				t.Fail()
			}
		}
	})
	t.Run("Take and Drop", func(t *testing.T) {
		if list.Nil().Take(0) != list.Nil() {
			t.Fail()
		}
		if list.Nil().Drop(0) != list.Nil() {
			t.Fail()
		}
		if list.Nil().TakeRight(0) != list.Nil() {
			t.Fail()
		}
		if list.Nil().DropRight(0) != list.Nil() {
			t.Fail()
		}
		l := list.List(1, 2, 3, 4, 5, 6)
		if !list.Equal(l.Take(4), list.List(1, 2, 3, 4)) {
			t.Fail()
		}
		if !list.Equal(l.Drop(4), list.List(5, 6)) {
			t.Fail()
		}
		if !list.Equal(list.Cons(1, 2, 3, "d").Drop(2), list.Cons(3, "d")) {
			t.Fail()
		}
		if list.Cons(1, 2, 3, "d").Drop(3) != "d" {
			t.Fail()
		}
		if !list.Equal(l.TakeRight(2), list.List(5, 6)) {
			t.Fail()
		}
		if !list.Equal(list.Cons(1, 2, 3, "d").TakeRight(2), list.Cons(2, 3, "d")) {
			t.Fail()
		}
		if list.Cons(1, 2, 3, "d").TakeRight(0) != "d" {
			t.Fail()
		}
		if !list.Equal(l.DropRight(2), list.List(1, 2, 3, 4)) {
			t.Fail()
		}
		if !list.Equal(l, list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
		if list.Nil().NTake(0) != list.Nil() {
			t.Fail()
		}
		if !list.Equal(l.NTake(4), list.List(1, 2, 3, 4)) {
			t.Fail()
		}
		if list.Nil().NDropRight(0) != list.Nil() {
			t.Fail()
		}
		l = list.List(1, 2, 3, 4, 5, 6)
		if !list.Equal(l.NDropRight(2), list.List(1, 2, 3, 4)) {
			t.Fail()
		}
	})
	t.Run("SplitAt", func(t *testing.T) {
		l := list.List(1, 2, 3, 4, 5, 6)
		if p, s := l.SplitAt(3); !list.Equal(p, list.List(1, 2, 3)) || !list.Equal(s, list.List(4, 5, 6)) {
			t.Fail()
		}
		if p, s := l.SplitAt(0); p != list.Nil() || !list.Equal(s, l) {
			t.Fail()
		}
		if p, s := l.NSplitAt(0); p != list.Nil() || !list.Equal(s, l) {
			t.Fail()
		}
		if p, s := l.SplitAt(6); !list.Equal(p, l) || s != list.Nil() {
			t.Fail()
		}
		if p, s := l.NSplitAt(6); !list.Equal(p, l) || s != list.Nil() {
			t.Fail()
		}
		if !list.Equal(l, list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
		if p, s := l.NSplitAt(3); !list.Equal(p, list.List(1, 2, 3)) || !list.Equal(s, list.List(4, 5, 6)) {
			t.Fail()
		}
		if p, s := list.Cons(1, 2, 3, 4).SplitAt(3); !list.Equal(p, list.List(1, 2, 3)) || s != 4 {
			t.Fail()
		}
		if p, s := list.Cons(1, 2, 3, 4).NSplitAt(3); !list.Equal(p, list.List(1, 2, 3)) || s != 4 {
			t.Fail()
		}
	})
	t.Run("Last and LastPair", func(t *testing.T) {
		if list.Nil().LastPair() != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.Cons(1, 2, 3).LastPair(), list.Cons(2, 3)) {
			t.Fail()
		}
		if l := list.Cons(1, 2, 3); l.LastPair() != list.Cdr(l) {
			t.Fail()
		}
		if list.List(1, 2, 3).Last() != 3 {
			t.Fail()
		}
		if list.Cons(1, 2, 3).Last() != 2 {
			t.Fail()
		}
		if list.Nil().Last() != nil {
			t.Fail()
		}
	})
}

func TestMiscellaneous(t *testing.T) {
	t.Run("AppendToSlice", func(t *testing.T) {
		for i, x := range list.List(3, 4, 5).AppendToSlice([]int{0, 1, 2}).([]int) {
			if i != x {
				t.Fail()
			}
		}
		for i, x := range list.List(0, 1, 2, 3, 4).ToSlice() {
			if i != x.(int) {
				t.Fail()
			}
		}
	})
	t.Run("FromSlice", func(t *testing.T) {
		if list.FromSlice([]string{}) != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.FromSlice([]int{1, 2, 3}), list.List(1, 2, 3)) {
			t.Fail()
		}
	})
	t.Run("AppendTabulate", func(t *testing.T) {
		if !list.Equal(list.AppendTabulate(5, func(i int) *list.Pair {
			if i%2 == 0 {
				return list.List(i)
			}
			return nil
		}), list.List(0, 2, 4)) {
			t.Fail()
		}
		if list.AppendTabulate(0, func(i int) *list.Pair { return list.List(42) }) != nil {
			t.Fail()
		}
		if list.AppendTabulate(5, func(i int) *list.Pair { return nil }) != nil {
			t.Fail()
		}
		if !list.Equal(list.NAppendTabulate(5, func(i int) *list.Pair {
			if i%2 == 0 {
				return list.List(i)
			}
			return nil
		}), list.List(0, 2, 4)) {
			t.Fail()
		}
		if list.NAppendTabulate(0, func(i int) *list.Pair { return list.List(42) }) != nil {
			t.Fail()
		}
		if list.NAppendTabulate(5, func(i int) *list.Pair { return nil }) != nil {
			t.Fail()
		}
	})
	t.Run("Length", func(t *testing.T) {
		if list.Nil().Length() != 0 {
			t.Fail()
		}
		if list.List(1, 2, 3).Length() != 3 {
			t.Fail()
		}
		if l, ok := list.Nil().NonCircularLength(); !ok || l != 0 {
			t.Fail()
		}
		if l, ok := list.List(1, 2, 3).NonCircularLength(); !ok || l != 3 {
			t.Fail()
		}
		if l, ok := list.Circular(1, 2, 3).NonCircularLength(); ok || l != -1 {
			t.Fail()
		}
	})
	t.Run("Append", func(t *testing.T) {
		if list.Append() != list.Nil() {
			t.Fail()
		}
		if list.NAppend() != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.Append(list.Nil(), list.List(1, 2, 3)), list.List(1, 2, 3)) {
			t.Fail()
		}
		if !list.Equal(list.NAppend(list.Nil(), list.List(1, 2, 3)), list.List(1, 2, 3)) {
			t.Fail()
		}
		if !list.Equal(list.Append(list.List(1, 2, 3), list.Nil()), list.List(1, 2, 3)) {
			t.Fail()
		}
		if !list.Equal(list.NAppend(list.List(1, 2, 3), list.Nil()), list.List(1, 2, 3)) {
			t.Fail()
		}
		if !list.Equal(list.Append(list.List(1, 2, 3), list.List(4, 5, 6)), list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
		if !list.Equal(list.NAppend(list.List(1, 2, 3), list.List(4, 5, 6)), list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
		result, last := list.AppendLast(list.List(1, 2, 3), list.List(4, 5, 6))
		last.Cdr = 7
		if !list.Equal(result, list.Cons(1, 2, 3, 4, 5, 6, 7)) {
			t.Fail()
		}
		result, last = list.NAppendLast(list.List(1, 2, 3), list.List(4, 5, 6))
		last.Cdr = 7
		if !list.Equal(result, list.Cons(1, 2, 3, 4, 5, 6, 7)) {
			t.Fail()
		}
	})
	t.Run("Reverse", func(t *testing.T) {
		if list.Nil().Reverse() != list.Nil() {
			t.Fail()
		}
		if list.Nil().NReverse() != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.List(1, 2, 3).Reverse(), list.List(3, 2, 1)) {
			t.Fail()
		}
		if !list.Equal(list.List(1, 2, 3).NReverse(), list.List(3, 2, 1)) {
			t.Fail()
		}
		result, last := list.List(3, 2, 1).ReverseLast()
		last.Cdr = 4
		if !list.Equal(result, list.Cons(1, 2, 3, 4)) {
			t.Fail()
		}
		result, last = list.List(3, 2, 1).NReverseLast()
		last.Cdr = 4
		if !list.Equal(result, list.Cons(1, 2, 3, 4)) {
			t.Fail()
		}
		if !list.Equal(list.List(3, 2, 1).AppendReverse(list.List(4, 5, 6)), list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
		if !list.Equal(list.List(3, 2, 1).NAppendReverse(list.List(4, 5, 6)), list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
	})
	t.Run("Zip", func(t *testing.T) {
		list.PairForEach(func(ps ...*list.Pair) {
			if list.Car(ps[0]) != list.Caar(ps[1]) {
				t.Fail()
			}
			if ps[1].Car.(*list.Pair).Cdr.(*list.Pair) != nil {
				t.Fail()
			}
		}, list.List(1, 2, 3), list.Zip(list.List(1, 2, 3)))
		list.PairForEach(func(ps ...*list.Pair) {
			if !list.Equal(list.Car(ps[0]), list.Car(ps[1])) {
				t.Fail()
			}
		}, list.Zip(list.List("one", "two", "three"), list.List(1, 2, 3), list.Circular(true, false)),
			list.List(list.List("one", 1, true), list.List("two", 2, false), list.List("three", 3, true)))
	})
	t.Run("Unzip", func(t *testing.T) {
		lists := list.Unzip(2, list.List(1, 2, 3))
		if !list.Equal(lists[0], list.List(1)) &&
			!list.Equal(lists[1], list.List(2)) {
			t.Fail()
		}
		lists = list.Unzip(3, list.List(1, "one", "a"), list.List(2, "two", "b"), list.List(3, "three", "c"))
		if !list.Equal(lists[0], list.List(1, 2, 3)) &&
			!list.Equal(lists[1], list.List("one", "two", "three")) &&
			!list.Equal(lists[2], list.List("a", "b", "c")) {
			t.Fail()
		}
	})
	t.Run("Count", func(t *testing.T) {
		if list.Nil().Count(func(x interface{}) bool { return true }) != 0 {
			t.Fail()
		}
		if list.List(1, 2, 3, 4, 5).Count(func(x interface{}) bool { return x.(int)%2 == 0 }) != 2 {
			t.Fail()
		}
		if list.Count(func(xs ...interface{}) bool { return xs[0].(int) < xs[1].(int) }, list.List(1, 2, 3, 4, 5), list.List(3, 2, 4, 3, 2)) != 2 {
			t.Fail()
		}
	})
}

func TestFold(t *testing.T) {
	t.Run("Fold", func(t *testing.T) {
		if list.List(1, 2, 3, 4, 5).Fold(func(x, y interface{}) interface{} { return x.(int) + y.(int) }, 0) != 15 {
			t.Fail()
		}
		if !list.Equal(list.List(1, 2, 3, 4, 5).Fold(func(t, x interface{}) interface{} { return list.Cons(x, t) }, list.List(42)), list.List(1, 2, 3, 4, 5).AppendReverse(list.List(42))) {
			t.Fail()
		}
		if list.List(1, "2", 3, "4", 5).Fold(func(count, x interface{}) interface{} {
			if _, ok := x.(string); ok {
				return count.(int) + 1
			}
			return count
		}, 0) != 2 {
			t.Fail()
		}
		if !list.Equal(list.Fold(func(t interface{}, xs ...interface{}) interface{} { return list.Cons(xs[0], xs[1], t) }, list.Nil(), list.List("a", "b", "c"), list.List(1, 2, 3, 4, 5)), list.List("c", 3, "b", 2, "a", 1)) {
			t.Fail()
		}
	})
	t.Run("FoldRight", func(t *testing.T) {
		if !list.Equal(list.List(1, 2, 3, 4, 5).FoldRight(func(t, x interface{}) interface{} { return list.Cons(x, t) }, list.Nil()), list.List(1, 2, 3, 4, 5)) {
			t.Fail()
		}
		if !list.Equal(list.List(1, 2, 3, 4, 5).FoldRight(func(t, x interface{}) interface{} {
			if x.(int)%2 == 0 {
				return list.Cons(x, t)
			}
			return t
		}, list.Nil()), list.List(2, 4)) {
			t.Fail()
		}
		if !list.Equal(list.FoldRight(func(t interface{}, xs ...interface{}) interface{} { return list.Cons(xs[0], xs[1], t) }, list.Nil(), list.List("a", "b", "c"), list.List(1, 2, 3, 4, 5)), list.List("a", 1, "b", 2, "c", 3)) {
			t.Fail()
		}
	})
	t.Run("PairFold", func(t *testing.T) {
		if !list.Equal(list.List(1, 2, 3, 4, 5).PairFold(func(tail interface{}, pair *list.Pair) interface{} { pair.Cdr = tail; return pair }, list.Nil()), list.List(1, 2, 3, 4, 5).NReverse()) {
			t.Fail()
		}
	})
	t.Run("PairFoldRight", func(t *testing.T) {
		list.ForEach(func(ls ...interface{}) {
			if !list.Equal(ls[0], ls[1]) {
				t.Fail()
			}
		}, list.List(1, 2, 3).PairFoldRight(func(t interface{}, x *list.Pair) interface{} { return list.Cons(x, t) }, list.Nil()).(*list.Pair), list.List(list.List(1, 2, 3), list.List(2, 3), list.List(3)))
	})
	t.Run("Reduce", func(t *testing.T) {
		if list.List(2, 4, 9, 3, 8).Reduce(func(x, y interface{}) interface{} {
			if x.(int) > y.(int) {
				return x
			}
			return y
		}, 0) != 9 {
			t.Fail()
		}
	})
	t.Run("ReduceRight", func(t *testing.T) {
		if !list.Equal(list.List(list.List(1, 2, 3), list.List(4, 5, 6)).ReduceRight(func(t, x interface{}) interface{} { return list.Append(x.(*list.Pair), t.(*list.Pair)) }, list.Nil()), list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
	})
	t.Run("Unfold", func(t *testing.T) {
		if !list.Equal(list.Unfold(
			func(x interface{}) bool { return x.(int) > 10 },
			func(x interface{}) interface{} { return x.(int) * x.(int) },
			func(x interface{}) interface{} { return x.(int) + 1 },
			1, nil), list.List(1, 4, 9, 16, 25, 36, 49, 64, 81, 100)) {
			t.Fail()
		}
		if !list.Equal(list.Unfold(
			list.IsNilPair,
			list.Car,
			list.Cdr,
			list.List(1, 2, 3), nil), list.List(1, 2, 3)) {
			t.Fail()
		}
		if !list.Equal(list.Unfold(
			list.IsNilPair,
			list.Car,
			list.Cdr,
			list.List(1, 2, 3),
			func(interface{}) interface{} {
				return list.List(4, 5, 6)
			}), list.List(1, 2, 3, 4, 5, 6)) {
			t.Fail()
		}
	})
	t.Run("UnfoldRight", func(t *testing.T) {
		if !list.Equal(list.UnfoldRight(
			func(x interface{}) bool { return x.(int) == 0 },
			func(x interface{}) interface{} { return x.(int) * x.(int) },
			func(x interface{}) interface{} { return x.(int) - 1 },
			10, list.Nil()), list.List(1, 4, 9, 16, 25, 36, 49, 64, 81, 100)) {
			t.Fail()
		}
		if !list.Equal(list.UnfoldRight(
			list.IsNilPair,
			list.Car,
			list.Cdr,
			list.List(1, 2, 3), list.Nil()), list.List(1, 2, 3).Reverse()) {
			t.Fail()
		}
		if !list.Equal(list.UnfoldRight(
			list.IsNilPair,
			list.Car,
			list.Cdr,
			list.List(3, 2, 1),
			list.List(4, 5, 6)), list.List(3, 2, 1).AppendReverse(list.List(4, 5, 6))) {
			t.Fail()
		}
	})
	t.Run("Map", func(t *testing.T) {
		if !list.Equal(list.List(list.List("a", "b"), list.List("d", "e"), list.List("g", "h")).Map(list.Cadr), list.List("b", "e", "h")) {
			t.Fail()
		}
		if !list.Equal(list.Map(func(xs ...interface{}) interface{} { return xs[0].(int) + xs[1].(int) }, list.List(1, 2, 3), list.List(4, 5, 6)), list.List(5, 7, 9)) {
			t.Fail()
		}
		if !list.Equal(list.Map(func(xs ...interface{}) interface{} { return xs[0].(int) + xs[1].(int) }, list.List(3, 1, 4, 1), list.Circular(1, 0)), list.List(4, 1, 5, 1)) {
			t.Fail()
		}
		if !list.Equal(list.List(list.List("a", "b"), list.List("d", "e"), list.List("g", "h")).NMap(list.Cadr), list.List("b", "e", "h")) {
			t.Fail()
		}
		if !list.Equal(list.NMap(func(xs ...interface{}) interface{} { return xs[0].(int) + xs[1].(int) }, list.List(1, 2, 3), list.List(4, 5, 6)), list.List(5, 7, 9)) {
			t.Fail()
		}
		if !list.Equal(list.NMap(func(xs ...interface{}) interface{} { return xs[0].(int) + xs[1].(int) }, list.List(3, 1, 4, 1), list.Circular(1, 0)), list.List(4, 1, 5, 1)) {
			t.Fail()
		}
	})
	t.Run("ForEach", func(t *testing.T) {
		var v [5]int
		list.List(0, 1, 2, 3, 4).ForEach(func(x interface{}) {
			i := x.(int)
			v[i] = i * i
		})
		if v != [...]int{0, 1, 4, 9, 16} {
			t.Fail()
		}
		list.ForEach(func(xs ...interface{}) {
			v[xs[0].(int)] = xs[1].(int)
		}, list.List(0, 1, 2, 3, 4), list.List(3, 1, 4, 1, 5))
		if v != [...]int{3, 1, 4, 1, 5} {
			t.Fail()
		}
	})
	t.Run("AppendMap", func(t *testing.T) {
		if !list.Equal(list.List(1, 3, 8).AppendMap(func(x interface{}) *list.Pair { return list.List(x, -x.(int)) }), list.List(1, -1, 3, -3, 8, -8)) {
			t.Fail()
		}
		if !list.Equal(list.List(1, 3, 8).NAppendMap(func(x interface{}) *list.Pair { return list.List(x, -x.(int)) }), list.List(1, -1, 3, -3, 8, -8)) {
			t.Fail()
		}
		if !list.Equal(list.AppendMap(func(xs ...interface{}) *list.Pair { return list.List(xs[0], xs[1]) }, list.List(1, 3, 8), list.List(4, 7, 9)), list.List(1, 4, 3, 7, 8, 9)) {
			t.Fail()
		}
		if !list.Equal(list.NAppendMap(func(xs ...interface{}) *list.Pair { return list.List(xs[0], xs[1]) }, list.List(1, 3, 8), list.List(4, 7, 9)), list.List(1, 4, 3, 7, 8, 9)) {
			t.Fail()
		}
	})
	t.Run("PairForEach", func(t *testing.T) {
		var v []*list.Pair
		list.List(1, 2, 3).PairForEach(func(pair *list.Pair) {
			v = append(v, pair)
		})
		for i, x := range v {
			if !list.Equal(x, list.Tabulate(3-i, func(j int) interface{} { return i + j + 1 })) {
				t.Fail()
			}
		}
		v = v[:0]
		list.PairForEach(func(pairs ...*list.Pair) {
			v = append(v, list.Append(pairs[0], pairs[1]))
		}, list.List(1, 2, 3), list.List(4, 5, 6))
		for i, x := range v {
			if !list.Equal(x, list.Append(
				list.Tabulate(3-i, func(j int) interface{} { return i + j + 1 }),
				list.Tabulate(3-i, func(j int) interface{} { return i + j + 4 }))) {
				t.Fail()
			}
		}
	})
	t.Run("FilterMap", func(t *testing.T) {
		if !list.Equal(list.List("a", 1, "b", 3, "c", 7).FilterMap(func(x interface{}) (interface{}, bool) {
			if i, ok := x.(int); ok {
				return i * i, true
			}
			return nil, false
		}), list.List(1, 9, 49)) {
			t.Fail()
		}
		if !list.Equal(list.FilterMap(func(xs ...interface{}) (interface{}, bool) {
			i, ok := xs[0].(int)
			if !ok {
				return nil, false
			}
			j, ok := xs[1].(int)
			if !ok {
				return nil, false
			}
			return i + j, true
		}, list.List(1, 2, 3, 4, 5, 6), list.List("a", 1, "b", 3, "c", 7)), list.List(3, 7, 13)) {
			t.Fail()
		}
		if !list.Equal(list.List("a", 1, "b", 3, "c", 7).NFilterMap(func(x interface{}) (interface{}, bool) {
			if i, ok := x.(int); ok {
				return i * i, true
			}
			return nil, false
		}), list.List(1, 9, 49)) {
			t.Fail()
		}
		if !list.Equal(list.NFilterMap(func(xs ...interface{}) (interface{}, bool) {
			i, ok := xs[0].(int)
			if !ok {
				return nil, false
			}
			j, ok := xs[1].(int)
			if !ok {
				return nil, false
			}
			return i + j, true
		}, list.List(1, 2, 3, 4, 5, 6), list.List("a", 1, "b", 3, "c", 7)), list.List(3, 7, 13)) {
			t.Fail()
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("Filter", func(t *testing.T) {
		if !list.Equal(list.List(0, 7, 8, 8, 43, -4).Filter(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(0, 8, 8, -4)) {
			t.Fail()
		}
		if !list.Equal(list.List(0, 7, 8, 8, 43, -4).NFilter(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(0, 8, 8, -4)) {
			t.Fail()
		}
	})
	t.Run("Partition", func(t *testing.T) {
		if in, out := list.List("one", 2, 3, "four", "five", 6).Partition(func(x interface{}) bool { _, ok := x.(string); return ok }); !list.Equal(in, list.List("one", "four", "five")) || !list.Equal(out, list.List(2, 3, 6)) {
			t.Fail()
		}
		if in, out := list.List("one", 2, 3, "four", "five", 6).NPartition(func(x interface{}) bool { _, ok := x.(string); return ok }); !list.Equal(in, list.List("one", "four", "five")) || !list.Equal(out, list.List(2, 3, 6)) {
			t.Fail()
		}
	})
	t.Run("Remove", func(t *testing.T) {
		if !list.Equal(list.List(0, 7, 8, 8, 43, -4).Remove(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(7, 43)) {
			t.Fail()
		}
		if !list.Equal(list.List(0, 7, 8, 8, 43, -4).NRemove(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(7, 43)) {
			t.Fail()
		}
	})
}

func TestSearch(t *testing.T) {
	t.Run("Find", func(t *testing.T) {
		if x, ok := list.List(3, 1, 4, 1, 5, 9).Find(func(x interface{}) bool { return x.(int)%2 == 0 }); !ok || x != 4 {
			t.Fail()
		}
		if _, ok := list.List(3, 1, 1, 5, 9).Find(func(x interface{}) bool { return x.(int)%2 == 0 }); ok {
			t.Fail()
		}
	})
	t.Run("FindTail", func(t *testing.T) {
		if !list.Equal(list.List(3, 1, 37, -8, -5, 0, 0).FindTail(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(-8, -5, 0, 0)) {
			t.Fail()
		}
		if !list.Equal(list.List(3, 1, 37, -5).FindTail(func(x interface{}) bool { return x.(int)%2 == 0 }), list.Nil()) {
			t.Fail()
		}
	})
	t.Run("TakeWhile and DropWhile", func(t *testing.T) {
		if !list.Equal(list.List(2, 18, 3, 10, 22, 9).TakeWhile(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(2, 18)) {
			t.Fail()
		}
		if !list.Equal(list.List(2, 18, 3, 10, 22, 9).NTakeWhile(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(2, 18)) {
			t.Fail()
		}
		if !list.Equal(list.List(2, 18, 3, 10, 22, 9).DropWhile(func(x interface{}) bool { return x.(int)%2 == 0 }), list.List(3, 10, 22, 9)) {
			t.Fail()
		}
	})
	t.Run("Span and Break", func(t *testing.T) {
		if prefix, suffix := list.List(2, 18, 3, 10, 22, 9).Span(func(x interface{}) bool { return x.(int)%2 == 0 }); !list.Equal(prefix, list.List(2, 18)) || !list.Equal(suffix, list.List(3, 10, 22, 9)) {
			t.Fail()
		}
		if prefix, suffix := list.List(2, 18, 3, 10, 22, 9).NSpan(func(x interface{}) bool { return x.(int)%2 == 0 }); !list.Equal(prefix, list.List(2, 18)) || !list.Equal(suffix, list.List(3, 10, 22, 9)) {
			t.Fail()
		}
		if prefix, suffix := list.List(3, 1, 4, 1, 5, 9).Break(func(x interface{}) bool { return x.(int)%2 == 0 }); !list.Equal(prefix, list.List(3, 1)) || !list.Equal(suffix, list.List(4, 1, 5, 9)) {
			t.Fail()
		}
		if prefix, suffix := list.List(3, 1, 4, 1, 5, 9).NBreak(func(x interface{}) bool { return x.(int)%2 == 0 }); !list.Equal(prefix, list.List(3, 1)) || !list.Equal(suffix, list.List(4, 1, 5, 9)) {
			t.Fail()
		}
	})
	t.Run("Any", func(t *testing.T) {
		if !list.List("a", 3, "b", 2.7).Any(func(x interface{}) bool { _, ok := x.(int); return ok }) {
			t.Fail()
		}
		if list.List("a", 3.1, "b", 2.7).Any(func(x interface{}) bool { _, ok := x.(int); return ok }) {
			t.Fail()
		}
		if !list.Any(func(xs ...interface{}) bool { return xs[0].(int) < xs[1].(int) }, list.List(3, 1, 4, 1, 5), list.List(2, 7, 1, 8, 2)) {
			t.Fail()
		}
	})
	t.Run("Every", func(t *testing.T) {
		if list.List("a", 3, "b", 2.7).Every(func(x interface{}) bool { _, ok := x.(int); return ok }) {
			t.Fail()
		}
		if !list.List(1, 3, 4, 7).Every(func(x interface{}) bool { _, ok := x.(int); return ok }) {
			t.Fail()
		}
		if list.Every(func(xs ...interface{}) bool { return xs[0].(int) < xs[1].(int) }, list.List(3, 1, 4, 1, 5), list.List(2, 7, 1, 8, 2)) {
			t.Fail()
		}
	})
	t.Run("Index", func(t *testing.T) {
		if list.List(3, 1, 4, 1, 5, 9).Index(func(x interface{}) bool { return x.(int)%2 == 0 }) != 2 {
			t.Fail()
		}
		if list.Index(func(xs ...interface{}) bool { return xs[0].(int) < xs[1].(int) }, list.List(3, 1, 4, 1, 5, 9, 2, 5, 6), list.List(2, 7, 1, 8, 2)) != 1 {
			t.Fail()
		}
		if list.Index(func(xs ...interface{}) bool { return xs[0].(int) == xs[1].(int) }, list.List(3, 1, 4, 1, 5, 9, 2, 5, 6), list.List(2, 7, 1, 8, 2)) != -1 {
			t.Fail()
		}
	})
	t.Run("Member", func(t *testing.T) {
		if l := list.List("a", "b", "c"); l.Member("a") != l {
			t.Fail()
		}
		if l := list.List("a", "b", "c"); l.Member("b") != list.Cdr(l) {
			t.Fail()
		}
		if list.List("b", "c", "d").Member("a") != nil {
			t.Fail()
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("Delete", func(t *testing.T) {
		l := list.List(3, 1, 4, 1, 5)
		if !list.Equal(l.Delete(1), list.List(3, 4, 5)) {
			t.Fail()
		}
		if !list.Equal(l, list.List(3, 1, 4, 1, 5)) {
			t.Fail()
		}
		if !list.Equal(l.NDelete(1), list.List(3, 4, 5)) {
			t.Fail()
		}
		if !list.Equal(l, list.List(3, 4, 5)) {
			t.Fail()
		}
	})
	t.Run("DeleteDuplicates", func(t *testing.T) {
		l := list.List("a", "b", "a", "c", "a", "b", "c", "z")
		if !list.Equal(l.DeleteDuplicates(), list.List("a", "b", "c", "z")) {
			t.Fail()
		}
		if !list.Equal(l, list.List("a", "b", "a", "c", "a", "b", "c", "z")) {
			t.Fail()
		}
		if !list.Equal(l.NDeleteDuplicates(), list.List("a", "b", "c", "z")) {
			t.Fail()
		}
		if !list.Equal(l, list.List("a", "b", "c", "z")) {
			t.Fail()
		}
	})
}

func TestAssociationLists(t *testing.T) {
	t.Run("Assoc", func(t *testing.T) {
		e := list.List(list.List("a", 1), list.List("b", 2), list.List("c", 3))
		if l, ok := e.Assoc("a"); !ok || !list.Equal(l, list.List("a", 1)) {
			t.Fail()
		}
		if l, ok := e.Assoc("b"); !ok || !list.Equal(l, list.List("b", 2)) {
			t.Fail()
		}
		if l, ok := e.Assoc("d"); ok || l != nil {
			t.Fail()
		}
	})
}

func TestSets(t *testing.T) {
	t.Run("SetLessThanEqual", func(t *testing.T) {
		if !list.SetLessThanEqual(list.List("a"), list.List("a", "b", "a"), list.List("a", "b", "c", "c")) {
			t.Fail()
		}
		if !list.SetLessThanEqual() {
			t.Fail()
		}
		if !list.SetLessThanEqual(list.List("a")) {
			t.Fail()
		}
	})
	t.Run("SetEqual", func(t *testing.T) {
		if !list.SetEqual(list.List("b", "e", "a"), list.List("a", "e", "b"), list.List("e", "e", "b", "a")) {
			t.Fail()
		}
		if !list.SetEqual() {
			t.Fail()
		}
		if !list.SetEqual(list.List("a")) {
			t.Fail()
		}
	})
	t.Run("Adjoin", func(t *testing.T) {
		if !list.SetEqual(list.List("a", "b", "c", "d", "c", "e").Adjoin("a", "e", "i", "o", "u"), list.List("u", "o", "i", "a", "b", "c", "d", "c", "e")) {
			t.Fail()
		}
	})
	t.Run("SetUnion", func(t *testing.T) {
		if !list.Equal(list.SetUnion(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("u", "o", "i", "a", "b", "c", "d", "e")) {
			t.Fail()
		}
		if !list.Equal(list.SetUnion(list.List("a", "a", "c"), list.List("x", "a", "x")), list.List("x", "a", "a", "c")) {
			t.Fail()
		}
		if list.SetUnion() != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.SetUnion(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
		if !list.Equal(list.NSetUnion(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("u", "o", "i", "a", "b", "c", "d", "e")) {
			t.Fail()
		}
		if !list.Equal(list.NSetUnion(list.List("a", "a", "c"), list.List("x", "a", "x")), list.List("x", "a", "a", "c")) {
			t.Fail()
		}
		if list.NSetUnion() != list.Nil() {
			t.Fail()
		}
		if !list.Equal(list.NSetUnion(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
	})
	t.Run("SetIntersection", func(t *testing.T) {
		if !list.Equal(list.SetIntersection(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("a", "e")) {
			t.Fail()
		}
		if !list.Equal(list.SetIntersection(list.List("a", "x", "y", "a"), list.List("x", "a", "x", "z")), list.List("a", "x", "a")) {
			t.Fail()
		}
		if !list.Equal(list.SetIntersection(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
		if !list.Equal(list.NSetIntersection(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("a", "e")) {
			t.Fail()
		}
		if !list.Equal(list.NSetIntersection(list.List("a", "x", "y", "a"), list.List("x", "a", "x", "z")), list.List("a", "x", "a")) {
			t.Fail()
		}
		if !list.Equal(list.NSetIntersection(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
	})
	t.Run("SetDifference", func(t *testing.T) {
		if !list.Equal(list.SetDifference(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("b", "c", "d")) {
			t.Fail()
		}
		if !list.Equal(list.SetDifference(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
		if !list.Equal(list.NSetDifference(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("b", "c", "d")) {
			t.Fail()
		}
		if !list.Equal(list.NSetDifference(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
	})
	t.Run("SetXor", func(t *testing.T) {
		if !list.SetEqual(list.SetXor(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("d", "c", "b", "i", "o", "u")) {
			t.Fail()
		}
		if list.SetXor() != nil {
			t.Fail()
		}
		if !list.Equal(list.SetXor(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
		if !list.SetEqual(list.NSetXor(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u")), list.List("d", "c", "b", "i", "o", "u")) {
			t.Fail()
		}
		if list.NSetXor() != nil {
			t.Fail()
		}
		if !list.Equal(list.NSetXor(list.List("a", "b", "c")), list.List("a", "b", "c")) {
			t.Fail()
		}
	})
	t.Run("SetDifferenceAndIntersection", func(t *testing.T) {
		difference, intersection := list.SetDifferenceAndIntersection(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u"))
		if !list.Equal(difference, list.List("b", "c", "d")) {
			t.Fail()
		}
		if !list.Equal(intersection, list.List("a", "e")) {
			t.Fail()
		}
		difference, intersection = list.SetDifferenceAndIntersection(list.List("a", "b", "c"))
		if !list.Equal(difference, list.List("a", "b", "c")) {
			t.Fail()
		}
		if intersection != nil {
			t.Fail()
		}
		difference, intersection = list.NSetDifferenceAndIntersection(list.List("a", "b", "c", "d", "e"), list.List("a", "e", "i", "o", "u"))
		if !list.Equal(difference, list.List("b", "c", "d")) {
			t.Fail()
		}
		if !list.Equal(intersection, list.List("a", "e")) {
			t.Fail()
		}
		difference, intersection = list.NSetDifferenceAndIntersection(list.List("a", "b", "c"))
		if !list.Equal(difference, list.List("a", "b", "c")) {
			t.Fail()
		}
		if intersection != nil {
			t.Fail()
		}
	})
}
