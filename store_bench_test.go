package main

import (
	"fmt"
	"testing"
)

func BenchmarkPut(b *testing.B) {
	s := NewStore(nil)

	for n := 0; n < b.N; n++ {
		s.Put(fmt.Sprintf("foo-%d", n), fmt.Sprintf("bar-%d", n), n)
	}
}
