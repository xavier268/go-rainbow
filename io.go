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
	nbChain      int
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
	binary.Write(buf, mode, uint64(hd.nbChain))
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
	hd.nbChain = int(l)
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
	hd.major, hd.minor = Version()
	hd.chainLen = r.cl
	hd.nbChain = len(r.chains)
	hd.halgo = r.halgo
	hd.used = r.used.String()
	return hd
}

// Check compatibility of header with current Rainbow
func (r *Rainbow) checkHeader(hd *header) error {
	if string(hd.magic[:]) != "go-rainbow" {
		return errors.New("invalid file type")
	}
	M, m := Version()
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
	r.SortChains()
	buf := bufio.NewWriter(writer)
	h := r.getHeader()
	if e := binary.Write(buf, mode, uint64(len(h.toBytes()))); e != nil {
		return e
	}
	if _, e := buf.Write(h.toBytes()); e != nil {
		return e
	}
	for _, c := range r.chains {
		for i := 0; i < r.cl; i++ {
			if e := binary.Write(buf, mode, c.start); e != nil {
				return e
			}
			if e := binary.Write(buf, mode, c.end); e != nil {
				return e
			}
		}
	}
	return buf.Flush()
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

	// read chains
	for nc := 0; nc < hd.nbChain; nc++ {
		for i := 0; i < r.cl; i++ {
			c := new(Chain)
			if e := binary.Read(buf, mode, c.start); e != nil {
				return e
			}
			if e := binary.Read(buf, mode, c.end); e != nil {
				return e
			}
			r.AddChain(c)
		}
	}

	// Sort (and dedup) when finished loading.
	r.SortChains()
	return nil
}
