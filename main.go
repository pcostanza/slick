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
		fmt.Println(err)
		os.Exit(1)
	}

	output, err := compiler.Compile(in)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err = out.Write(output); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = out.Close(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("done")
}
