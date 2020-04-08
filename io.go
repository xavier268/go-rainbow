package rainbow

import (
	"bufio"
	"bytes"
	"crypto"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
)

// header for the files
type header struct {
	// fixed size
	magic        string
	major, minor int
	chainLen     int
	halgo        crypto.Hash
	used         string
}

func (hd *header) String() string {
	return fmt.Sprintf("%+v", *hd)
}

const magic = "go-rainbow"

var mode = binary.LittleEndian

func (hd *header) Equal(h *header) bool {
	return hd.magic == h.magic &&
		hd.major == h.major &&
		hd.minor == h.minor &&
		hd.chainLen == h.chainLen &&
		hd.halgo == h.halgo &&
		hd.used == h.used
}

func (hd *header) toBytes() []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, mode, uint64(len(hd.magic)))
	binary.Write(buf, mode, []byte(hd.magic))
	binary.Write(buf, mode, uint64(hd.major))
	binary.Write(buf, mode, uint64(hd.minor))
	binary.Write(buf, mode, uint64(hd.chainLen))
	binary.Write(buf, mode, uint64(hd.halgo))
	binary.Write(buf, mode, uint64(len(hd.used)))
	binary.Write(buf, mode, []byte(hd.used))
	return append([]byte{}, buf.Bytes()...)
}

func (hd *header) fromBytes(data []byte) *header {
	buf := bytes.NewReader(data)
	var l uint64
	var bb []byte
	if err := binary.Read(buf, mode, &l); err != nil {
		panic(err)
	}
	bb = make([]byte, l)
	binary.Read(buf, mode, bb)
	hd.magic = string(bb)
	binary.Read(buf, mode, &l)
	hd.major = int(l)
	binary.Read(buf, mode, &l)
	hd.minor = int(l)
	binary.Read(buf, mode, &l)
	hd.chainLen = int(l)
	binary.Read(buf, mode, &l)
	hd.halgo = crypto.Hash(l)
	binary.Read(buf, mode, &l)
	bb = make([]byte, l)
	binary.Read(buf, mode, bb)
	hd.used = string(bb)
	return hd
}

// getHeader from current Rainbow
func (r *Rainbow) getHeader() *header {
	hd := new(header)
	hd.magic = "go-rainbow"
	hd.major, hd.minor, _ = Version()
	hd.chainLen = r.cl
	hd.halgo = r.halgo
	hd.used = r.used.String()
	return hd
}

// Check compatibility of header with current Rainbow
func (r *Rainbow) checkHeader(hd *header) error {
	if string(hd.magic[:]) != "go-rainbow" {
		return errors.New("invalid file type")
	}
	M, m, _ := Version()
	if M != hd.major {
		return fmt.Errorf("current program version (%d) does not match file version (%d)", M, hd.major)
	}
	if m < hd.minor {
		return fmt.Errorf("current program version (%d.%d) is obsolete for the file(%d.%d)", M, m, hd.major, hd.minor)
	}
	if hd.chainLen != r.cl {
		return errors.New("chain length do not match")
	}
	if hd.halgo != r.halgo {
		return errors.New("incorrect hash algorithm")
	}
	if u, ok := new(big.Int).SetString(hd.used, 10); !ok || u.Cmp(r.used) != 0 {
		return errors.New("reduce signature are different")
	}
	return nil
}

// Save rainbow table content
func (r *Rainbow) Save(writer io.Writer) error {
	r.DedupChains()
	buf := bufio.NewWriter(writer)
	h := r.getHeader()
	if e := binary.Write(buf, mode, uint64(len(h.toBytes()))); e != nil {
		return e
	}
	if _, e := buf.Write(h.toBytes()); e != nil {
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
	fmt.Println("Saved ", h)
	return e
}

// Load will load chains from the reader,
// adding them to the existing rainbow table.
// File compatibility is (approximately) verified.
func (r *Rainbow) Load(reader io.Reader) error {

	buf := bufio.NewReader(reader)

	// read and check header
	var hl uint64
	if e := binary.Read(buf, mode, &hl); e != nil {
		return e
	}
	bb := make([]byte, hl)
	if _, e := buf.Read(bb); e != nil {
		return e
	}
	hd := new(header).fromBytes(bb)
	if e := r.checkHeader(hd); e != nil {
		return e
	}

	fmt.Println("Loading ", hd)

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
