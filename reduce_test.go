package rainbow

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestAlphaReduce(t *testing.T) {
	n := 12
	r := getAlphaReduceFunc(n)
	h := []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 41}
	p := []byte{}
	s := string(r(2, h, p))
	if len(s) != n {
		t.Log(s)
		t.Fatal("unexpected password length")
	}

	n = 7
	r = getAlphaReduceFunc(n)
	s = string(r(2, h, p))
	ss := string(r(3, h, p))
	if len(s) != n || len(ss) != n {
		t.Log(h, p, s, ss)
		t.Fatal("unexpected password length")
	}
	if strings.Compare(s, ss) == 0 {
		t.Log(h, p, s, ss)
		t.Fatal("index does not have an influence on the output")
	}

	n = 7
	r = getAlphaReduceFunc(n)
	s = string(r(0, h, p))
	ss = string(r(1, h, p))
	if len(s) != n || len(ss) != n {
		t.Log(h, p, s, ss)
		t.Fatal("unexpected password length")
	}
	if strings.Compare(s, ss) == 0 {
		t.Log(h, p, s, ss)
		t.Fatal("index does not have an influence on the output")
	}
}

func TestBaseReduce(t *testing.T) {
	h := []byte{0, 1, 2, 3}
	r := GetBaseReduceFunc()
	var first, p []byte
	first = r(8, h, p)
	for step := 9; step < 100_000; step++ {
		p = r(step, h, p)
		if bytes.Compare(p, first) == 0 {
			t.Fatal("unexpected BaseReduce collision at step ", step)
		}
	}
}

func TestStringReduce(t *testing.T) {
	h := []byte{20, 1, 55, 3, 66}
	p := []byte{}
	r := GetStringReduceFunc(3, "Aéï", false)
	results := make(map[string]int) // store test results
	for step := 0; step < 50; step++ {
		p = r(step, h, p)
		results[string(p)]++
	}
	if len(results) != 16 {
		fmt.Println(" number of distincts strings : ", len(results))
		fmt.Println(results)
		t.Fatal("unexpected results length (variable length)")
	}

	r = GetStringReduceFunc(3, "Aéï", true)
	results = make(map[string]int) // store test results
	for step := 0; step < 50; step++ {
		p = r(step, h, p)
		results[string(p)]++
	}
	if len(results) != 12 {
		fmt.Println(" number of distincts strings : ", len(results))
		fmt.Println(results)
		t.Fatal("unexpected results length (fixed length)")
	}
}

func BenchmarkBaseReduce(b *testing.B) {
	h := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	r := GetBaseReduceFunc()
	var p []byte
	b.ResetTimer()
	for step := 0; step < b.N; step++ {
		p = r(step, h, p)
	}
}
func BenchmarkAlphaReduce16(b *testing.B) {
	benchmarkAlphaReduce(16, b)
}
func BenchmarkAlphaReduce8(b *testing.B) {
	benchmarkAlphaReduce(8, b)
}
func BenchmarkAlphaReduce4(b *testing.B) {
	benchmarkAlphaReduce(4, b)
}
func BenchmarkAlphaReduce2(b *testing.B) {
	benchmarkAlphaReduce(2, b)
}
func BenchmarkAlphaReduce1(b *testing.B) {
	benchmarkAlphaReduce(1, b)
}
func benchmarkAlphaReduce(nbChar int, b *testing.B) {
	red := getAlphaReduceFunc(nbChar)
	h := make([]byte, 16, 16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h = red(i, h, h)
	}
}
