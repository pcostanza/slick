# The Slick programming language

The Slick programming language is a Lisp/Scheme-style s-expression surface syntax for the Go programming language, with a few additional features. Apart from the additional features, it is a faithful mapping of all Go programming language features into s-expression notation, with very few, very minor intentional exceptions.

This is an early release and should be considered work in progress. A lot about Slick might still change, and it is not recommended to use in production unless you're adventurous.

Here is a Hello World in Slick:

```
(package main)

(import "fmt")

(func main () ()
  (fmt:Println "Hello, World!"))
```

Additional features:
* Support for Lisp-style macros.
* Support for Scheme-style quotation and quasiquotation.
* Support for Common-Lisp-style reader / read tables.
* Support for spliced blocks and spliced declarations.

Deviations from Go:
* Block comments nest.
* All numbers start with a digit. Especially, floating point numbers cannot start with a `.`.
* An identifier cannot start with a `_` if it has more than one character.

These are the only deviations. Especially, like Go, Slick is statically typed, distinguishes between statements and expressions, etc., etc.

See the [Slick programming language specification](slick-specification.md) for the full picture.

## Macros

Macros are provided by way of Go's [plugin facility](https://golang.org/pkg/plugin/). For this reason, if you want to use macros, this currently only works on Linux, FreeBSD, and macOS. You should still be able to use Slick on other platforms without macro support, and may be able to cross-compile from Linux, FreeBSD, and macOS, but we haven't tested this yet.

An advantage of using plugins to provide macros is that there is no need for a Slick interpreter. The current implementation is a pure Slick-to-Go compiler.

## Lisp/Scheme-style lists for Go

The Slick package also includes a comprehensive library for pair/cons-cell-based list processing with all kinds of bells and whisles that advanced Lisp/Scheme programmers are used to. The library is designed in such a way that it can be used from Go as well as Slick. See the folder `list` for more details.

## How to build Slick

So far, we only provide the Slick-to-Go compiler. It is a simple compiler that doesn't perform any type checks or advanced semantic analysis on its own, but relies on the Go compiler to do most of the actual work. There is also no build system, or so - the compiler so far can only process one file at a time.

### Building the core Slick compiler

You can build the core Slick compiler just by issuing `go build` in the main folder.

### Building the core Slick plugin

The core Slick plugin provides implementations for quotation and quasiquotation as a built-in set of macros implemented in Slick itself. Therefore, you first need to build the Slick compiler before you can compile the Slick plugin. Proceed as follows:

* Change to the directory `lib/slick`.
* Compile the Slick plugin to Go using the Slick compiler: `slick plugin.slick plugin.go`.
* [Optional] Format the code to make it look nicer: `go fmt plugin.go`.
* Build the Slick plugin: `go build -buildmode=plugin plugin.go`.
* Change back to the root folder: `../..`.
* Copy the plugin to the plugins folder: `cp lib/slick/plugin.so plugins`.

### Building your own macro libraries

A library typically consists of some runtime support functions and some compile-time macro functions.

Let's say you want to implement a library for Lisp/Scheme-style binding forms. Here is a step-by-step guide how to do this:

* Create a folder `bindings` and change to it.
* Create a `go.mod` file for the Go module system. In our example, we add only one line: `module github.com/exascience/bindings`. Feel free to use your own package name and location.
* Create a file `bind.slick` with the following contents. (This is not a very useful library, but provided only for illustration.)

```
    (package bindings)

    (func Bind ((f (func ((x (interface))))) (x (interface))) ()
      (f x))
```

* Compile the Slick code to Go: `slick bind.slick bind.go`.
* [Optional] Format the code to make it look nicer: `go fmt bind.go`.
* Create a folder `slick` and change to it.
* Create a file `plugin.slick` with the following contents. This file provides the macro function `LetStar` which is similar to the `let*` binding form found in many Lisp and Scheme dialects.

```
    (package main)

    (import
      "github.com/exascience/slick/list"
      "github.com/exascience/slick/compiler"
      '(bl "github.com/exascience/bindings"))

    (use '(bp "github.com/exascience/bindings"))

    (func LetStar ((form (* list:Pair)) (_ compiler:Environment))
                  ((newForm (interface)) (_ error))
      (:= bindings (list:Cadr form))
      (:= body (list:Cddr form))
      (if (== bindings (list:Nil))
        (return (values `(splice ,@body) nil))
        (begin
          (:= firstBinding (list:Car bindings))
          (:= restBindings (list:Cdr bindings))
          (return (values
                    `(bl:Bind (func ((,(list:Car firstBinding) (interface))) ()
                                (bp:LetStar (,@restBindings) ,@body))
                              ,(list:Cadr firstBinding))
                    nil)))))
```

* Set the `SLICKROOT` environment variable to the root folder of your Slick installation: `export SLICKROOT=~/develop/go/slick`, so that the default Slick plugin (with support for quotation and quasiquotation) can be found.
* Compile the Slick code to Go: `slick plugin.slick plugin.go`.
* [Optional] Format the code to make it look nicer: `go fmt plugin.go`.
* Build the plugin: `go build -buildmode=plugin plugin.go`.

### Using a macro library

Let's try to use the bindings library in an example project.

* Create a `test` folder and change to it.
* Create a file `hello.slick` with the following contents:

```
    (package main)

    (import "fmt")

    (use "github.com/exascience/bindings")

    (func main () ()
      (bindings:LetStar ((x "Hello, World!"))
        (fmt:Println x)))
```

* Create a folder for storing the `bindings` plugin that we want to use in this code: `mkdir -p plugins/github.com/exascience/bindings/slick`.
* Copy the plugin from above into this folder: `cp ~/develop/go/bindings/slick/plugin.so plugins/github.com/exascience/bindings/slick`.
* Set the `SLICKPATH` environment variable to the current folder, so that the `bindings` plugin can be found: `export SLICKPATH=.`.
* Compile the Slick code to Go: `slick hello.slick hello.go`.
* [Optional] Format the code to make it look nicer: `go fmt hello.go`.
* Run the program: `go run hello.go`.

## What's next?

The next step is to make sure that user-defined read tables can be used in source code; and that "funcall" forms can be declared (the latter is not straightforward to explain, not sure it will actually work, but also not that important).

Other things to do:

* Add static type checks, semantic analysis, and support for compiler environments directly into the Slick compiler.
* Add support for building complete packages rather than single files.
* Get rid of `SLICKROOT` and `SLICKPATH`, and embed Slick into the Go module system somehow.
* Add support for local macros (`macrolet`), which will actually require to also implement an interpreter for Slick.
* Stress-test the compiler, try out all the language features, see if they work well and are easy to read and write.
* Write some tutorials, example code, better documentation, etc.

If you want to help with any of this, you are very welcome. Feel free to contact me at pascal dot costanza at imec dot be
