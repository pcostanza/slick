package main

import (
	"fmt"
	"os"

	"github.com/exascience/slick/reader"

	"github.com/exascience/slick/compiler"
)

func main() {
	in, err := reader.NewReader(nil, os.Args[1], nil, nil)
	if err != nil {
		panic(err)
	}

	output, err := compiler.Compile(in)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}

	if _, err = out.Write(output); err != nil {
		panic(err)
	}

	if err = out.Close(); err != nil {
		panic(err)
	}
	fmt.Println("done")
}
