package main

import (
	"crypto"

	"github.com/xavier268/go-rainbow"
)

func main() {

	// create a rainbow table using MD5 hash
	rainbow.New(crypto.MD5, 1000).
		// compile the reduce name space
		CompileAlphabet("abcdef", 3, 3).
		CompileAlphabet("1234567890", 0, 1)
		//Build()
	for i := 0; i < 1000; i++ {

	}
}
