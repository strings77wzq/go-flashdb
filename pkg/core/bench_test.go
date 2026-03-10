package core

import (
	"strconv"
	"testing"
)

func BenchmarkDB_Set(b *testing.B) {
	db := NewDB(0)
	key := []byte("benchmark_key")
	value := []byte("benchmark_value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.SetString(string(key), value)
	}
}

func BenchmarkDB_Get(b *testing.B) {
	db := NewDB(0)
	key := "benchmark_key"
	value := []byte("benchmark_value")
	db.SetString(key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetStringData(key)
	}
}

func BenchmarkDB_SetParallel(b *testing.B) {
	db := NewDB(0)
	value := []byte("benchmark_value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + strconv.Itoa(i)
			db.SetString(key, value)
			i++
		}
	})
}

func BenchmarkDB_GetParallel(b *testing.B) {
	db := NewDB(0)

	for i := 0; i < 10000; i++ {
		key := "key_" + strconv.Itoa(i)
		value := []byte("value_" + strconv.Itoa(i))
		db.SetString(key, value)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + strconv.Itoa(i%10000)
			db.GetStringData(key)
			i++
		}
	})
}

func BenchmarkDB_MSet(b *testing.B) {
	db := NewDB(0)
	args := make([][]byte, 20)
	for i := 0; i < 20; i += 2 {
		args[i] = []byte("key_" + strconv.Itoa(i))
		args[i+1] = []byte("value_" + strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(args); j += 2 {
			db.SetString(string(args[j]), args[j+1])
		}
	}
}

func BenchmarkDB_MGet(b *testing.B) {
	db := NewDB(0)

	for i := 0; i < 10; i++ {
		key := "key_" + strconv.Itoa(i)
		value := []byte("value_" + strconv.Itoa(i))
		db.SetString(key, value)
	}

	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			db.GetStringData(key)
		}
	}
}

func BenchmarkDB_Incr(b *testing.B) {
	db := NewDB(0)
	key := "counter"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		incrBy(db, key, 1)
	}
}

func BenchmarkDB_Del(b *testing.B) {
	db := NewDB(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "del_key_" + strconv.Itoa(i)
		db.SetString(key, []byte("value"))
		db.data.Delete(key)
	}
}

func BenchmarkConcurrentDict_Set(b *testing.B) {
	dict := NewConcurrentDict()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + strconv.Itoa(i)
			dict.Set(key, []byte("value"))
			i++
		}
	})
}

func BenchmarkConcurrentDict_Get(b *testing.B) {
	dict := NewConcurrentDict()

	for i := 0; i < 10000; i++ {
		key := "key_" + strconv.Itoa(i)
		dict.Set(key, []byte("value"))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + strconv.Itoa(i%10000)
			dict.Get(key)
			i++
		}
	})
}
