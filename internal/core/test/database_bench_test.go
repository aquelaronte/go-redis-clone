package core_test

import (
	"fmt"
	"go-redis-clone/internal/core"
	"sync/atomic"
	"testing"
)

func benchKey(prefix string, i int64) []byte {
	return fmt.Appendf(nil, "bench:%s:%d", prefix, i)
}

func BenchmarkDatabase(b *testing.B) {
	cases := []struct {
		name string
		run  func(b *testing.B)
	}{
		{"SetSameKey", func(b *testing.B) {
			key := []byte("bench:set:same")
			value := []byte("value")
			for b.Loop() {
				core.SET(key, value)
			}
		}},
		{"SetUniqueKeys", func(b *testing.B) {
			value := []byte("value")
			var i int64
			for b.Loop() {
				core.SET(benchKey("set-unique", i), value)
				i++
			}
		}},
		{"GetHit", func(b *testing.B) {
			key := []byte("bench:get:hit")
			core.SET(key, []byte("value"))
			for b.Loop() {
				core.GET(key)
			}
		}},
		{"GetMiss", func(b *testing.B) {
			key := []byte("bench:get:miss:not-set")
			for b.Loop() {
				core.GET(key)
			}
		}},
		{"GetWithTTL", func(b *testing.B) {
			key := []byte("bench:get:ttl")
			core.SET(key, []byte("value"))
			core.EXPIRE(key, 3600)
			for b.Loop() {
				core.GET(key)
			}
		}},
		// Cycles through a fixed pool of pre-set keys so each iteration deletes
		// a key that exists. The pool is re-populated when exhausted.
		{"Del", func(b *testing.B) {
			const pool = 1024
			value := []byte("value")
			keys := make([][]byte, pool)
			for i := range pool {
				keys[i] = benchKey("del", int64(i))
			}

			repopulate := func() {
				for i := range pool {
					core.SET(keys[i], value)
				}
			}
			repopulate()

			i := 0
			for b.Loop() {
				if i == pool {
					b.StopTimer()
					repopulate()
					i = 0
					b.StartTimer()
				}
				core.DEL(keys[i])
				i++
			}
		}},
		{"SetParallel", func(b *testing.B) {
			var counter int64
			value := []byte("value")
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					id := atomic.AddInt64(&counter, 1)
					core.SET(benchKey("set-parallel", id), value)
				}
			})
		}},
		{"GetParallel", func(b *testing.B) {
			key := []byte("bench:get:parallel")
			core.SET(key, []byte("value"))
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					core.GET(key)
				}
			})
		}},
		// 90% reads / 10% writes against a small key space — exercises
		// the RWMutex read/write upgrade path under contention.
		{"MixedParallel", func(b *testing.B) {
			const keyspace = 64
			value := []byte("value")
			for i := range int64(keyspace) {
				core.SET(benchKey("mixed", i), value)
			}

			var counter int64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					n := atomic.AddInt64(&counter, 1)
					key := benchKey("mixed", n%keyspace)
					if n%10 == 0 {
						core.SET(key, value)
					} else {
						core.GET(key)
					}
				}
			})
		}},
	}

	for _, tc := range cases {
		b.Run(tc.name, tc.run)
	}
}
