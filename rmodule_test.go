package rainbow

import (
	"crypto"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestRBuilder1(t *testing.T) {
	min, max := 2, 5
	red := New(crypto.MD5, 10).CompileAlphabet("Aéi", min, max).buildReduce()
	p, h := []byte{}, make([]byte, 15)
	rd := rand.New(rand.NewSource(42))
	results := make(map[string]int)
	for i := 0; i < 10_000; i++ {
		rd.Read(h)
		p = red(i, h, p)
		results[string(p)]++
	}
	if len(results) != 3*3+3*3*3+3*3*3*3+3*3*3*3*3 { // 360
		fmt.Printf("min %d max %d results len %d\n%+v\n", min, max, len(results), results)
		t.Fatal("unexpected result length distributon")
	}
}

func TestRBuilder2(t *testing.T) {
	min, max := 3, 3
	red := New(crypto.MD5, 10).CompileAlphabet("aécd", min, max).buildReduce()
	p, h := []byte{}, make([]byte, 15)
	rd := rand.New(rand.NewSource(42))
	results := make(map[string]int)
	for i := 0; i < 10_000; i++ {
		rd.Read(h)
		p = red(i, h, p)
		results[string(p)]++
	}
	if len(results) != 4*4*4 {
		fmt.Printf("min %d max %d results len %d\n%+v\n", min, max, len(results), results)
		t.Fatal("unexpected result length")
	}
}
func TestRBuilder3AutoExtend(t *testing.T) {
	min, max := 3, 200
	red := New(crypto.MD5, 10).CompileAlphabet("abc", min, max).buildReduce()
	p, h := []byte{}, make([]byte, 15)
	rd := rand.New(rand.NewSource(42))
	results := make(map[string]int)
	for i := 0; i < 100; i++ {
		rd.Read(h)
		p = red(i, h, p)
		results[string(p)]++
	}

	if len(results) < 100 {
		fmt.Println("autoextended : ", len(results), results)
		fmt.Printf("min %d max %d results len %d\n", min, max, len(results))
		t.Fatal("unexpected collisions")
	}
}

func BenchmarkRBuilderAlphabet(b *testing.B) {
	min, max := 7, 12
	red := New(crypto.MD5, 10).CompileAlphabet("abcdefghijklmnopqrstuvwxyz", min, max).buildReduce()
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

func TestVisualCompileWords(t *testing.T) {

	r := New(crypto.MD5, 10).
		CompileWordList("words_test.txt").
		CompileAlphabet("0123456789", 0, 2).
		Build()

	n := 20
	h := []byte{1, 3, 55, 6, 4, 44, 55, 88, 99, 77, 22, 33, 11, 121, 65, 4, 5, 5, 55, 4, 4}
	fmt.Println("==============================")
	fmt.Println("List of words concatened with 0 to 2 digits")
	for i := 0; i < n; i++ {
		p := r.rf(i, h, []byte{})
		fmt.Println(string(p))
	}
	fmt.Println("==============================")
}

func TestVisualCompileTransform(t *testing.T) {

	trf := func(p []byte) []byte {
		return []byte("**" + strings.ToUpper(string(p)) + "**")
	}

	r := New(crypto.MD5, 10).
		CompileWordList("words_test.txt").
		CompileTransform(trf, nil, nil).
		Build()

	n := 1000
	h := make([]byte, 16)
	rand.Read(h)

	capi := 0

	fmt.Println("List of words, 1/3rd of them capitalized ")
	for i := 0; i < n; i++ {
		p := r.rf(i, h, []byte{})
		if p[0] == byte('*') {
			capi++
		}
	}
	ratio := float64(capi) / float64(n)
	fmt.Printf("Capitalized Actual : %2.1f%%\t Target : %2.1f%%\n", 100.*ratio, 100./3.)
	if ratio < .3 || ratio > 0.35 {
		t.Fatal("unexpected ratio of capitalized letters")
	}

}
