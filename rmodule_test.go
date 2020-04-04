package rainbow

import (
	"fmt"
	"math/big"
	"testing"
)

func TestExtract(t *testing.T) {

	bb := new(big.Int).SetInt64(62)
	n := new(big.Int).SetInt64(3)
	v := new(big.Int)
	vv := 0

	bb, vv = extract(bb, n, v)
	if vv != 1 && bb.Cmp(new(big.Int).SetInt64(20)) != 0 {
		t.FailNow()
	}

	bb, vv = extract(bb, n, v)
	if vv != 2 && bb.Cmp(new(big.Int).SetInt64(6)) != 0 {
		t.FailNow()
	}
}

func TestRBuilder1(t *testing.T) {
	min, max := 2, 5
	red := NewRBuilder().CompileAlphabet("Aéi", min, max).Build()
	p, h := []byte{}, []byte{2, 5, 12, 6, 54, 44, 55, 89, 7, 65, 46, 5, 4}
	results := make(map[string]int)
	for i := 0; i < 10000; i++ {
		p = red(i, h, p)
		results[string(p)]++
	}
	if len(results) != 3*3+3*3*3+3*3*3*3+3*3*3*3*3 {
		fmt.Printf("min %d max %d results len %d\n%+v\n", min, max, len(results), results)
		t.Fatal("unexpected result length")
	}
}

func TestRBuilder2(t *testing.T) {
	min, max := 3, 3
	red := NewRBuilder().CompileAlphabet("aécd", min, max).Build()
	p, h := []byte{}, []byte{2, 5, 12, 6, 54, 44, 55, 89, 7, 65, 46, 5, 4}
	results := make(map[string]int)
	for i := 0; i < 10000; i++ {
		p = red(i, h, p)
		results[string(p)]++
	}
	if len(results) != 4*4*4 {
		fmt.Printf("min %d max %d results len %d\n%+v\n", min, max, len(results), results)
		t.Fatal("unexpected result length")
	}
}

func BenchmarkRBuilderAlphabet(b *testing.B) {
	min, max := 7, 12
	red := NewRBuilder().CompileAlphabet("abcdefghijklmnopqrstuvwxyz", min, max).Build()
	p, h := []byte{}, []byte{2, 5,
		12, 6, 54, 44, 55, 89, 7,
		12, 6, 54, 44, 55, 89, 7,
		12, 6, 54, 44, 55, 89, 7,
		12, 6, 54, 44, 55, 89, 7,
		65, 46, 5, 4}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p = red(i, h, p)
	}

}
