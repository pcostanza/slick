// Package list implements a list library for Go that mimicks functionality that
// is typically found in Lisp and Scheme languages which is based on pairs or cons cells.
//
// To a large extent, his package follows the SRFI 1 specification for Scheme authored by Olin Shivers,
// which is part of the Scheme Requests For Implementations library.
// See https://srfi.schemers.org/srfi-1/srfi-1.html for SRFI 1, and https://srfi.schemers.org
// for Scheme Requests For Implementation.
//
// Here are some, but not all details in which package deviates from SRFI 1:
//
// * Go accepts variable number of parameters in function calls through slices. This package does not attempt
// to use lists for this purpose instead, so there is sometimes (but hopefully not very often) a
// need to convert between lists and slices. This package provides utility functions to support this:
// AppendToSlice, ToSlice, and FromSlice.
//
// * Go does not allow for special characters in identifiers. That's why we cannot adopt
// Scheme's naming convention for linear-update ("destructive") functions whose names
// end in an exclamation mark (!). We therefore adopt Common Lisp's naming convention
// where such functions start with "N" (as in "Non-consing") instead, such as in
// NAppend, NDelete, NReverse, and so on.
//
// * Several functions exist both as single list-parameter variants, which are implemented
// as methods on *Pair, and as multiple list-parameter variants, which are implemented as
// standalone functions. This makes it easier to handle the more common single list-parameter
// cases more efficiently, and use a somewhat more pleasant syntax.
//
// * In Scheme, several functions return either some result value or Scheme's false value (#f)
// to indicate there is no result. Since Go is statically typed without support for choice
// types, we instead most of the time opted for returning multiple values, one for the primary
// return value, and a second boolean value indicating success or failure.
//
// * Cons corresponds to Scheme's cons* or list*. Use NewPair for constructing new pairs.
//
// * A proper list must end in Nil(). If it ends in nil (of type interface{}), then it's
// considered a dotted list. A dotted list is a non-circular list that does not end in Nil(),
// but in anything other than a value of type *Pair.
//
// * The selectors Car, Cdr, Caar, Cddr, etc., are "liberal" in that they do not panic if the
// argument is not a proper list. They are more similar to Common Lisp's corresponding selectors
// rather than Scheme's in this regard. If you need stricter checks, just use the Car and Cdr fields
// directly.
//
// * In the various folding functions, intermediate results are passed first, not last.
//
// * Iota is not supported due to a lack of generic number operations in Go.
//
// * Concatenate is not supported because it addresses a very Scheme-specific issue only.
//
// General discussion:
//
// Linear-update ("destructive") functions may alter and recycle cons cells from the argument list.
// They are allowed to, but not required to, but the implementation attempts to actually recycle the
// argument lists.
//
// List-filtering functions such as Filter or Delete do not disorder lists. Elements appear in the
// answer list in the same order as they appear in the argument list.
//
// Because lists are an inherently sequential data structure (unlike, say, slices), list-inspection
// functions such as Find, FindTail, Any, and Every commit to a left-to-right traversal order of their
// argument list.
//
// However, constructor functions, such as Tabulate, and many mapping functions do not specify the dynamic
// order in which their functional argument is applied to its various values.
//
// "Linear update" functions
//
// Functionality is provided both in "pure" and "linear-update" (potentially destructive) forms whenever this
// makes sense. A "pure" function has no side-effects, and in particular does not alter its arguments in any way.
// A "linear update" function is allowed -- but not required -- to side-effect its arguments in order to construct
// its result. "Linear update" functions are typically given names starting with "N". So, for example,
// NAppend(list1, list2) is allowed to construct its result by simply assigning the Cdr of the last pair of
// list1 to point to list2 and then returning list1 (unless list1 is the empty list, in which case it would
// simply return list2).
//
// What this means is that you may only apply linear-update functions to values that you know are
// "dead" -- values that will never be used again in your program. This must be so, since you can't rely on
// the value passed to a linear-update function after that function has been called. It might be
// unchanged; it might be altered.
//
// The "linear" in "linear update" doesn't mean "linear time" or "linear space" or any sort of multiple-of-n
// kind of meaning. It's a fancy term that type theorists and pure functional programmers use to describe
// systems where you are only allowed to have exactly one reference to each variable. This provides a
// guarantee that the value bound to a variable is bound to no other variable. So when you use a variable
// in a variable reference, you "use it up." Knowing that no one else has a pointer to that value means
// a system primitive is free to side-effect its arguments to produce what is, observationally, a pure-functional
// result.
//
// In the context of this library, "linear update" means you, the programmer, know there are no other live
// references to the value passed to the function -- after passing the value to one of these functions,
// the value of the old pointer is indeterminate. Basically, you are licensing these functions to alter
// the data structure if it feels like it -- you have declared you don't care either way.
//
// You get no help from Go or this library in checking that the values you claim are "linear" really are. So
// you better get it right. Or play it safe and use the non-N functions -- it doesn't do any good to compute
// quickly if you get the wrong answer.
package list
