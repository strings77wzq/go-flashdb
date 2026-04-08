package main

import (
	"flag"
	"fmt"
	"os"

	"goflashdb/pkg/benchmark"
)

func main() {
	host := flag.String("h", "127.0.0.1", "Server host")
	port := flag.Int("p", 6379, "Server port")
	concurrent := flag.Int("c", 50, "Number of concurrent connections")
	requests := flag.Int("n", 10000, "Total number of requests")
	dataSize := flag.Int("d", 10, "Data size for SET commands")
	test := flag.String("t", "ping", "Test type: ping, set, get, lpush, hset")

	flag.Parse()

	bench := benchmark.NewBenchmark(*host, *port, *concurrent, *requests, 1)
	bench.SetDataSize(*dataSize)

	var result benchmark.Result

	switch *test {
	case "ping":
		fmt.Println("Running PING benchmark...")
		result = bench.RunPing()
	case "set":
		fmt.Println("Running SET benchmark...")
		result = bench.RunSet()
	case "get":
		fmt.Println("Running GET benchmark...")
		result = bench.RunGet()
	case "lpush":
		fmt.Println("Running LPUSH benchmark...")
		result = bench.RunLPush()
	case "hset":
		fmt.Println("Running HSET benchmark...")
		result = bench.RunHSet()
	default:
		fmt.Printf("Unknown test type: %s\n", *test)
		os.Exit(1)
	}

	fmt.Println(result.String())
}
