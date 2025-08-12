package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var (
	threads    int
	cpuPercent float64 // 0-100 (% of one core) used during the request's time budget
	memMB      int     // 0-1024 (MB) to allocate per request
	timeMS     int     // 0-1000 (ms) total request duration
	jitter     float64 // 0.0-1.0 (+/- up to 100% at 1.0)
)

func init() {
	flag.IntVar(&threads, "threads", 1, "Number of OS threads (GOMAXPROCS)")
	flag.Float64Var(&cpuPercent, "cpu", 0, "Target CPU utilization per request (0-100) of one core")
	flag.IntVar(&memMB, "mem", 0, "Memory to allocate per request in MB (0-1024)")
	flag.IntVar(&timeMS, "time", 0, "Total time per request in ms (0-1000)")
	flag.Float64Var(&jitter, "jitter", 0, "Jitter factor (0-1.0), applied +/- per request")
	flag.Parse()
}

func main() {
	runtime.GOMAXPROCS(threads)
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", handler)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	addr := ":8080"
	fmt.Printf("Listening on %s | threads=%d cpu=%.1f%% mem=%dMB time=%dms jitter=%.2f\n",
		addr, threads, cpuPercent, memMB, timeMS, jitter)

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Apply jitter per request
	cpuJ := clampFloat(applyJitter(cpuPercent, jitter), 0, 100)
	memJ := int(clampFloat(applyJitter(float64(memMB), jitter), 0, 1024))
	timeJms := int(math.Round(clampFloat(applyJitter(float64(timeMS), jitter), 0, 60_000))) // guard upper bound

	start := time.Now()

	// Allocate & touch memory (per-request)
	var memBlock []byte
	if memJ > 0 {
		memBlock = make([]byte, memJ*1024*1024)
		touchPages(memBlock) // touch one byte per page to commit without huge overhead
	}

	// Combined latency + CPU utilization
	total := time.Duration(timeJms) * time.Millisecond
	cpuLoadForDuration(cpuJ, total)

	elapsed := time.Since(start)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w,
		"OK | target_time=%dms target_cpu=%.1f%% target_mem=%dMB jitter=%.2f | elapsed=%dms\n",
		timeJms, cpuJ, memJ, jitter, int(elapsed.Milliseconds()))

	// Prevent optimization
	runtime.KeepAlive(memBlock)
}

// Keep ~percent CPU busy over 'total' duration, returning at ~'total' wall time.
func cpuLoadForDuration(percent float64, total time.Duration) {
	if total <= 0 {
		return
	}
	switch {
	case percent <= 0:
		time.Sleep(total)
		return
	case percent >= 100:
		busyFor(total)
		return
	}

	// Use short windows to approximate utilization.
	// Aim for at least ~10 slices; cap typical slice at 10ms.
	window := 10 * time.Millisecond
	if total < window {
		window = total
	}

	var elapsed time.Duration
	for elapsed < total {
		remaining := total - elapsed
		win := window
		if remaining < win {
			win = remaining
		}

		busy := time.Duration(float64(win) * (percent / 100.0))
		idle := win - busy

		if busy > 0 {
			busyFor(busy)
		}
		if idle > 0 {
			time.Sleep(idle)
		}
		elapsed += win
	}
}

// Burn CPU for ~d by doing floating-point work in a tight loop.
func busyFor(d time.Duration) {
	deadline := time.Now().Add(d)
	x := 0.0001
	for time.Now().Before(deadline) {
		// a bit of math to keep the optimizer honest
		x = x*1.0000001 + math.Sqrt(x+1.2345)
		if x > 1e9 {
			x = 0.0001
		}
	}
	_ = x
}

func applyJitter(value, jitter float64) float64 {
	if jitter <= 0 {
		return value
	}
	scale := 1 + (rand.Float64()*2-1)*jitter // 1 ± jitter
	return value * scale
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Touch one byte per 4KiB page to force physical allocation without writing the whole buffer.
func touchPages(b []byte) {
	const page = 4096
	for i := 0; i < len(b); i += page {
		b[i] = 1
	}
	if len(b) > 0 {
		b[len(b)-1] = 1
	}
}
