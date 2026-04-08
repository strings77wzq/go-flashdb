package benchmark

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Benchmark struct {
	host       string
	port       int
	concurrent int
	requests   int
	pipeline   int
	totalDur   int64
	dataSize   int
}

type Result struct {
	Requests   int
	Duration   time.Duration
	QPS        float64
	LatencyMin time.Duration
	LatencyMax time.Duration
	LatencyAvg time.Duration
	LatencyP50 time.Duration
	LatencyP95 time.Duration
	LatencyP99 time.Duration
	Errors     int
}

func NewBenchmark(host string, port int, concurrent int, requests int, pipeline int) *Benchmark {
	return &Benchmark{
		host:       host,
		port:       port,
		concurrent: concurrent,
		requests:   requests,
		pipeline:   pipeline,
	}
}

func (b *Benchmark) RunPing() Result {
	return b.runCommand("PING", nil, "PING")
}

func (b *Benchmark) SetDataSize(size int) {
	b.dataSize = size
}

func (b *Benchmark) RunSet() Result {
	key := "bench:key"
	value := generateRandomString(b.dataSize)
	return b.runCommand("SET", []string{key, value}, "SET %s %s", key, value)
}

func (b *Benchmark) RunGet() Result {
	key := "bench:key"
	b.runCommand("SET", []string{key, "value"}, "SET %s value", key)
	time.Sleep(10 * time.Millisecond)
	return b.runCommand("GET", []string{key}, "GET %s", key)
}

func (b *Benchmark) RunLPush() Result {
	key := "bench:list"
	value := generateRandomString(10)
	return b.runCommand("LPUSH", []string{key, value}, "LPUSH %s %s", key, value)
}

func (b *Benchmark) RunHSet() Result {
	key := "bench:hash"
	field := "field"
	value := generateRandomString(20)
	return b.runCommand("HSET", []string{key, field, value}, "HSET %s %s %s", key, field, value)
}

func (b *Benchmark) runCommand(cmd string, args []string, format string, formatArgs ...interface{}) Result {
	start := time.Now()

	var wg sync.WaitGroup
	var totalRequests int64
	var totalErrors int64
	var latencies []time.Duration
	var latenciesMu sync.Mutex

	requestsPerClient := b.requests / b.concurrent
	if requestsPerClient == 0 {
		requestsPerClient = 1
	}

	for i := 0; i < b.concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.host, b.port))
			if err != nil {
				atomic.AddInt64(&totalErrors, int64(requestsPerClient))
				return
			}
			defer conn.Close()

			for j := 0; j < requestsPerClient; j++ {
				reqStart := time.Now()

				request := buildRequest(cmd, args)
				_, err := conn.Write(request)
				if err != nil {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}

				respBuf := make([]byte, 4096)
				_, err = conn.Read(respBuf)
				if err != nil {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}

				latency := time.Since(reqStart)
				latenciesMu.Lock()
				latencies = append(latencies, latency)
				latenciesMu.Unlock()

				atomic.AddInt64(&totalRequests, 1)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	errors := int(totalErrors)
	completed := int(totalRequests)

	latencyP50, latencyP95, latencyP99 := calculatePercentiles(latencies)
	avgLatency := calculateAvg(latencies)

	return Result{
		Requests:   completed,
		Duration:   duration,
		QPS:        float64(completed) / duration.Seconds(),
		LatencyMin: getMin(latencies),
		LatencyMax: getMax(latencies),
		LatencyAvg: avgLatency,
		LatencyP50: latencyP50,
		LatencyP95: latencyP95,
		LatencyP99: latencyP99,
		Errors:     errors,
	}
}

func buildRequest(cmd string, args []string) []byte {
	var req []byte
	req = append(req, '*')
	req = append(req, []byte(strconv.Itoa(len(args)+1))...)
	req = append(req, '\r', '\n')

	req = append(req, '$')
	req = append(req, []byte(strconv.Itoa(len(cmd)))...)
	req = append(req, '\r', '\n')
	req = append(req, []byte(cmd)...)
	req = append(req, '\r', '\n')

	for _, arg := range args {
		req = append(req, '$')
		req = append(req, []byte(strconv.Itoa(len(arg)))...)
		req = append(req, '\r', '\n')
		req = append(req, []byte(arg)...)
		req = append(req, '\r', '\n')
	}

	return req
}

func calculatePercentiles(latencies []time.Duration) (p50, p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	quickSort(sorted, 0, len(sorted)-1)

	p50Idx := int(float64(len(sorted)) * 0.50)
	p95Idx := int(float64(len(sorted)) * 0.95)
	p99Idx := int(float64(len(sorted)) * 0.99)

	if p50Idx >= len(sorted) {
		p50Idx = len(sorted) - 1
	}
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return sorted[p50Idx], sorted[p95Idx], sorted[p99Idx]
}

func calculateAvg(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var total int64
	for _, l := range latencies {
		total += l.Nanoseconds()
	}
	return time.Duration(total / int64(len(latencies)))
}

func getMin(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	min := latencies[0]
	for _, l := range latencies {
		if l < min {
			min = l
		}
	}
	return min
}

func getMax(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	max := latencies[0]
	for _, l := range latencies {
		if l > max {
			max = l
		}
	}
	return max
}

func quickSort(arr []time.Duration, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func partition(arr []time.Duration, low, high int) int {
	pivot := arr[high]
	i := low - 1
	for j := low; j < high; j++ {
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

func (r *Result) String() string {
	return fmt.Sprintf(`Summary:
  Requests completed: %d
  Requests failed: %d
  Total duration: %.2f seconds
  QPS: %.2f
  Latency min: %.2f ms
  Latency max: %.2f ms
  Latency avg: %.2f ms
  Latency P50: %.2f ms
  Latency P95: %.2f ms
  Latency P99: %.2f ms`,
		r.Requests,
		r.Errors,
		r.Duration.Seconds(),
		r.QPS,
		float64(r.LatencyMin.Microseconds())/1000,
		float64(r.LatencyMax.Microseconds())/1000,
		float64(r.LatencyAvg.Microseconds())/1000,
		float64(r.LatencyP50.Microseconds())/1000,
		float64(r.LatencyP95.Microseconds())/1000,
		float64(r.LatencyP99.Microseconds())/1000,
	)
}

func ParseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func Round(x float64) float64 {
	return math.Round(x*100) / 100
}
