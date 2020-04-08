package rainbow

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var mode = binary.LittleEndian

// Check compatibility of header with current Rainbow
func (r *Rainbow) checkSignature(s string) bool {
	return s == r.signature
}

// Save rainbow table content
func (r *Rainbow) Save(writer io.Writer) error {
	r.DedupChains()

	// save signature-length and signature
	buf := bufio.NewWriter(writer)
	s := r.signature
	if e := binary.Write(buf, mode, uint64(len(s))); e != nil {
		return e
	}
	if _, e := buf.Write([]byte(s)); e != nil {
		return e
	}

	// save nb of chains
	binary.Write(buf, mode, uint64(len(r.chains)))

	// save the chains
	n := 0
	for _, c := range r.chains {
		if e := binary.Write(buf, mode, c.start); e != nil {
			return e
		}
		if e := binary.Write(buf, mode, c.end); e != nil {
			return e
		}
		n++
		if n%1000 == 0 {
			fmt.Println(n, "chains saved")
		}
	}
	e := buf.Flush()
	fmt.Println(n, "chains saved")
	fmt.Println("Saved ", r.signature)
	return e
}

// Load will load chains from the reader,
// adding them to the existing rainbow table.
// File compatibility is (approximately) verified.
func (r *Rainbow) Load(reader io.Reader) error {

	buf := bufio.NewReader(reader)

	// read and check signature
	var sl uint64
	if e := binary.Read(buf, mode, &sl); e != nil {
		return e
	}
	bb := make([]byte, sl)
	if _, e := buf.Read(bb); e != nil {
		return e
	}
	if r.signature != string(bb) {
		return errors.New("Cannot load because signatures do not match")
	}

	// read nb of chains
	var nb uint64
	binary.Read(buf, mode, &nb)
	// read chains
	var e error

	for n := uint64(0); e == nil && n < nb; n++ {

		start := make([]byte, r.hsize)
		end := make([]byte, r.hsize)
		if e = binary.Read(buf, mode, start); e != nil {
			break
		}
		if e = binary.Read(buf, mode, end); e != nil {
			break
		}
		c := new(Chain)
		c.start = start
		c.end = end
		r.AddChain(c)
		if (n+1)%1000 == 0 || e != nil {
			fmt.Println(n+1, "chains loaded")
		}
	}

	// Dedup (and sort) when finished loading.
	r.DedupChains()

	// ignore eof
	if e == io.EOF {
		return nil
	}

	// unexpected errors
	return e
}
