# Lisp/Scheme-style lists for Go

Package github.com/exascience/slick/list implements a list library for Go that mimicks functionality that is typically found in Lisp and Scheme languages which is based on pairs or cons cells.

To a large extent, this package follows the SRFI 1 specification for Scheme authored by Olin Shivers, which is part of the Scheme Requests For Implementations library. See https://srfi.schemers.org/srfi-1/srfi-1.html for SRFI 1, and https://srfi.schemers.org for Scheme Requests For Implementation.

Although this package is part of the slick repository, it is intended to be usable independently from the Slick programming language in Go projects, and to be maintained with such use in mind.
