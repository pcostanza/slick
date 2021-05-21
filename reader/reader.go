package reader

/*
The reader package is mostly based on Common Lisp concepts, with a few
differences:
  - Fewer and different macro and dispatch macro runes.
  - No read-time evaluation, and no plans to include it.
  - No case conversion, and no plans to include it.
  - The name of the keyword package is _keyword.
  - Number syntax is the same as in Go. Exception: Floating-point
    numbers cannot start with a dot, because all numbers have to start
    with a digit.
  - String syntax is the same as in Go.
  - Backquote, comma, and comma-at are not processed by the reader,
    but are just short forms for quasiquote, unquote, and
    unquote-splicing which are processed elsewhere.
  - Identifiers are a superset of Go identifiers and allow special
    runes, just like in Common Lisp and Scheme. However, only Go
    identifiers can eventually end up in generated source code, so
    identifiers with special runes can only be used for
    compile-time entities, such as macro names.
  - As an exception, identifiers starting with underscore are not
    allowed by the reader, unless it's the single-rune underscore
    placeholder name. Identifiers starting with underscore are
    reserved for internal uses, such as gensyms that should not
    accidentally clash with user-defined names. (Identifiers starting
    with underscores are acceptable Go syntyax, which means gensyms
    can be used for base-level entities.)
*/

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"go/scanner"
	"go/token"

	"github.com/pcostanza/slick/lib"
	"github.com/pcostanza/slick/list"
)

type (
	Macro         func(*Reader) interface{}
	DispatchMacro func(rd *Reader, dispatchRune rune, dispatchRuneOffset int) interface{}

	Table struct {
		macroRunes         map[rune]Macro
		dispatchMacroRunes map[rune]map[rune]DispatchMacro
		terminating        map[rune]bool
	}
)

func NewTable() *Table {
	return &Table{
		macroRunes:         make(map[rune]Macro),
		dispatchMacroRunes: make(map[rune]map[rune]DispatchMacro),
		terminating:        make(map[rune]bool),
	}
}

func MakeTable(macroRunes map[rune]Macro, dispatchMacroRunes map[rune]map[rune]DispatchMacro, terminating map[rune]bool) *Table {
	return &Table{
		macroRunes:         macroRunes,
		dispatchMacroRunes: dispatchMacroRunes,
		terminating:        terminating,
	}
}

func CopyTable(rt *Table) *Table {
	result := NewTable()
	for key, val := range rt.macroRunes {
		result.macroRunes[key] = val
	}
	for key, val := range rt.dispatchMacroRunes {
		dt := make(map[rune]DispatchMacro)
		for skey, sval := range val {
			dt[skey] = sval
		}
		result.dispatchMacroRunes[key] = dt
	}
	for key, val := range rt.terminating {
		result.terminating[key] = val
	}
	return result
}

var StandardTable = &Table{
	macroRunes: map[rune]Macro{
		'(':  listMacro,
		')':  ErrorMacro,
		'\'': quoteMacro,
		';':  lineCommentMacro,
		'"':  stringMacro,
		'`':  quasiquoteMacro,
		',':  unquoteMacro,
	},
	dispatchMacroRunes: map[rune]map[rune]DispatchMacro{
		'#': map[rune]DispatchMacro{
			'`':  rawStringMacro,
			'\\': runeMacro,
			';':  formCommentMacro,
			'|':  blockCommentMacro,
		},
	},
	terminating: map[rune]bool{'"': true, '\'': true, '(': true, ')': true, ',': true, ';': true, '`': true},
}

func init() {
	StandardTable.macroRunes['#'] = dispatchMacroReader(StandardTable.dispatchMacroRunes['#'])
}

func (rt *Table) SetMacroRune(r rune, f Macro, terminating bool) {
	rt.macroRunes[r] = f
	if terminating {
		rt.terminating[r] = true
	} else {
		delete(rt.terminating, r)
	}
}

func (rt *Table) GetMacroRune(r rune) (f Macro, terminating bool) {
	return rt.macroRunes[r], rt.terminating[r]
}

func (rt *Table) MakeDisptachMacroRune(r rune, terminating bool) {
	subtable := make(map[rune]DispatchMacro)
	rt.dispatchMacroRunes[r] = subtable
	rt.SetMacroRune(r, dispatchMacroReader(subtable), terminating)
}

func (rt *Table) SetDispatchMacroRune(dispRune, subRune rune, f DispatchMacro) {
	rt.dispatchMacroRunes[dispRune][subRune] = f
}

func (rt *Table) GetDispatchMacroRune(dispRune, subRune rune) DispatchMacro {
	return rt.dispatchMacroRunes[dispRune][subRune]
}

type PackageResolver struct {
	PackageToPath, PathToPackage map[string]string
}

func NewPackageResolver() *PackageResolver {
	return &PackageResolver{
		make(map[string]string),
		make(map[string]string),
	}
}

func (r PackageResolver) ResolveSymbol(pkg, ident string) (*lib.Symbol, error) {
	if pkg == "" || pkg == "_keyword" {
		return lib.Intern(pkg, ident), nil
	}
	if path, ok := r.PackageToPath[pkg]; ok {
		return lib.Intern(path, ident), nil
	}
	return nil, fmt.Errorf("The package of symbol %v:%v cannot be resolved.", pkg, ident)
}

func (r *PackageResolver) EncloseSymbol(sym *lib.Symbol) (*lib.Symbol, bool) {
	if sym.Package == "" || sym.Package == "_keyword" {
		return sym, false
	}
	if name, ok := r.PathToPackage[sym.Package]; ok {
		return lib.Intern(name, sym.Identifier), false
	}
	newName := path.Base(sym.Package)
	if _, ok := r.PackageToPath[newName]; ok {
		for counter := 1; ; counter++ {
			modName := fmt.Sprintf("%v%v", newName, counter)
			if _, ok := r.PackageToPath[modName]; !ok {
				newName = modName
				break
			}
		}
	}
	r.PackageToPath[newName] = sym.Package
	r.PathToPackage[sym.Package] = newName
	return lib.Intern(newName, sym.Identifier), true
}

type formRange struct {
	from, to int
}

type Reader struct {
	*PackageResolver
	file     *token.File
	Errors   scanner.ErrorList
	src      []byte
	table    *Table
	ranges   map[*list.Pair]formRange
	ch       rune
	offset   int
	rdOffset int
}

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, s); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

const bom = 0xFEFF

func NewReader(fset *token.FileSet, filename string, src interface{}, table *Table) (*Reader, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	source, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}
	if table == nil {
		table = StandardTable
	}
	rd := &Reader{
		PackageResolver: NewPackageResolver(),
		file:            fset.AddFile(filename, -1, len(source)),
		src:             source,
		table:           table,
		ranges:          make(map[*list.Pair]formRange),
		ch:              ' ',
	}
	rd.NextRune()
	if rd.ch == bom {
		rd.NextRune()
	}
	if err := rd.Errors.Err(); err != nil {
		return nil, err
	}
	return rd, nil
}

func (rd *Reader) File() *token.File {
	return rd.file
}

func (rd *Reader) Table() *Table {
	return rd.table
}

func (rd *Reader) Offset() int {
	return rd.offset
}

func (rd *Reader) AddForm(form *list.Pair, from, to int) {
	rd.ranges[form] = formRange{from: from, to: to}
}

func (rd *Reader) FormPos(form *list.Pair) (pos, end token.Pos) {
	if formRange, ok := rd.ranges[form]; ok {
		pos = rd.file.Pos(formRange.from)
		end = rd.file.Pos(formRange.to)
	}
	return
}

func (rd *Reader) NextRune() rune {
	if rd.rdOffset < len(rd.src) {
		rd.offset = rd.rdOffset
		if rd.ch == '\n' {
			rd.file.AddLine(rd.offset)
		}
		r, w := rune(rd.src[rd.rdOffset]), 1
		switch {
		case r == 0:
			rd.Error(rd.offset, "illegal rune NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(rd.src[rd.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				rd.Error(rd.offset, "illegal UTF-8 encoding")
			} else if r == bom && rd.offset > 0 {
				rd.Error(rd.offset, "illegal byte order mark")
			}
		}
		rd.rdOffset += w
		rd.ch = r
		return r
	}
	rd.offset = len(rd.src)
	if rd.ch == '\n' {
		rd.file.AddLine(rd.offset)
	}
	rd.ch = -1 // eof
	return -1
}

func (rd *Reader) Bytes() []byte {
	return rd.src
}

func (rd *Reader) Rune() rune {
	return rd.ch
}

type BadForm struct {
	from token.Pos
	to   token.Pos
}

func (b *BadForm) Pos() token.Pos {
	return b.from
}

func (b *BadForm) End() token.Pos {
	return b.to
}

func (rd *Reader) BadForm(fromOffset, toOffset int) *BadForm {
	return &BadForm{
		from: rd.file.Pos(fromOffset),
		to:   rd.file.Pos(toOffset),
	}
}

func (rd *Reader) Error(offset int, msg string) {
	rd.Errors.Add(rd.file.Position(rd.file.Pos(offset)), msg)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9' || r >= utf8.RuneSelf && unicode.IsDigit(r)
}

func (rd *Reader) SkipSpace() {
	for rd.ch != -1 && unicode.IsSpace(rd.ch) {
		rd.NextRune()
	}
}

func ErrorMacro(rd *Reader) interface{} {
	offset := rd.offset
	rd.Error(offset, fmt.Sprintf("invalid macro rune %c", rd.Rune()))
	rd.NextRune()
	return rd.BadForm(offset, rd.offset)
}

func (rd *Reader) ReadDelimitedList(delimiter rune) interface{} {
	offset := rd.offset
	rd.NextRune()
	var result *list.Pair
	for {
		rd.SkipSpace()
		if rd.Rune() == delimiter {
			rd.NextRune()
			result = result.NReverse()
			rd.AddForm(result, offset, rd.offset)
			return result
		}
		element := rd.Read()
		if element == io.EOF {
			rd.Error(offset, "incomplete list")
			return rd.BadForm(offset, rd.offset)
		}
		result = list.NewPair(element, result)
	}
}

func listMacro(rd *Reader) interface{} {
	return rd.ReadDelimitedList(')')
}

var (
	quote           = lib.Intern("", "quote")
	quasiquote      = lib.Intern("", "quasiquote")
	unquote         = lib.Intern("", "unquote")
	unquoteSplicing = lib.Intern("", "unquote-splicing")
)

func quoteMacro(rd *Reader) interface{} {
	offset := rd.offset
	rd.NextRune()
	element := rd.Read()
	if element == io.EOF {
		rd.Error(offset, "incomplete quote")
		return rd.BadForm(offset, rd.offset)
	}
	result := list.List(quote, element)
	rd.AddForm(result, offset, rd.offset)
	return result
}

func quasiquoteMacro(rd *Reader) interface{} {
	offset := rd.offset
	rd.NextRune()
	element := rd.Read()
	if element == io.EOF {
		rd.Error(offset, "incomplete quasiquote")
		return rd.BadForm(offset, rd.offset)
	}
	result := list.List(quasiquote, element)
	rd.AddForm(result, offset, rd.offset)
	return result
}

func unquoteMacro(rd *Reader) interface{} {
	offset := rd.offset
	var splicing bool
	if rd.NextRune() == '@' {
		splicing = true
		rd.NextRune()
	}
	element := rd.Read()
	if element == io.EOF {
		rd.Error(offset, "incomplete unquote")
		return rd.BadForm(offset, rd.offset)
	}
	if splicing {
		result := list.List(unquoteSplicing, element)
		rd.AddForm(result, offset, rd.offset)
		return result
	}
	result := list.List(unquote, element)
	rd.AddForm(result, offset, rd.offset)
	return result
}

func lineCommentMacro(rd *Reader) interface{} {
	rd.NextRune()
	for {
		if r := rd.NextRune(); r == '\n' || r == -1 {
			rd.NextRune()
			return nil
		}
	}
}

func (rd *Reader) readHexBytes(n int) (result int, ok bool) {
	for i, r := 0, rd.Rune(); i < n; i, r = i+1, rd.NextRune() {
		if '0' <= r && r <= '9' {
			result = result<<4 + int(r-'0')
		} else if 'a' <= r && r <= 'f' {
			result = result<<4 + int(r-'a') + 10
		} else if 'A' <= r && r <= 'F' {
			result = result<<4 + int(r-'A') + 10
		} else {
			rd.Error(rd.offset, "invalid hex digit")
			return result, false
		}
	}
	return result, true
}

func (rd *Reader) readOctalByte() (result int, ok bool) {
	for i, r := 0, rd.Rune(); i < 3; i, r = i+1, rd.NextRune() {
		if '0' <= r && r <= '7' {
			result = result<<3 + int(r-'0')
		} else {
			rd.Error(rd.offset, "invalid octal digit")
			return result, false
		}
	}
	return result, true
}

func stringMacro(rd *Reader) interface{} {
	offset := rd.offset
	d := rd.Rune()
	fastForward := func() {
		for {
			if r := rd.NextRune(); r == -1 || r == '\n' || r == d {
				break
			} else if r == '\\' {
				rd.NextRune()
			}
		}
		rd.NextRune()
	}
	var result bytes.Buffer
	for {
		r := rd.NextRune()
		if r == -1 || r == '\n' {
			rd.Error(offset, "incomplete string literal")
			rd.NextRune()
			return rd.BadForm(offset, rd.offset)
		}
		if r == d {
			rd.NextRune()
			return result.String()
		}
		if r != '\\' {
			if _, err := result.WriteRune(r); err != nil {
				rd.Error(offset, err.Error())
				fastForward()
				return rd.BadForm(offset, rd.offset)
			}
			continue
		}
		escapeOffset := rd.offset
		r = rd.NextRune()
		var err error
		switch r {
		case -1:
			rd.Error(escapeOffset, "incomplete escape in string literal")
			return rd.BadForm(offset, rd.offset)
		case 'a':
			_, err = result.WriteRune('\a')
		case 'b':
			_, err = result.WriteRune('\b')
		case 'f':
			_, err = result.WriteRune('\f')
		case 'n':
			_, err = result.WriteRune('\n')
		case 'r':
			_, err = result.WriteRune('\r')
		case 't':
			_, err = result.WriteRune('\t')
		case 'v':
			_, err = result.WriteRune('\v')
		case '\\':
			_, err = result.WriteRune('\\')
		case '"':
			_, err = result.WriteRune('"')
		case 'x':
			rd.NextRune()
			if b, ok := rd.readHexBytes(2); ok {
				err = result.WriteByte(byte(b))
			}
		case '0', '1', '2', '3', '4', '5', '6', '7':
			if b, ok := rd.readOctalByte(); ok {
				err = result.WriteByte(byte(b))
			}
		case 'u':
			rd.NextRune()
			if u, ok := rd.readHexBytes(4); ok {
				_, err = result.WriteRune(rune(u))
			}
		case 'U':
			rd.NextRune()
			if u, ok := rd.readHexBytes(8); ok {
				_, err = result.WriteRune(rune(u))
			}
		default:
			rd.Error(escapeOffset, "invalid escape in string literal")
		}
		if err != nil {
			rd.Error(escapeOffset, err.Error())
			fastForward()
			return rd.BadForm(offset, rd.offset)
		}
	}
}

func rawStringMacro(rd *Reader, _ rune, dispatchRuneOffset int) interface{} {
	d := rd.Rune()
	fastForward := func() {
		for {
			if r := rd.NextRune(); r == -1 || r == d {
				break
			}
		}
		rd.NextRune()
	}
	var result bytes.Buffer
	for {
		r := rd.NextRune()
		if r == -1 {
			rd.Error(dispatchRuneOffset, "incomplete raw string literal")
			return rd.BadForm(dispatchRuneOffset, rd.offset)
		}
		if r == d {
			rd.NextRune()
			return result.String()
		}
		if r == '\r' {
			continue
		}
		if _, err := result.WriteRune(r); err != nil {
			rd.Error(dispatchRuneOffset, err.Error())
			fastForward()
			return rd.BadForm(dispatchRuneOffset, rd.offset)
		}
	}
}

func runeMacro(rd *Reader, _ rune, dispatchRuneOffset int) interface{} {
	r := rd.NextRune()
	if r != '\\' {
		return r
	}
	r = rd.NextRune()
	switch r {
	case -1:
		rd.Error(dispatchRuneOffset, "incomplete rune literal")
	case 'a':
		rd.NextRune()
		return '\a'
	case 'b':
		rd.NextRune()
		return '\b'
	case 'f':
		rd.NextRune()
		return '\f'
	case 'n':
		rd.NextRune()
		return '\n'
	case 'r':
		rd.NextRune()
		return '\r'
	case 's':
		rd.NextRune()
		return ' '
	case 't':
		rd.NextRune()
		return '\t'
	case 'v':
		rd.NextRune()
		return '\v'
	case '\\':
		rd.NextRune()
		return '\\'
	case '\'':
		rd.NextRune()
		return '\''
	case 'x':
		rd.NextRune()
		if b, ok := rd.readHexBytes(2); ok {
			return b
		}
	case '0', '1', '2', '3', '4', '5', '6', '7':
		if b, ok := rd.readOctalByte(); ok {
			return b
		}
	case 'u':
		rd.NextRune()
		if u, ok := rd.readHexBytes(4); ok {
			return u
		}
	case 'U':
		rd.NextRune()
		if u, ok := rd.readHexBytes(8); ok {
			return u
		}
	default:
		rd.Error(dispatchRuneOffset, "invalid escape in rune literal")
	}
	return rd.BadForm(dispatchRuneOffset, rd.offset)
}

func formCommentMacro(rd *Reader, _ rune, dispatchRuneOffset int) interface{} {
	rd.NextRune()
	if rd.Read() == io.EOF {
		rd.Error(dispatchRuneOffset, "incomplete form comment")
		return rd.BadForm(dispatchRuneOffset, rd.offset)
	}
	return nil
}

func blockCommentMacro(rd *Reader, c1 rune, dispatchRuneOffset int) interface{} {
	c2 := rd.Rune()
	rd.NextRune()
	level := 1
	for {
		switch r := rd.Rune(); r {
		case -1:
			rd.Error(dispatchRuneOffset, "incomplete block comment")
			return rd.BadForm(dispatchRuneOffset, rd.offset)
		case c1:
			if rd.NextRune() == c2 {
				rd.NextRune()
				level++
			}
		case c2:
			if rd.NextRune() == c1 {
				rd.NextRune()
				if level--; level == 0 {
					return nil
				}
			}
		default:
			rd.NextRune()
		}
	}
}

func dispatchMacroReader(subtable map[rune]DispatchMacro) Macro {
	return func(rd *Reader) interface{} {
		offset := rd.offset
		r := rd.Rune()
		if readerMacro, ok := subtable[rd.NextRune()]; ok {
			return readerMacro(rd, r, offset)
		}
		rd.Error(offset, "invalid dispatch macro rune")
		return rd.BadForm(offset, rd.offset)
	}
}

func (rd *Reader) readIdentifier() string {
	offset := rd.offset
	var buf bytes.Buffer
	r := rd.Rune()
	for {
		if !validRune(r) || r == ':' || rd.table.terminating[r] {
			return buf.String()
		}
		if _, err := buf.WriteRune(r); err != nil {
			rd.Error(offset, err.Error())
			for {
				r = rd.NextRune()
				if !validRune(r) || r == ':' || rd.table.terminating[r] {
					return buf.String()
				}
			}
		}
		r = rd.NextRune()
	}
}

func (rd *Reader) readSymbol() interface{} {
	offset := rd.offset
	ok := true
	pkg := rd.readIdentifier()
	if (pkg != "_") && strings.HasPrefix(pkg, "_") {
		rd.Error(offset, "invalid package name or identifier")
		ok = false
	}
	if rd.Rune() != ':' {
		if pkg == "" {
			rd.Error(offset, "empty identifier")
			ok = false
		}
		if ok {
			return lib.Intern("", pkg)
		}
		return rd.BadForm(offset, rd.offset)
	}
	if pkg == "" {
		pkg = "_keyword"
	}
	rd.NextRune()
	ident := rd.readIdentifier()
	if rd.Rune() == ':' {
		rd.Error(offset, "invalid package prefix")
		ok = false
	} else if ident == "" {
		rd.Error(offset, "empty identifier")
		ok = false
	}
	if (ident != "_") && strings.HasPrefix(ident, "_") {
		rd.Error(offset, "invalid identifier")
		ok = false
	}
	if ok {
		if sym, err := rd.ResolveSymbol(pkg, ident); err != nil {
			return rd.BadForm(offset, rd.offset)
		} else {
			return sym
		}
	}
	return rd.BadForm(offset, rd.offset)
}

func isNumRune(r rune) bool {
	if '0' <= r && r <= '9' {
		return true
	}
	switch r {
	case '_', 'x', 'X', 'b', 'B', 'o', 'O', '.', 'e', 'E', 'p', 'P', '+', '-':
		return true
	default:
		return false
	}
}

func (rd *Reader) readNumber() interface{} {
	offset := rd.offset
	var buf bytes.Buffer
	var flt bool
	r := rd.Rune()
	for ; isNumRune(r); r = rd.NextRune() {
		if r == '.' || r == 'e' || r == 'E' || r == 'p' || r == 'P' {
			flt = true
		}
		if _, err := buf.WriteRune(r); err != nil {
			rd.Error(offset, err.Error())
			for isNumRune(rd.NextRune()) {
			}
			return rd.BadForm(offset, rd.offset)
		}
	}
	str := buf.String()
	if r == 'i' {
		rd.NextRune()
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			rd.Error(offset, err.Error())
			return rd.BadForm(offset, rd.offset)
		}
		return complex(0, val)
	}
	if flt {
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			rd.Error(offset, err.Error())
			return rd.BadForm(offset, rd.offset)
		}
		return val
	}
	var result big.Int
	if val, ok := result.SetString(str, 0); ok {
		return val
	}
	rd.Error(offset, "invalid number syntax")
	return rd.BadForm(offset, rd.offset)
}

func validRune(r rune) bool {
	return '!' <= r && r <= '~' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (rd *Reader) Read() interface{} {
	for {
		rd.SkipSpace()
		r := rd.Rune()
		if r == -1 {
			return io.EOF
		}
		if readerMacro, ok := rd.table.macroRunes[r]; ok {
			if form := readerMacro(rd); form != nil {
				return form
			}
			continue
		}
		if dispatchReaderMacro, ok := rd.table.dispatchMacroRunes[r]; ok {
			offset := rd.offset
			rd.NextRune()
			s := rd.Rune()
			if s == -1 {
				return io.EOF
			}
			if readerMacro, ok := dispatchReaderMacro[s]; ok {
				if form := readerMacro(rd, s, offset); form != nil {
					return form
				}
				continue
			}
			rd.Error(offset, fmt.Sprintf("subrune %q not defined for dispatch rune %q", s, r))
			rd.NextRune()
		}
		if isDigit(r) {
			return rd.readNumber()
		}
		if validRune(r) {
			return rd.readSymbol()
		}
		rd.Error(rd.offset, "invalid rune")
		rd.NextRune()
	}
}

var (
	pkg = lib.Intern("", "package")
	imp = lib.Intern("", "import")
	use = lib.Intern("", "use")
)

type SourceFile struct {
	PackageClause        *list.Pair
	ImportDeclarations   []*list.Pair
	UseDeclarations      []*list.Pair
	TopLevelDeclarations []*list.Pair
}

func (rd *Reader) ReadSourceFile() *SourceFile {
	result := &SourceFile{}
	rd.SkipSpace()
	offset := rd.offset
	if form, ok := rd.Read().(*list.Pair); ok && form != nil && form.Car == pkg {
		result.PackageClause = form
	} else {
		rd.Error(offset, "missing package clause")
	}
	rd.SkipSpace()
	element := rd.Read()
	form, ok := element.(*list.Pair)
	readTopLevelForms := func(sym *lib.Symbol) (result []*list.Pair) {
		for ok && form != nil && (sym == nil || form.Car == sym) {
			result = append(result, form)
			rd.SkipSpace()
			element = rd.Read()
			form, ok = element.(*list.Pair)
		}
		return
	}
	result.ImportDeclarations = readTopLevelForms(imp)
	result.UseDeclarations = readTopLevelForms(use)
	result.TopLevelDeclarations = readTopLevelForms(nil)
	if element != io.EOF {
		rd.Error(rd.offset, "invalid top level form")
	}
	return result
}
