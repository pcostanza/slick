package lib

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Symbol struct {
	Package, Identifier string
}

func (sym *Symbol) String() string {
	switch sym.Package {
	case "":
		return sym.Identifier
	case "_keyword":
		return ":" + sym.Identifier
	default:
		return sym.Package + ":" + sym.Identifier
	}
}

var symbols sync.Map

func Intern(pkg, ident string) *Symbol {
	sym := Symbol{pkg, ident}
	actual, _ := symbols.LoadOrStore(sym, &sym)
	return actual.(*Symbol)
}

var gensymCounter int64

func Gensym(prefix string) *Symbol {
	ncounter := atomic.AddInt64(&gensymCounter, 1)
	if prefix == "" {
		return Intern("", fmt.Sprintf("_g%v", ncounter))
	}
	if prefix[0] == '_' {
		return Intern("", fmt.Sprintf("%v%v", prefix, ncounter))
	}
	return Intern("", fmt.Sprintf("_%v%v", prefix, ncounter))
}
