package fimgs

import (
	"runtime"
	"testing"
)

func makeColorArray0(len int) [][]int64 {
	data := make([]int64, len*3)
	res := make([][]int64, len)
	for i := 0; i < len; i++ {
		res[i] = data[i*3 : i*3+3]
	}
	return res
}

func BenchmarkMakeColorArray(b *testing.B) {
	b.Run("make [][]i64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := makeColorArray0(10000)
			runtime.KeepAlive(x)
		}
	})
	b.Run("make [][3]i64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := makeColorArray(10000)
			runtime.KeepAlive(x)
		}
	})
}
