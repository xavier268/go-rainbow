package main

import (
	"crypto"
	"crypto/md5"
	"fmt"

	"github.com/xavier268/go-rainbow"
)

func main() {

	// create a rainbow table using MD5 hash, chain length 1_000
	r := rainbow.New(crypto.MD5, 1_000).
		// compile the reduce name space
		CompileAlphabet("ab", 3, 3).
		CompileAlphabet("123", 0, 1).
		Build()

	fmt.Println("Starting to build the table, please wait ...")

	// add 10_000 chains
	for i := 0; i < 10_000; i++ {
		c := r.NewChain()
		r.AddChain(c)
	}

	fmt.Println("Table construction completed")

	// create a few testing hash for a valid passwords
	md5 := md5.New()

	md5.Reset()
	md5.Write([]byte("aba3"))
	h1 := md5.Sum([]byte{})

	md5.Reset()
	md5.Write([]byte("bba"))
	h2 := md5.Sum([]byte{})

	// lookup these hashes in the tables
	fmt.Println("Looking for 2 passwords ...")
	if p, found := r.Lookup(h1); found {
		fmt.Println("Found password : ", string(p))
	}
	if p, found := r.Lookup(h2); found {
		fmt.Println("Found password : ", string(p))
	}
	fmt.Println("No more password found")
}
