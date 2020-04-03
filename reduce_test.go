package rainbow

import (
	"strings"
	"testing"
)

func TestAlphaReduce(t *testing.T) {
	n := 12
	r := GetAlphaReduceFunc(n)
	h := []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 41}
	p := []byte{}
	s := string(r(2, h, p))
	if len(s) != n {
		t.Log(s)
		t.Fatal("unexpected password length")
	}

	n = 7
	r = GetAlphaReduceFunc(n)
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
	r = GetAlphaReduceFunc(n)
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
	red := GetAlphaReduceFunc(nbChar)
	h := make([]byte, 16, 16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h = red(i, h, h)
	}
}
