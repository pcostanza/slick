package compiler

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"strconv"
	"strings"
	"unicode"

	"github.com/exascience/slick/lib"
	"github.com/exascience/slick/list"
	"github.com/exascience/slick/reader"
)

type (
	compiler struct {
		reader *reader.Reader
		header []byte
	}

	Environment struct {
	}

	macro = func(form *list.Pair, env Environment) (newForm interface{}, err error)
)

func (cmp *compiler) init(rd *reader.Reader) {
	cmp.reader = rd
}

var slickPath, slickPlugins, slickRoot, libPlugin string

func init() {
	slickPath = os.Getenv("SLICKPATH")
	var err error
	if slickPath == "" {
		slickPath, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		slickPath = filepath.Join(slickPath, "slick")
	}
	slickPlugins = filepath.Join(slickPath, "plugins")
	slickRoot = os.Getenv("SLICKROOT")
	if slickRoot == "" {
		slickRoot, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		slickRoot = filepath.Join(slickRoot, "slick")
	}
	libPlugin = filepath.Join(slickRoot, "plugins", "plugin.so")
}

func (cmp *compiler) resolvePlugin(path string) *plugin.Plugin {
	if path[0] == '#' {
		path = path[1:]
	}
	fullPath := filepath.Join(slickPlugins, path, "slick/plugin.so")
	p, err := plugin.Open(fullPath)
	if err != nil {
		panic(err)
	}
	return p
}

func (cmp *compiler) resolveLibPlugin() *plugin.Plugin {
	p, err := plugin.Open(libPlugin)
	if err != nil {
		panic(err)
	}
	return p
}

func (cmp *compiler) encloseSymbol(sym *lib.Symbol) *lib.Symbol {
	nsym, enclosed := cmp.reader.EncloseSymbol(sym)
	if enclosed {
		cmp.header = append(cmp.header, "import "...)
		cmp.header = append(cmp.header, nsym.Package...)
		cmp.header = append(cmp.header, " \""...)
		cmp.header = append(cmp.header, sym.Package...)
		cmp.header = append(cmp.header, "\"\n"...)
	}
	return nsym
}

type bailout struct{}

func (cmp *compiler) error(form *list.Pair, msg string) {
	pos, _ := cmp.reader.FormPos(form)
	epos := cmp.reader.File().Position(pos)
	n := len(cmp.reader.Errors)
	if n > 0 && cmp.reader.Errors[n-1].Pos.Line == epos.Line {
		return
	}
	if n > 10 {
		panic(bailout{})
	}

	cmp.reader.Errors.Add(epos, msg)
}

func in(key *lib.Symbol, keys []*lib.Symbol) bool {
	for _, skey := range keys {
		if key == skey {
			return true
		}
	}
	return false
}

func (cmp *compiler) checkf(outer, form *list.Pair, keys ...*lib.Symbol) {
	for form != list.Nil() {
		if !in(form.Car.(*lib.Symbol), keys) {
			cmp.error(outer, fmt.Sprintf("invalid key %v, must be one of %v", form.Car, keys))
		}
		form = form.Cdr.(*list.Pair).Cdr.(*list.Pair)
	}
}

func getf(form *list.Pair, key *lib.Symbol) (interface{}, bool) {
	for form != list.Nil() {
		if form.Car == key {
			return form.Cdr.(*list.Pair).Car, true
		}
		form = form.Cdr.(*list.Pair).Cdr.(*list.Pair)
	}
	return nil, false
}

func isValidGoIdentifier(lit string) bool {
	if len(lit) == 0 {
		return false
	}
	for i, r := range lit {
		if !unicode.IsLetter(r) && r != '_' && (i == 0 || !unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func isValidSimpleIdentifier(sym *lib.Symbol) bool {
	return sym.Package == "" && isValidGoIdentifier(sym.Identifier)
}

func isValidQualifiedIdentifier(sym *lib.Symbol) bool {
	return isValidGoIdentifier(sym.Package) && sym.Package != "_" &&
		isValidGoIdentifier(sym.Identifier) && sym.Identifier != "_"
}

func isValidIdentifier(sym *lib.Symbol) bool {
	if sym.Package == "" {
		return isValidGoIdentifier(sym.Identifier)
	}
	return isValidQualifiedIdentifier(sym)
}

var (
	_break          = lib.Intern("", "break")
	_chan           = lib.Intern("", "chan")
	_chan_right     = lib.Intern("", "chan<-")
	_chan_left      = lib.Intern("", "<-chan")
	_const          = lib.Intern("", "const")
	_continue       = lib.Intern("", "continue")
	_defer          = lib.Intern("", "defer")
	_fallthrough    = lib.Intern("", "fallthrough")
	_for            = lib.Intern("", "for")
	_while          = lib.Intern("", "while")
	_loop           = lib.Intern("", "loop")
	_range          = lib.Intern("", "range")
	_func           = lib.Intern("", "func")
	_go             = lib.Intern("", "go")
	_goto           = lib.Intern("", "goto")
	_if             = lib.Intern("", "if")
	_ifStar         = lib.Intern("", "if*")
	_import         = lib.Intern("", "import")
	_interface      = lib.Intern("", "interface")
	_map            = lib.Intern("", "map")
	_package        = lib.Intern("", "package")
	_return         = lib.Intern("", "return")
	_select         = lib.Intern("", "select")
	_struct         = lib.Intern("", "struct")
	_switch         = lib.Intern("", "switch")
	_switchStar     = lib.Intern("", "switch*")
	_typeSwitch     = lib.Intern("", "type-switch")
	_typeSwitchStar = lib.Intern("", "type-switch*")
	_default        = lib.Intern("", "default")
	_type           = lib.Intern("", "type")
	_var            = lib.Intern("", "var")
	_array          = lib.Intern("", "array")
	_begin          = lib.Intern("", "begin")
	_declare        = lib.Intern("", "declare")
	_ellipsis       = lib.Intern("", "...")
	_use            = lib.Intern("", "use")
	_ptr            = lib.Intern("", "*")
	_slice          = lib.Intern("", "slice")
	_type_alias     = lib.Intern("", "type-alias")
	_values         = lib.Intern("", "values")

	_make        = lib.Intern("", "make")
	_make_struct = lib.Intern("", "make-struct")
	_make_array  = lib.Intern("", "make-array")
	_make_slice  = lib.Intern("", "make-slice")
	_make_map    = lib.Intern("", "make-map")
	_slot        = lib.Intern("", "slot")
	_at          = lib.Intern("", "at")
	_assert      = lib.Intern("", "assert")
	_convert     = lib.Intern("", "convert")

	_and_equal     = lib.Intern("", "&=")
	_and_not_equal = lib.Intern("", "&^=")
	_arrow_right   = lib.Intern("", "->")
	_colon_equal   = lib.Intern("_keyword", "=")
	_div_equal     = lib.Intern("", "/=")
	_equal         = lib.Intern("", "=")
	_lshift_equal  = lib.Intern("", "<<=")
	_minus_equal   = lib.Intern("", "-=")
	_minus_minus   = lib.Intern("", "--")
	_mul_equal     = lib.Intern("", "*=")
	_or_equal      = lib.Intern("", "|=")
	_plus_equal    = lib.Intern("", "+=")
	_plus_plus     = lib.Intern("", "++")
	_rem_equal     = lib.Intern("", "%=")
	_rshift_equal  = lib.Intern("", ">>=")
	_xor_equal     = lib.Intern("", "^=")

	_plus          = lib.Intern("", "+")
	_minus         = lib.Intern("", "-")
	_or            = lib.Intern("", "|")
	_bang          = lib.Intern("", "!")
	_xor           = lib.Intern("", "^")
	_mul           = lib.Intern("", "*")
	_div           = lib.Intern("", "/")
	_rem           = lib.Intern("", "%")
	_shl           = lib.Intern("", "<<")
	_shr           = lib.Intern("", ">>")
	_and           = lib.Intern("", "&")
	_address       = lib.Intern("", "&")
	_and_not       = lib.Intern("", "&^")
	_arrow_left    = lib.Intern("", "<-")
	_bool_and      = lib.Intern("", "&&")
	_bool_or       = lib.Intern("", "||")
	_equal_equal   = lib.Intern("", "==")
	_not_equal     = lib.Intern("", "!=")
	_less          = lib.Intern("", "<")
	_less_equal    = lib.Intern("", "<=")
	_greater       = lib.Intern("", ">")
	_greater_equal = lib.Intern("", ">=")

	_quote            = lib.Intern("", "quote")
	_quasiquote       = lib.Intern("", "quasiquote")
	_unquote          = lib.Intern("", "unquote")
	_unquote_splicing = lib.Intern("", "unquote-splicing")
	_splice           = lib.Intern("", "splice")
)

var (
	keyDocumentation = lib.Intern("_keyword", "documentation")
	keyEqual         = lib.Intern("_keyword", "=")
	keyTag           = lib.Intern("_keyword", "tag")
	keyType          = lib.Intern("_keyword", "type")
)

func formatComment(result []byte, comment string) []byte {
	for _, line := range strings.Split(comment, "\n") {
		result = append(result, '/', '/', ' ')
		result = append(result, strings.TrimSpace(line)...)
		result = append(result, '\n')
	}
	return result
}

func formatIdentifier(result []byte, ident *lib.Symbol) []byte {
	if ident.Package != "" {
		result = append(result, ident.Package...)
		result = append(result, '.')
	}
	return append(result, ident.Identifier...)
}

func (cmp *compiler) compilePackageClause() (result []byte) {
	rd := cmp.reader
	form, ok := rd.Read().(*list.Pair)
	if !ok {
		cmp.reader.Error(0, "package clause is not a list")
	}
	pkgClause := form.ToSlice()
	if len(pkgClause) < 2 || len(pkgClause) > 3 {
		cmp.error(form, "package clause has invalid length")
	}
	if pkgClause[0] != _package {
		cmp.error(form, "package clause starts with invalid keyword")
	}
	sym, ok := pkgClause[1].(*lib.Symbol)
	if !ok {
		cmp.error(form, "package name is not an identifier")
	}
	if !isValidSimpleIdentifier(sym) || sym.Identifier == "_" {
		cmp.error(form, "invalid package name")
	}
	if len(pkgClause) == 3 {
		if comment, ok := pkgClause[2].(string); !ok {
			cmp.error(form, "package comment is not a string")
		} else {
			result = formatComment(result, comment)
		}
	}
	result = append(result, "package "...)
	result = append(result, sym.Identifier...)
	result = append(result, '\n', '\n')
	return
}

func (cmp *compiler) compileGenDecl(result []byte, keyword string, allowLeadComment bool, form *list.Pair, f func(element interface{}) (string, []byte)) []byte {
	cdr := form.Cdr.(*list.Pair)
	leadComment := false

	oneDecl := func() []byte {
		comment, decl := f(cdr.Car)
		if len(decl) == 0 {
			return result
		}
		if comment != "" {
			if leadComment {
				result = append(result, '/', '/', '\n')
			}
			result = formatComment(result, comment)
		}
		result = append(result, keyword...)
		result = append(result, ' ')
		result = append(result, decl...)
		return append(result, '\n', '\n')
	}

	if cdr.Cdr == list.Nil() {
		return oneDecl()
	}

	if allowLeadComment {
		if comment, ok := cdr.Car.(string); ok {
			result = formatComment(result, comment)
			leadComment = true
			cdr = cdr.Cdr.(*list.Pair)
			if cdr.Cdr == list.Nil() {
				return oneDecl()
			}
		}
	}

	result = append(result, keyword...)
	result = append(result, ' ', '(', '\n')
	cdr.ForEach(func(element interface{}) {
		comment, decl := f(element)
		if len(decl) == 0 {
			return
		}
		if comment != "" {
			result = formatComment(result, comment)
		}
		result = append(result, decl...)
	})
	return append(result, ')', '\n', '\n')
}

func isValidImport(lit string) bool {
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	for _, r := range lit {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return false
		}
	}
	return lit != ""
}

func (cmp *compiler) compileImportDecl(form *list.Pair) {
	cmp.header = cmp.compileGenDecl(cmp.header, "import", false, form, func(element interface{}) (comment string, decl []byte) {
		if inner, ok := element.(string); ok {
			pkg := path.Base(inner)
			if _, ok := cmp.reader.PackageToPath[pkg]; ok {
				cmp.error(form, "ambiguous import")
			}
			cmp.reader.PackageToPath[pkg] = inner
			cmp.reader.PathToPackage[inner] = pkg
			decl = append(decl, '"')
			decl = append(decl, inner...)
			decl = append(decl, '"', '\n')
			return
		}
		inner, ok := element.(*list.Pair)
		if !ok {
			cmp.error(form, "invalid import clause")
		}
		imp := inner.ToSlice()
		if len(imp) < 2 || len(imp) > 3 {
			cmp.error(inner, "import clause has invalid length")
		}
		var quoted bool
		if imp[0] == _quote {
			quoted = true
			if len(imp) != 2 {
				cmp.error(inner, "invalid quoted import")
			}
			inner, ok = imp[1].(*list.Pair)
			if !ok {
				cmp.error(form, "invalid quoted import")
			}
			imp = inner.ToSlice()
			if len(imp) < 2 || len(imp) > 3 {
				cmp.error(inner, "quoted import clause has invalid length")
			}
		}
		ident, ok := imp[0].(*lib.Symbol)
		if !ok {
			cmp.error(inner, "import name is not an identifier")
		}
		if !isValidSimpleIdentifier(ident) {
			cmp.error(inner, "invalid import identifier")
		}
		importName := ident.Identifier
		path, ok := imp[1].(string)
		if !ok {
			cmp.error(inner, "import path is not a string")
		}
		if !isValidImport(path) {
			cmp.error(inner, "invalid import path: "+path)
		}
		if len(imp) == 3 {
			if comment, ok = imp[2].(string); !ok {
				cmp.error(inner, "import comment is not a string")
			}
		}
		if importName != "_" {
			if _, ok := cmp.reader.PackageToPath[importName]; ok {
				cmp.error(form, "ambiguous import")
			}
			cmp.reader.PackageToPath[importName] = path
			if !quoted {
				cmp.reader.PathToPackage[path] = importName
			}
		}
		if !quoted {
			decl = append(decl, importName...)
			decl = append(decl, ' ', '"')
			decl = append(decl, path...)
			decl = append(decl, '"', '\n')
		}
		return
	})
}

func (cmp *compiler) compileUseDecl(form *list.Pair) {
	cmp.header = cmp.compileGenDecl(cmp.header, "use", false, form, func(element interface{}) (_ string, _ []byte) {
		if inner, ok := element.(string); ok {
			pkg := path.Base(inner)
			if _, ok := cmp.reader.PackageToPath[pkg]; ok {
				cmp.error(form, "ambiguous use declaration")
			}
			cmp.resolvePlugin(inner)
			cmp.reader.PackageToPath[pkg] = "#" + inner
			return
		}
		inner, ok := element.(*list.Pair)
		if !ok {
			cmp.error(form, "invalid use clause")
		}
		imp := inner.ToSlice()
		if len(imp) < 2 || len(imp) > 3 {
			cmp.error(inner, "use clause has invalid length")
		}
		var quoted bool
		if imp[0] == _quote {
			quoted = true
			if len(imp) != 2 {
				cmp.error(inner, "invalid quoted use declaration")
			}
			inner, ok = imp[1].(*list.Pair)
			if !ok {
				cmp.error(form, "invalid quoted use declaration")
			}
			imp = inner.ToSlice()
			if len(imp) < 2 || len(imp) > 3 {
				cmp.error(inner, "quoted use clause has invalid length")
			}
		}
		ident, ok := imp[0].(*lib.Symbol)
		if !ok {
			cmp.error(inner, "plugin name is not an identifier")
		}
		if !isValidSimpleIdentifier(ident) {
			cmp.error(inner, "invalid plugin identifier")
		}
		pluginName := ident.Identifier
		path, ok := imp[1].(string)
		if !ok {
			cmp.error(inner, "plugin path is not a string")
		}
		if !isValidImport(path) {
			cmp.error(inner, "invalid plugin path: "+path)
		}
		if len(imp) == 3 {
			if _, ok := imp[2].(string); !ok {
				cmp.error(inner, "plugin comment is not a string")
			}
		}
		if pluginName != "_" {
			if _, ok := cmp.reader.PackageToPath[pluginName]; ok {
				cmp.error(form, "ambiguous use declaration")
			}
			cmp.reader.PackageToPath[pluginName] = "#" + path
			if !quoted {
				cmp.resolvePlugin(path)
			}
		}
		return
	})
}

func (cmp *compiler) compileValueSpec(form *list.Pair) func(element interface{}) (string, []byte) {
	iota := 0
	return func(element interface{}) (comment string, decl []byte) {
		defer func() { iota++ }()
		switch e := element.(type) {
		case *list.Pair:
			var syms []*lib.Symbol
			switch s := e.Car.(type) {
			case *list.Pair:
				syms = s.AppendToSlice(syms).([]*lib.Symbol)
			case *lib.Symbol:
				syms = []*lib.Symbol{s}
			default:
				cmp.error(e, "invalid identifier(s)")
			}
			for _, ident := range syms {
				if !isValidSimpleIdentifier(ident) {
					cmp.error(e, fmt.Sprintf("invalid identifier %v", ident.Identifier))
				}
			}

			rest := e.Cdr.(*list.Pair)

			cmp.checkf(e, rest, keyType, keyEqual, keyDocumentation)

			typForm, typ := getf(rest, keyType)
			valForm, val := getf(rest, keyEqual)
			docForm, doc := getf(rest, keyDocumentation)

			switch form.Car {
			case _var:
				if !typ && !val {
					cmp.error(e, "missing variable type or initialization")
				}
			case _const:
				if !val && (iota == 0 || typ) {
					cmp.error(e, "missing constant value")
				}
			}

			decl = append(decl, syms[0].Identifier...)
			for _, ident := range syms[1:] {
				decl = append(decl, ',', ' ')
				decl = append(decl, ident.Identifier...)
			}

			if typ {
				decl = append(decl, ' ')
				decl = cmp.compileType(decl, e, typForm)
			}

			if val {
				decl = append(decl, ' ', '=', ' ')
				decl = cmp.compileExpression(decl, e, valForm)
			}

			if doc {
				var ok bool
				if comment, ok = docForm.(string); !ok {
					cmp.error(e, "comment is not a string")
				}
			}

		case *lib.Symbol:
			if !isValidSimpleIdentifier(e) {
				cmp.error(form, fmt.Sprintf("invalid identifier %v", e.Identifier))
			}
			switch form.Car {
			case _var:
				cmp.error(form, "missing variable type or initialization")
			case _const:
				if iota == 0 {
					cmp.error(form, "missing constant value")
				}
			}
			decl = append(decl, e.Identifier...)

		default:
			cmp.error(form, fmt.Sprintf("invalid declaration %v", element))
		}

		decl = append(decl, '\n')

		return
	}
}

func (cmp *compiler) compileTypeSpec(form *list.Pair, alias bool) func(element interface{}) (string, []byte) {
	return func(element interface{}) (comment string, decl []byte) {
		inner, ok := element.(*list.Pair)
		if !ok {
			cmp.error(form, "invalid type spec")
		}
		spec := inner.ToSlice()
		if len(spec) < 2 || len(spec) > 3 {
			cmp.error(inner, "type spec has invalid length")
		}
		ident, ok := spec[0].(*lib.Symbol)
		if !ok {
			cmp.error(inner, "invalid identifier")
		}
		if !isValidSimpleIdentifier(ident) {
			cmp.error(inner, fmt.Sprintf("invalid identifier %v", ident.Identifier))
		}
		decl = append(decl, ident.Identifier...)
		if alias {
			decl = append(decl, ' ', '=', ' ')
		} else {
			decl = append(decl, ' ')
		}
		if comment, ok = spec[1].(string); ok {
			if len(spec) < 3 {
				cmp.error(inner, "type spec has invalid length")
			}
			decl = cmp.compileType(decl, inner, spec[2])
		} else {
			if len(spec) > 2 {
				cmp.error(inner, "type spec has invalid length")
			}
			decl = cmp.compileType(decl, inner, spec[1])
		}
		return
	}
}

func (cmp *compiler) compileParameters(result []byte, form *list.Pair, ellipsisOk bool) []byte {
	if form == list.Nil() {
		return append(result, '(', ')')
	}
	result = append(result, '(')
	outer := form
	for {
		entryForm, ok := form.Car.(*list.Pair)
		form = form.Cdr.(*list.Pair)
		if !ok {
			cmp.error(outer, "invalid parameter list entry")
			continue
		}
		entry := entryForm.ToSlice()
		if (ellipsisOk && (len(entry) < 2 || len(entry) > 3)) || len(entry) != 2 {
			cmp.error(entryForm, "invalid parameter declaration length")
		}
		var names []*lib.Symbol
		switch e := entry[0].(type) {
		case *lib.Symbol:
			names = []*lib.Symbol{e}
		case *list.Pair:
			names = e.AppendToSlice(names).([]*lib.Symbol)
		}
		if len(names) == 0 {
			cmp.error(entryForm, fmt.Sprintf("invalid parameter names %v", entry[0]))
		}
		for _, name := range names {
			if !isValidSimpleIdentifier(name) {
				cmp.error(entryForm, fmt.Sprintf("invalid identifier %v", name))
			}
		}
		result = append(result, names[0].Identifier...)
		for _, name := range names[1:] {
			result = append(result, ',', ' ')
			result = append(result, name.Identifier...)
		}
		result = append(result, ' ')
		var typeForm interface{}
		if ellipsisOk && entry[1] == _ellipsis {
			if len(entry) != 3 {
				cmp.error(entryForm, "invalid parameter type")
				continue
			}
			if form != list.Nil() {
				cmp.error(entryForm, "variadic parameter is not the final entry in parameter list")
			}
			result = append(result, '.', '.', '.')
			typeForm = entry[2]
		}
		if len(entry) == 2 {
			typeForm = entry[1]
		}
		result = cmp.compileType(result, entryForm, typeForm)
		if form == list.Nil() {
			break
		}
		result = append(result, ',', ' ')
	}
	return append(result, ')')
}

func (cmp *compiler) compileFuncDecl(result []byte, form *list.Pair) []byte {
	head := []byte("func ")

	rest := form.Cdr.(*list.Pair)
	if first, ok := rest.Car.(*list.Pair); ok {
		head = cmp.compileParameters(head, first, false)
		head = append(head, ' ')
		rest = rest.Cdr.(*list.Pair)
	}

	ident, ok := rest.Car.(*lib.Symbol)
	if !ok {
		cmp.error(form, "function name is not an identifier")
	}
	if !isValidSimpleIdentifier(ident) || ident.Identifier == "_" {
		cmp.error(form, "invalid function name")
	}
	head = append(head, ident.Identifier...)
	head = append(head, ' ')
	rest = rest.Cdr.(*list.Pair)

	if rest == list.Nil() {
		result = append(result, head...)
		return append(result, '(', ')', '\n', '\n')
	}

	first, ok := rest.Car.(*list.Pair)
	if !ok {
		cmp.error(form, "missing parameter list in function declaration")
	} else {
		head = cmp.compileParameters(head, first, true)
		head = append(head, ' ')
		rest = rest.Cdr.(*list.Pair)
	}

	if rest == list.Nil() {
		result = append(result, head...)
		return append(result, '\n', '\n')
	}

	first, ok = rest.Car.(*list.Pair)
	if !ok {
		cmp.error(form, "missing result list in function declaration")
	} else if first != list.Nil() {
		head = cmp.compileParameters(head, first, false)
		head = append(head, ' ')
		rest = rest.Cdr.(*list.Pair)
	}

	if rest == list.Nil() {
		result = append(result, head...)
		return append(result, '\n', '\n')
	}

	if comment, ok := rest.Car.(string); ok {
		result = formatComment(result, comment)
		rest = rest.Cdr.(*list.Pair)
	}

	result = append(result, head...)

	if rest == list.Nil() {
		return append(result, '\n', '\n')
	}

	result = cmp.compileBlock(result, form, rest)
	return append(result, '\n')
}

func (cmp *compiler) compilePragma(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) != 2 {
		cmp.error(form, "declare form has invalid length")
	}
	declString, ok := decl[1].(string)
	if !ok {
		cmp.error(form, "declaration in declare form is not a string")
		return result
	}
	declString = strings.TrimSpace(declString)
	if len(declString) == 0 {
		cmp.error(form, "declaration in declare form is empty")
		return result
	}
	result = append(result, '\n', '/', '/')
	result = append(result, declString...)
	return append(result, '\n', '\n')
}

func (cmp *compiler) compileDecl(result []byte, form *list.Pair) []byte {
	var f func(element interface{}) (string, []byte)
	var keyword string
	for {
		switch form.Car {
		case _splice:
			block := form.ToSlice()
			for _, blockEntry := range block[1:] {
				result = cmp.compileDecl(result, blockEntry.(*list.Pair))
			}
			return result

		case _const:
			f = cmp.compileValueSpec(form)
			keyword = "const"

		case _var:
			f = cmp.compileValueSpec(form)
			keyword = "var"

		case _type:
			f = cmp.compileTypeSpec(form, false)
			keyword = "type"

		case _type_alias:
			f = cmp.compileTypeSpec(form, true)
			keyword = "type"

		case _func:
			return cmp.compileFuncDecl(result, form)

		case _declare:
			return cmp.compilePragma(result, form)

		default:
			if sym, ok := form.Car.(*lib.Symbol); ok {
				if len(sym.Package) > 0 && sym.Package[0] == '#' {
					p := cmp.resolvePlugin(sym.Package)
					macroSym, err := p.Lookup(sym.Identifier)
					if err != nil {
						cmp.error(form, "invalid macro invocation")
						return result
					}
					newForm, err := macroSym.(macro)(form, Environment{})
					if err != nil {
						cmp.error(form, fmt.Sprintf("error during macroexpansion: %v", err))
						return result
					}
					form = newForm.(*list.Pair)
					continue
				}
			}
			cmp.error(form, "invalid declaration")
			return result
		}

		return cmp.compileGenDecl(result, keyword, true, form, f)
	}
}

func (cmp *compiler) compileArrayType(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) != 3 {
		cmp.error(form, "invalid array type declaration")
	}
	result = append(result, '[')
	if decl[1] == _ellipsis {
		result = append(result, '.', '.', '.')
	} else {
		result = cmp.compileExpression(result, form, decl[1])
	}
	result = append(result, ']')
	return cmp.compileType(result, form, decl[2])
}

func (cmp *compiler) compileStructType(result []byte, form *list.Pair) []byte {
	rest := form.Cdr.(*list.Pair)
	if rest == list.Nil() {
		return append(result, "struct{}"...)
	}
	result = append(result, "struct{\n"...)
	rest.ForEach(func(element interface{}) {
		eForm, ok := element.(*list.Pair)
		if !ok {
			cmp.error(form, fmt.Sprintf("invalid struct type entry %v", element))
			return
		}
		rest := eForm.Cdr.(*list.Pair)
		cmp.checkf(eForm, rest, keyDocumentation, keyTag)
		docForm, doc := getf(rest, keyDocumentation)
		typForm, typ := getf(rest, keyType)
		tagForm, tag := getf(rest, keyTag)
		if doc {
			result = formatComment(result, docForm.(string))
		}
		if typ {
			var names []*lib.Symbol
			switch e := eForm.Car.(type) {
			case *lib.Symbol:
				names = []*lib.Symbol{e}
			case *list.Pair:
				names = e.AppendToSlice(names).([]*lib.Symbol)
			}
			if len(names) == 0 {
				cmp.error(eForm, fmt.Sprintf("invalid identifiers %v", form.Car))
			}
			for _, name := range names {
				if !isValidSimpleIdentifier(name) {
					cmp.error(eForm, fmt.Sprintf("invalid identifier %v", name))
				}
			}
			result = append(result, names[0].Identifier...)
			for _, name := range names[1:] {
				result = append(result, ',', ' ')
				result = append(result, name.Identifier...)
			}
			result = append(result, ' ')
			result = cmp.compileType(result, form, typForm)
		} else {
			result = cmp.compileType(result, form, form.Car)
		}
		if tag {
			if tag, ok := tagForm.(string); !ok {
				cmp.error(eForm, fmt.Sprintf("tag for struct field is not a string %v", tagForm))
			} else {
				result = append(result, ' ')
				result = append(result, fmt.Sprintf("%#q", tag)...)
			}
		}
		result = append(result, '\n')
	})
	return append(result, '}')
}

func (cmp *compiler) compilePointerType(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) != 2 {
		cmp.error(form, "invalid pointer type declaration")
	}
	result = append(result, '*')
	return cmp.compileType(result, form, decl[1])
}

func (cmp *compiler) compileFuncType(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) < 1 || len(decl) > 3 {
		cmp.error(form, "invalid function type declaration")
	}
	result = append(result, "func "...)
	if len(decl) == 1 {
		return append(result, '(', ')')
	}
	result = cmp.compileParameters(result, decl[1].(*list.Pair), true)
	if len(decl) == 3 && decl[2] != list.Nil() {
		result = append(result, ' ')
		result = cmp.compileParameters(result, decl[2].(*list.Pair), false)
	}
	return result
}

func (cmp *compiler) compileInterfaceType(result []byte, form *list.Pair) []byte {
	rest := form.Cdr.(*list.Pair)
	if rest == list.Nil() {
		return append(result, "interface{}"...)
	}
	result = append(result, "interface{\n"...)
	rest.ForEach(func(element interface{}) {
		switch e := element.(type) {
		case *lib.Symbol:
			sym := cmp.encloseSymbol(e)
			if !isValidQualifiedIdentifier(sym) {
				cmp.error(form, fmt.Sprintf("invalid identifier %v", sym))
				return
			}
			result = formatIdentifier(result, sym)
			result = append(result, '\n')
		case *list.Pair:
			spec := e.ToSlice()
			if len(spec) < 1 || len(spec) > 4 {
				cmp.error(e, fmt.Sprintf("invalid interface type entry %v", element))
				return
			}
			if spec[1] == keyDocumentation {
				if len(spec) != 3 {
					cmp.error(e, fmt.Sprintf("invalid interface type entry %v", element))
					return
				}
				ident, ok := spec[0].(*lib.Symbol)
				sym := cmp.encloseSymbol(ident)
				if !ok || !isValidQualifiedIdentifier(sym) {
					cmp.error(e, fmt.Sprintf("invalid identifier %v", sym))
					return
				}
				result = formatComment(result, spec[2].(string))
				result = formatIdentifier(result, sym)
				return
			}
			if len(spec) == 4 {
				result = formatComment(result, spec[3].(string))
			}
			if name, ok := spec[0].(*lib.Symbol); !ok || !isValidSimpleIdentifier(name) || name.Identifier == "_" {
				cmp.error(e, fmt.Sprintf("invalid interface type entry name %v", spec[0]))
			} else {
				result = formatIdentifier(result, name)
				result = append(result, ' ')
			}
			if len(spec) == 1 {
				result = append(result, '(', ')', '\n')
				break
			}
			result = cmp.compileParameters(result, spec[1].(*list.Pair), true)
			if len(spec) >= 3 && spec[2] != list.Nil() {
				result = append(result, ' ')
				result = cmp.compileParameters(result, spec[2].(*list.Pair), false)
			}
			result = append(result, '\n')
		default:
			cmp.error(form, fmt.Sprintf("invalid interface type entry %v", element))
		}
	})
	return append(result, '}')
}

func (cmp *compiler) compileSliceType(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) != 2 {
		cmp.error(form, "invalid slice type declaration")
	}
	result = append(result, '[', ']')
	return cmp.compileType(result, form, decl[1])
}

func (cmp *compiler) compileMapType(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) != 3 {
		cmp.error(form, "invalid map type declaration")
	}
	result = append(result, "map["...)
	result = cmp.compileType(result, form, decl[1])
	result = append(result, ']')
	return cmp.compileType(result, form, decl[2])
}

func (cmp *compiler) compileChannelType(result []byte, form *list.Pair) []byte {
	decl := form.ToSlice()
	if len(decl) != 2 {
		cmp.error(form, "invalid channel type declaration")
	}
	result = append(result, decl[0].(*lib.Symbol).Identifier...)
	result = append(result, ' ')
	return cmp.compileType(result, form, decl[1])
}

func (cmp *compiler) compileType(result []byte, outer *list.Pair, form interface{}) []byte {
	switch typeForm := form.(type) {
	case *lib.Symbol:
		sym := cmp.encloseSymbol(typeForm)
		if !isValidIdentifier(sym) {
			cmp.error(outer, fmt.Sprintf("invalid identifier %v", sym))
			return result
		}
		return formatIdentifier(result, sym)
	case *list.Pair:
		switch typeForm.Car {
		case _array:
			return cmp.compileArrayType(result, typeForm)
		case _struct:
			return cmp.compileStructType(result, typeForm)
		case _ptr:
			return cmp.compilePointerType(result, typeForm)
		case _func:
			return cmp.compileFuncType(result, typeForm)
		case _interface:
			return cmp.compileInterfaceType(result, typeForm)
		case _slice:
			return cmp.compileSliceType(result, typeForm)
		case _map:
			return cmp.compileMapType(result, typeForm)
		case _chan, _chan_right, _chan_left:
			return cmp.compileChannelType(result, typeForm)
		default:
			cmp.error(typeForm, "unknown type keyword")
			return result
		}
	default:
		cmp.error(outer, "invalid type declaration")
		return result
	}
}

func (cmp *compiler) compileBlock(result []byte, outer, form *list.Pair) []byte {
	if form == list.Nil() {
		return append(result, '{', '}', ' ')
	}
	result = append(result, '{', '\n')
	form.ForEach(func(element interface{}) {
		result = cmp.compileStatement(result, outer, element, false)
	})
	return append(result, '}', '\n')
}

func (cmp *compiler) compileSimpleStatement(result []byte, form *list.Pair) []byte {
	if form == nil {
		return append(result, '\n')
	}
	slice := form.ToSlice()
	switch slice[0] {
	case _arrow_right:
		if len(slice) != 3 {
			cmp.error(form, "invalid channel send statement")
			return result
		}
		result = cmp.compileExpression(result, form, slice[1])
		result = append(result, ' ', '<', '-', ' ')
		result = cmp.compileExpression(result, form, slice[2])
	case _plus_plus, _minus_minus:
		if len(slice) != 2 {
			cmp.error(form, "invalid inc/dec statement")
			return result
		}
		result = cmp.compileExpression(result, form, slice[1])
		result = append(result, slice[0].(*lib.Symbol).Identifier...)
	case _equal, _plus_equal, _minus_equal, _or_equal, _xor_equal, _mul_equal, _div_equal, _rem_equal,
		_lshift_equal, _rshift_equal, _and_equal, _and_not_equal:
		if len(slice) != 3 {
			cmp.error(form, "invalid assignment statement")
		}
		result = cmp.compileExpression(result, form, slice[1])
		result = append(result, ' ')
		result = append(result, slice[0].(*lib.Symbol).Identifier...)
		result = append(result, ' ')
		result = cmp.compileExpression(result, form, slice[2])
	case _colon_equal:
		if len(slice) != 3 {
			cmp.error(form, "invalid short variable definition")
		}
		var names []*lib.Symbol
		switch e := slice[1].(type) {
		case *lib.Symbol:
			names = []*lib.Symbol{e}
		case *list.Pair:
			names = e.AppendToSlice(names).([]*lib.Symbol)
		}
		if len(names) == 0 {
			cmp.error(form, fmt.Sprintf("invalid identifiers %v", slice[1]))
			break
		}
		for _, name := range names {
			if !isValidSimpleIdentifier(name) {
				cmp.error(form, fmt.Sprintf("invalid identifier %v", name))
			}
		}
		result = append(result, names[0].Identifier...)
		for _, name := range names[1:] {
			result = append(result, ',', ' ')
			result = append(result, name.Identifier...)
		}
		result = append(result, ' ', ':', '=', ' ')
		result = cmp.compileExpression(result, form, slice[2])
	default:
		result = cmp.compileExpression(result, form, form)
	}
	return append(result, '\n')
}

func (cmp *compiler) compileImplicitBlock(result []byte, outer, form *list.Pair) []byte {
	if form == list.Nil() {
		return append(result, '\n')
	}
	form.ForEach(func(element interface{}) {
		result = cmp.compileStatement(result, outer, element, false)
	})
	return result
}

func (cmp *compiler) compileSelectStatement(result []byte, form *list.Pair) []byte {
	result = append(result, "select {\n"...)
	var defaultSeen bool
	form.Cdr.(*list.Pair).ForEach(func(element interface{}) {
		clause := element.(*list.Pair)
		if clause.Car == _default {
			if defaultSeen {
				cmp.error(form, "multiple default cases")
			}
			defaultSeen = true
			result = append(result, "default:\n"...)
		} else {
			head := clause.Car.(*list.Pair)
			switch head.Car {
			case _arrow_right, _arrow_left, _equal, _colon_equal:
			default:
				cmp.error(form, "invalid select statement")
			}
			result = cmp.compileSimpleStatement(result, head)
			result = append(result, ':', '\n')
		}
		result = cmp.compileImplicitBlock(result, form, clause.Cdr.(*list.Pair))
	})
	return append(result, '}', '\n')
}

func (cmp *compiler) compileSwitchStatement(result []byte, form *list.Pair, star bool) []byte {
	rest := form.Cdr.(*list.Pair)
	result = append(result, "switch "...)
	if star {
		result = cmp.compileSimpleStatement(result, rest.Car.(*list.Pair))
		if result[len(result)-1] != '\n' {
			result = append(result, ';', ' ')
		}
		rest = rest.Cdr.(*list.Pair)
	}
	result = cmp.compileExpression(result, form, rest.Car)
	result = append(result, ' ', '{', '\n')
	var defaultSeen bool
	rest.Cdr.(*list.Pair).ForEach(func(element interface{}) {
		clause := element.(*list.Pair)
		if clause.Car == _default {
			if defaultSeen {
				cmp.error(form, "multiple default cases")
			}
			defaultSeen = true
			result = append(result, "default:\n"...)
		} else {
			result = append(result, "case "...)
			switch head := clause.Car.(type) {
			default:
				result = cmp.compileExpression(result, form, head)
			case *list.Pair:
				result = cmp.compileExpression(result, form, head.Car)
				head.Cdr.(*list.Pair).ForEach(func(element interface{}) {
					result = append(result, ',', ' ')
					result = cmp.compileExpression(result, form, element)
				})
			}
			result = append(result, ':', '\n')
		}
		result = cmp.compileImplicitBlock(result, form, clause.Cdr.(*list.Pair))
	})
	return append(result, '}', '\n')
}

func (cmp *compiler) compileTypeSwitchStatement(result []byte, form *list.Pair, star bool) []byte {
	rest := form.Cdr.(*list.Pair)
	result = append(result, "switch "...)
	if star {
		result = cmp.compileSimpleStatement(result, rest.Car.(*list.Pair))
		if result[len(result)-1] != '\n' {
			result = append(result, ';', ' ')
		}
		rest = rest.Cdr.(*list.Pair)
	}
	sym := rest.Car.(*lib.Symbol)
	if !isValidSimpleIdentifier(sym) {
		cmp.error(form, "invalid variable declaration")
	}
	if sym.Identifier != "_" {
		result = append(result, sym.Identifier...)
		result = append(result, ' ', ':', '=', ' ')
	}
	rest = rest.Cdr.(*list.Pair)
	result = cmp.compilePrimaryExpression(result, form, rest.Car)
	result = append(result, ".(type) {\n"...)
	var defaultSeen bool
	rest.Cdr.(*list.Pair).ForEach(func(element interface{}) {
		clause := element.(*list.Pair)
		if clause.Car == _default {
			if defaultSeen {
				cmp.error(form, "multiple default cases")
			}
			defaultSeen = true
			result = append(result, "default:\n"...)
		} else {
			result = append(result, "case "...)
			switch head := clause.Car.(type) {
			case *lib.Symbol:
				result = cmp.compileType(result, form, head)
			case *list.Pair:
				result = cmp.compileType(result, form, head.Car)
				head.Cdr.(*list.Pair).ForEach(func(element interface{}) {
					result = append(result, ',', ' ')
					result = cmp.compileType(result, form, element)
				})
			default:
				cmp.error(form, "invalid type-switch case")
			}
			result = append(result, ':', '\n')
		}
		result = cmp.compileImplicitBlock(result, form, clause.Cdr.(*list.Pair))
	})
	return append(result, '}', '\n')
}

func (cmp *compiler) compileForStatement(result []byte, form *list.Pair) []byte {
	rest := form.Cdr.(*list.Pair)
	result = append(result, "for "...)
	clause := rest.Car.(*list.Pair).ToSlice()
	if len(clause) > 3 {
		cmp.error(form, "invalid for statement")
	}
	if len(clause) == 0 {
		return cmp.compileBlock(result, form, rest.Cdr.(*list.Pair))
	}
	if len(clause) > 0 {
		if clause[0] != list.Nil() {
			result = cmp.compileSimpleStatement(result, clause[0].(*list.Pair))
		}
	}
	if result[len(result)-1] != '\n' {
		result = append(result, ';', ' ')
	}
	if len(clause) > 1 {
		if clause[1] != list.Nil() {
			result = cmp.compileExpression(result, form, clause[1])
		}
	}
	if result[len(result)-1] != '\n' {
		result = append(result, ';', ' ')
	}
	if len(clause) > 2 {
		if clause[2] != list.Nil() {
			result = cmp.compileSimpleStatement(result, clause[2].(*list.Pair))
		}
	}
	result = append(result, ' ')
	return cmp.compileBlock(result, form, rest.Cdr.(*list.Pair))
}

func (cmp *compiler) compileWhileStatement(result []byte, form *list.Pair) []byte {
	rest := form.Cdr.(*list.Pair)
	result = append(result, "for "...)
	result = cmp.compileExpression(result, form, rest.Car)
	result = append(result, ' ')
	return cmp.compileBlock(result, form, rest.Cdr.(*list.Pair))
}

func (cmp *compiler) compileLoopStatement(result []byte, form *list.Pair) []byte {
	result = append(result, "for "...)
	return cmp.compileBlock(result, form, form.Cdr.(*list.Pair))
}

func (cmp *compiler) compileRangeStatement(result []byte, form *list.Pair) []byte {
	rest := form.Cdr.(*list.Pair)
	result = append(result, "for "...)
	clause := rest.Car.(*list.Pair).ToSlice()
	if len(clause) != 3 {
		cmp.error(form, "invalid range statement")
	}
	switch clause[0] {
	case _colon_equal:
		var names []*lib.Symbol
		switch e := clause[1].(type) {
		case *lib.Symbol:
			names = []*lib.Symbol{e}
		case *list.Pair:
			names = e.AppendToSlice(names).([]*lib.Symbol)
		}
		if len(names) == 0 {
			cmp.error(form, fmt.Sprintf("invalid identifiers %v", clause[1]))
		}
		for _, name := range names {
			if !isValidSimpleIdentifier(name) {
				cmp.error(form, fmt.Sprintf("invalid identifier %v", name))
			}
		}
		result = append(result, names[0].Identifier...)
		for _, name := range names[1:] {
			result = append(result, ',', ' ')
			result = append(result, name.Identifier...)
		}
		result = append(result, " := range "...)
	case _equal:
		result = cmp.compileExpression(result, form, clause[1])
		result = append(result, " = range "...)
	default:
		cmp.error(form, "invalid range statement")
	}
	result = cmp.compileExpression(result, form, clause[2])
	return cmp.compileBlock(result, form, rest.Cdr.(*list.Pair))
}

func (cmp *compiler) compileStatement(result []byte, outer *list.Pair, stmt interface{}, atBlock bool) []byte {
	for {
		switch form := stmt.(type) {
		case *lib.Symbol:
			if form.Package == "_keyword" {
				if !isValidGoIdentifier(form.Identifier) || form.Identifier == "_" {
					cmp.error(outer, fmt.Sprintf("invalid label name %v", stmt))
				}
				result = append(result, form.Identifier...)
				return append(result, ':', '\n')
			}
			result = cmp.compileExpression(result, outer, stmt)
			return append(result, '\n')
		case *list.Pair:
			if form == nil {
				return cmp.compileSimpleStatement(result, form)
			}
			switch form.Car {
			case _const, _type, _type_alias, _var:
				return cmp.compileDecl(result, form)
			case _arrow_right, _plus_plus, _minus_minus, _equal, _plus_equal, _minus_equal, _or_equal, _xor_equal,
				_mul_equal, _div_equal, _rem_equal, _lshift_equal, _rshift_equal, _and_equal, _and_not_equal, _colon_equal:
				return cmp.compileSimpleStatement(result, form)
			case _go, _defer:
				return cmp.compileDelayedStatement(result, form)
			case _break, _continue, _goto:
				return cmp.compileJumpStatement(result, form)
			case _return:
				return cmp.compileReturnStatement(result, form)
			case _fallthrough:
				return cmp.compileFallthroughStatement(result, form)
			case _splice:
				return cmp.compileImplicitBlock(result, form, form.Cdr.(*list.Pair))
			case _begin:
				if atBlock {
					return cmp.compileImplicitBlock(result, form, form.Cdr.(*list.Pair))
				}
				return cmp.compileBlock(result, form, form.Cdr.(*list.Pair))
			case _if:
				return cmp.compileIfStatement(result, form)
			case _ifStar:
				return cmp.compileIfStarStatement(result, form)
			case _for:
				return cmp.compileForStatement(result, form)
			case _while:
				return cmp.compileWhileStatement(result, form)
			case _loop:
				return cmp.compileLoopStatement(result, form)
			case _range:
				return cmp.compileRangeStatement(result, form)
			case _switch:
				return cmp.compileSwitchStatement(result, form, false)
			case _switchStar:
				return cmp.compileSwitchStatement(result, form, true)
			case _typeSwitch:
				return cmp.compileTypeSwitchStatement(result, form, false)
			case _typeSwitchStar:
				return cmp.compileTypeSwitchStatement(result, form, true)
			case _select:
				return cmp.compileSelectStatement(result, form)
			default:
				if sym, ok := form.Car.(*lib.Symbol); ok {
					if len(sym.Package) > 0 && sym.Package[0] == '#' {
						p := cmp.resolvePlugin(sym.Package)
						if macroSym, err := p.Lookup(sym.Identifier); err != nil {
							cmp.error(outer, "invalid macro invocation")
						} else if newForm, err := macroSym.(macro)(form, Environment{}); err != nil {
							cmp.error(outer, fmt.Sprintf("error during macroexpansion: %v", err))
						} else {
							stmt = newForm
							continue
						}
					}
				}
				result = cmp.compileExpression(result, form, form)
				return append(result, '\n')
			}
		default:
			cmp.error(outer, fmt.Sprintf("invalid statement %v", stmt))
			return result
		}
	}
}

func (cmp *compiler) compileIfStatement(result []byte, form *list.Pair) []byte {
	stmt := form.ToSlice()
	if len(stmt) < 3 || len(stmt) > 4 {
		cmp.error(form, "invalid if statement")
	}
	result = append(result, "if "...)
	result = cmp.compileExpression(result, form, stmt[1])
	result = append(result, ' ', '{', '\n')
	result = cmp.compileStatement(result, form, stmt[2], true)
	if len(stmt) == 4 {
		result = append(result, "} else {\n"...)
		result = cmp.compileStatement(result, form, stmt[3], true)
	}
	return append(result, '}', '\n')
}

func (cmp *compiler) compileIfStarStatement(result []byte, form *list.Pair) []byte {
	stmt := form.ToSlice()
	if len(stmt) < 4 || len(stmt) > 5 {
		cmp.error(form, "invalid if* statement")
	}
	result = append(result, "if "...)
	result = cmp.compileSimpleStatement(result, stmt[1].(*list.Pair))
	if result[len(result)-1] != '\n' {
		result = append(result, ';', ' ')
	}
	result = cmp.compileExpression(result, form, stmt[2])
	result = append(result, ' ', '{', '\n')
	result = cmp.compileStatement(result, form, stmt[3], true)
	if len(stmt) == 5 {
		result = append(result, "} else {\n"...)
		result = cmp.compileStatement(result, form, stmt[4], true)
	}
	return append(result, '}', '\n')
}

func (cmp *compiler) compileFallthroughStatement(result []byte, form *list.Pair) []byte {
	if form.Cdr != list.Nil() {
		cmp.error(form, "invalid fallthrough statement")
	}
	return append(result, "fallthrough\n"...)
}

func (cmp *compiler) compileReturnStatement(result []byte, form *list.Pair) []byte {
	stmt := form.ToSlice()
	if len(stmt) > 2 {
		cmp.error(form, "invalid number of return values")
	}
	if len(stmt) == 0 {
		return append(result, "return\n"...)
	}
	result = append(result, "return "...)
	result = cmp.compileExpression(result, form, stmt[1])
	return append(result, '\n')
}

func (cmp *compiler) compileDelayedStatement(result []byte, form *list.Pair) []byte {
	del := form.ToSlice()
	if len(del) != 2 {
		cmp.error(form, fmt.Sprintf("invalid statement %v", form))
	}
	result = append(result, del[0].(*lib.Symbol).Identifier...)
	result = append(result, ' ')
	return cmp.compileExpression(result, form, del[1])
}

func (cmp *compiler) compileJumpStatement(result []byte, form *list.Pair) []byte {
	stmt := form.ToSlice()
	if form.Car == _goto {
		if len(stmt) != 2 {
			cmp.error(form, "invalid goto statement")
		}
	} else if len(stmt) > 2 {
		cmp.error(form, "invalid break/continue statement")
	}
	result = append(result, stmt[0].(*lib.Symbol).Identifier...)
	if len(stmt) == 2 {
		if label, ok := stmt[1].(*lib.Symbol); !ok || !isValidSimpleIdentifier(label) || label.Identifier == "_" {
			cmp.error(form, fmt.Sprintf("invalid jump target %v", stmt[1]))
		} else {
			result = append(result, ' ')
			result = append(result, label.Identifier...)
		}
	}
	return append(result, '\n')
}

func (cmp *compiler) compileStructLiteral(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) < 2 || len(expr)%2 == 1 {
		cmp.error(form, "invalid struct literal")
	}
	result = append(result, '(')
	result = cmp.compileType(result, form, expr[1])
	result = append(result, '{')
	for i := 2; i < len(expr); i += 2 {
		switch s := expr[i].(type) {
		case *lib.Symbol:
			if !isValidSimpleIdentifier(s) {
				cmp.error(form, fmt.Sprintf("invalid key %v in struct literal", s))
			}
			result = append(result, s.Identifier...)
		default:
			cmp.error(form, fmt.Sprintf("invalid key %v in struct literal", s))
		}
		result = append(result, ':', ' ')
		result = cmp.compileExpression(result, form, expr[i+1])
		result = append(result, ',', ' ')
	}
	return append(result, '}', ')')
}

func (cmp *compiler) compileVectorLiteral(result []byte, kind string, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) < 2 {
		cmp.error(form, fmt.Sprintf("invalid %v literal", kind))
	}
	result = append(result, '(')
	result = cmp.compileType(result, form, expr[1])
	result = append(result, '{')
	for i := 2; i < len(expr); i++ {
		result = cmp.compileExpression(result, form, expr[i])
		result = append(result, ',', ' ')
	}
	return append(result, '}', ')')
}

func (cmp *compiler) compileMapLiteral(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) < 2 || len(expr)%2 == 1 {
		cmp.error(form, "invalid map literal")
	}
	result = append(result, '(')
	result = cmp.compileType(result, form, expr[1])
	result = append(result, '{')
	for i := 2; i < len(expr); i += 2 {
		result = cmp.compileExpression(result, form, expr[i])
		result = append(result, ':', ' ')
		result = cmp.compileExpression(result, form, expr[i+1])
		result = append(result, ',', ' ')
	}
	return append(result, '}', ')')
}

func (cmp *compiler) compileFuncLiteral(result []byte, form *list.Pair) []byte {
	result = append(result, "func "...)

	rest := form.Cdr.(*list.Pair)

	if rest == list.Nil() {
		return append(result, '(', ')', ' ', '{', '}', ' ')
	}

	first, ok := rest.Car.(*list.Pair)
	if !ok {
		cmp.error(form, "missing parameter list in function literal")
	} else {
		result = cmp.compileParameters(result, first, true)
		result = append(result, ' ')
		rest = rest.Cdr.(*list.Pair)
	}

	if rest == list.Nil() {
		return append(result, '{', '}', ' ')
	}

	first, ok = rest.Car.(*list.Pair)
	if !ok {
		cmp.error(form, "missing result list in function literal")
	} else if first != list.Nil() {
		result = cmp.compileParameters(result, first, false)
		result = append(result, ' ')
		rest = rest.Cdr.(*list.Pair)
	}

	return cmp.compileBlock(result, form, rest)
}

func (cmp *compiler) compileSlotExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) != 3 {
		cmp.error(form, "invalid slot expression")
	}
	result = cmp.compileExpression(result, form, expr[1])
	result = append(result, '.')
	switch s := expr[2].(type) {
	case *lib.Symbol:
		if !isValidSimpleIdentifier(s) {
			cmp.error(form, fmt.Sprintf("invalid selector %v in slot expression", s))
		}
		return append(result, s.Identifier...)
	default:
		cmp.error(form, fmt.Sprintf("invalid selector %v in slot expression", s))
		return result
	}
}

func (cmp *compiler) compileIndexExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) != 3 {
		cmp.error(form, "invalid index expression")
	}
	result = cmp.compileExpression(result, form, expr[1])
	result = append(result, '[')
	result = cmp.compileExpression(result, form, expr[2])
	return append(result, ']')
}

func (cmp *compiler) compileSliceExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) < 3 || len(expr) > 5 {
		cmp.error(form, "invalid slice expression")
	}
	result = cmp.compileExpression(result, form, expr[1])
	result = append(result, '[')
	result = cmp.compileExpression(result, form, expr[2])
	result = append(result, ':')
	if len(expr) == 3 {
		return append(result, ']')
	}
	result = cmp.compileExpression(result, form, expr[3])
	if len(expr) == 4 {
		return result
	}
	result = append(result, ':')
	result = cmp.compileExpression(result, form, expr[4])
	return append(result, ']')
}

func (cmp *compiler) compileAssertExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) != 3 {
		cmp.error(form, "invalid type assertion")
	}
	result = cmp.compileExpression(result, form, expr[1])
	result = append(result, '.', '(')
	result = cmp.compileType(result, form, expr[2])
	return append(result, ')')
}

func (cmp *compiler) compileConvertExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) != 3 {
		cmp.error(form, "invalid type conversion")
	}
	result = append(result, '(')
	result = cmp.compileType(result, form, expr[2])
	result = append(result, ')', '(')
	result = cmp.compileExpression(result, form, expr[1])
	return append(result, ')')
}

func (cmp *compiler) compileOperatorExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) < 2 {
		cmp.error(form, "invalid operator expression")
	}
	if len(expr) == 2 {
		// unary expression
		switch expr[0] {
		case _plus, _minus, _bang, _xor, _ptr, _address, _arrow_left:
			result = append(result, expr[0].(*lib.Symbol).Identifier...)
			return cmp.compileExpression(result, form, expr[1])
		default:
			cmp.error(form, fmt.Sprintf("invalid operator %v in unary expression", expr[0]))
			return result
		}
	}
	switch expr[0] {
	case _plus, _minus, _mul, _div, _rem, _and, _and_not, _or, _xor, _shl, _shr, _bool_and, _bool_or:
		op := expr[0].(*lib.Symbol).Identifier
		result = append(result, '(')
		result = cmp.compileExpression(result, form, expr[1])
		for i := 2; i < len(expr); i++ {
			result = append(result, ' ')
			result = append(result, op...)
			result = append(result, ' ')
			result = cmp.compileExpression(result, form, expr[i])
		}
		return append(result, ')')
	case _equal_equal, _not_equal, _less, _less_equal, _greater, _greater_equal:
		if len(expr) != 3 {
			cmp.error(form, "invalid operator expression")
		}
		result = append(result, '(')
		result = cmp.compileExpression(result, form, expr[1])
		result = append(result, ' ')
		result = append(result, expr[0].(*lib.Symbol).Identifier...)
		result = append(result, ' ')
		result = cmp.compileExpression(result, form, expr[2])
		return append(result, ')')
	default:
		cmp.error(form, "invalid operator expression")
		return result
	}
}

func (cmp *compiler) compileCallExpression(result []byte, form *list.Pair) []byte {
	expr := form.ToSlice()
	if len(expr) == 0 {
		cmp.error(form, "invalid call expression")
	}
	result = cmp.compileExpression(result, form, expr[0])
	if len(expr) == 1 {
		return append(result, '(', ')')
	}
	result = append(result, '(')
	result = cmp.compileExpression(result, form, expr[1])
	if len(expr) > 2 {
		lastIndex := len(expr) - 1
		for i := 2; i < lastIndex; i++ {
			result = append(result, ',', ' ')
			result = cmp.compileExpression(result, form, expr[i])
		}
		if expr[lastIndex] == _ellipsis {
			result = append(result, '.', '.', '.')
		} else {
			result = append(result, ',', ' ')
			result = cmp.compileExpression(result, form, expr[lastIndex])
		}
	}
	if l := len(result) - 1; result[l] == '\n' {
		result = append(result[:l], ',', '\n')
	}
	return append(result, ')')
}

func (cmp *compiler) compileMakeExpression(result []byte, form *list.Pair) []byte {
	result = append(result, "make("...)
	rest := form.Cdr.(*list.Pair)
	result = cmp.compileType(result, form, rest.Car)
	rest.Cdr.(*list.Pair).ForEach(func(element interface{}) {
		result = append(result, ',', ' ')
		result = cmp.compileExpression(result, form, element)
	})
	return append(result, ')')
}

func (cmp *compiler) compileExpr(result []byte, form *list.Pair, element interface{}, operatorAllowed bool) []byte {
	for {
		switch e := element.(type) {
		case *lib.Symbol:
			sym := cmp.encloseSymbol(e)
			if !isValidIdentifier(sym) {
				cmp.error(form, fmt.Sprintf("Invalid identifier %v.", sym))
			}
			return formatIdentifier(result, sym)
		case *big.Int:
			return append(result, e.String()...)
		case float64:
			return strconv.AppendFloat(result, e, 'g', -1, 64)
		case complex128:
			result = append(result, '(')
			result = strconv.AppendFloat(result, real(e), 'g', -1, 64)
			result = append(result, ' ', '+', ' ')
			result = strconv.AppendFloat(result, imag(e), 'g', -1, 64)
			return append(result, 'i', ')')
		case rune:
			return append(result, fmt.Sprintf("%q", e)...)
		case string:
			return append(result, fmt.Sprintf("%q", e)...)
		case *list.Pair:
			if e == nil {
				sym := cmp.encloseSymbol(lib.Intern("github.com/exascience/slick/list", "Nil"))
				result = append(result, sym.Package...)
				result = append(result, '.')
				result = append(result, sym.Identifier...)
				return append(result, '(', ')')
			}
			switch e.Car {
			case _make:
				return cmp.compileMakeExpression(result, e)
			case _make_struct:
				return cmp.compileStructLiteral(result, e)
			case _make_array:
				return cmp.compileVectorLiteral(result, "array", e)
			case _make_slice:
				return cmp.compileVectorLiteral(result, "slice", e)
			case _make_map:
				return cmp.compileMapLiteral(result, e)
			case _func:
				return cmp.compileFuncLiteral(result, e)
			case _slot:
				return cmp.compileSlotExpression(result, e)
			case _at:
				return cmp.compileIndexExpression(result, e)
			case _slice:
				return cmp.compileSliceExpression(result, e)
			case _assert:
				return cmp.compileAssertExpression(result, e)
			case _convert:
				return cmp.compileConvertExpression(result, e)
			case _values:
				rest := e.Cdr.(*list.Pair)
				result = cmp.compileExpr(result, form, rest.Car, operatorAllowed)
				rest.Cdr.(*list.Pair).ForEach(func(element interface{}) {
					result = append(result, ',', ' ')
					result = cmp.compileExpr(result, form, element, operatorAllowed)
				})
				return result
			case _plus, _minus, _mul, _div, _rem, _bang, _xor, _and, _and_not, _or, _shl, _shr, _arrow_left,
				_bool_and, _bool_or, _equal_equal, _not_equal, _less, _less_equal, _greater, _greater_equal:
				if !operatorAllowed {
					cmp.error(form, "no operator expression allowed in this context")
				}
				return cmp.compileOperatorExpression(result, e)
			default:
				if sym, ok := e.Car.(*lib.Symbol); ok {
					switch sym {
					case _quote, _quasiquote, _unquote, _unquote_splicing:
						p := cmp.resolveLibPlugin()
						var macroSym plugin.Symbol
						var err error
						switch sym {
						case _quote:
							macroSym, err = p.Lookup("Quote")
						case _quasiquote:
							macroSym, err = p.Lookup("Quasiquote")
						case _unquote:
							macroSym, err = p.Lookup("Unquote")
						case _unquote_splicing:
							macroSym, err = p.Lookup("UnquoteSplicing")
						}
						if err != nil {
							cmp.error(form, "invalid special form")
						} else if newForm, err := macroSym.(macro)(e, Environment{}); err != nil {
							cmp.error(form, fmt.Sprintf("error during special form processing: %v", err))
						} else {
							element = newForm
							continue
						}
					}
					if len(sym.Package) > 0 && sym.Package[0] == '#' {
						p := cmp.resolvePlugin(sym.Package)
						if macroSym, err := p.Lookup(sym.Identifier); err != nil {
							cmp.error(form, "invalid macro invocation")
						} else if newForm, err := macroSym.(macro)(e, Environment{}); err != nil {
							cmp.error(form, fmt.Sprintf("error during macroexpansion: %v", err))
						} else {
							element = newForm
							continue
						}
					}
				}
				return cmp.compileCallExpression(result, e)
			}
		default:
			cmp.error(form, fmt.Sprintf("Invalid expression %v.", e))
			return result
		}
	}
}

func (cmp *compiler) compileExpression(result []byte, form *list.Pair, element interface{}) []byte {
	return cmp.compileExpr(result, form, element, true)
}

func (cmp *compiler) compilePrimaryExpression(result []byte, form *list.Pair, element interface{}) []byte {
	return cmp.compileExpr(result, form, element, false)
}

func (cmp *compiler) compileFile() []byte {
	defer func() {
		cmp.header = nil
	}()
	cmp.header = cmp.compilePackageClause()

	if cmp.reader.Errors.Len() != 0 {
		return nil
	}

	cmp.reader.SkipSpace()
	offset := cmp.reader.Offset()
	element := cmp.reader.Read()
	form, ok := element.(*list.Pair)

	for ok && form != nil && form.Car == _import {
		cmp.compileImportDecl(form)
		cmp.reader.SkipSpace()
		offset = cmp.reader.Offset()
		element = cmp.reader.Read()
		form, ok = element.(*list.Pair)
	}

	for ok && form != nil && form.Car == _use {
		cmp.compileUseDecl(form)
		cmp.reader.SkipSpace()
		offset = cmp.reader.Offset()
		element = cmp.reader.Read()
		form, ok = element.(*list.Pair)
	}

	var result []byte

	for ok && form != nil {
		result = cmp.compileDecl(result, form)
		cmp.reader.SkipSpace()
		offset = cmp.reader.Offset()
		element = cmp.reader.Read()
		if element == io.EOF {
			break
		}
		form, ok = element.(*list.Pair)
	}

	if !ok {
		cmp.reader.Error(offset, "invalid top-level form ")
	}

	if cmp.reader.Errors.Len() != 0 {
		return nil
	}

	cmp.header = append(cmp.header, '\n')
	result = append(cmp.header, result...)
	return result
}

func Compile(rd *reader.Reader) (result []byte, err error) {
	var cmp compiler
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if _, ok := e.(bailout); !ok {
			panic(e)
		}
		err = cmp.reader.Errors.Err()
	}()
	cmp.init(rd)
	return cmp.compileFile(), cmp.reader.Errors.Err()
}
