package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ==========================================
// QRIS Legacy Load Tester
// Simulates concurrent users to create
// real database contention & latency
// ==========================================

type Stats struct {
	TotalRequests   int64
	SuccessCount    int64
	ErrorCount      int64
	Latencies       []time.Duration
	mu              sync.Mutex
}

func (s *Stats) Record(latency time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Latencies = append(s.Latencies, latency)
	if err != nil {
		atomic.AddInt64(&s.ErrorCount, 1)
	} else {
		atomic.AddInt64(&s.SuccessCount, 1)
	}
	atomic.AddInt64(&s.TotalRequests, 1)
}

func (s *Stats) Report(label string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Latencies) == 0 {
		fmt.Printf("\n📊 %s: No requests completed\n", label)
		return
	}

	sort.Slice(s.Latencies, func(i, j int) bool {
		return s.Latencies[i] < s.Latencies[j]
	})

	total := len(s.Latencies)
	var sum time.Duration
	for _, l := range s.Latencies {
		sum += l
	}

	avg := sum / time.Duration(total)
	p50 := s.Latencies[total*50/100]
	p90 := s.Latencies[total*90/100]
	p95 := s.Latencies[total*95/100]
	p99 := s.Latencies[total*99/100]
	min := s.Latencies[0]
	max := s.Latencies[total-1]

	fmt.Printf("\n📊 %s Results:\n", label)
	fmt.Printf("   Total Requests : %d\n", total)
	fmt.Printf("   ✅ Success      : %d\n", s.SuccessCount)
	fmt.Printf("   ❌ Errors       : %d\n", s.ErrorCount)
	fmt.Printf("   ⏱  Avg Latency  : %s\n", avg)
	fmt.Printf("   ⏱  Min Latency  : %s\n", min)
	fmt.Printf("   ⏱  Max Latency  : %s\n", max)
	fmt.Printf("   📈 P50          : %s\n", p50)
	fmt.Printf("   📈 P90          : %s\n", p90)
	fmt.Printf("   📈 P95          : %s\n", p95)
	fmt.Printf("   📈 P99          : %s\n", p99)
	fmt.Printf("   🔥 Throughput   : %.1f req/s\n", float64(total)/sum.Seconds()*float64(total)/float64(total))
}

var client = &http.Client{
	Timeout: 30 * time.Second,
}

// createTransaction - POST /api/legacy/transactions/scan
func createTransaction(baseURL string, merchantID int, amount int64) (int, error) {
	body := map[string]interface{}{
		"qr_payload":  fmt.Sprintf("loadtest-%d-%d", merchantID, time.Now().UnixNano()),
		"merchant_id": merchantID,
		"amount":      amount,
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := client.Post(baseURL+"/transactions/scan", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if txnID, ok := data["transaction_id"].(float64); ok {
			return int(txnID), nil
		}
	}
	return 0, fmt.Errorf("unexpected response: %s", string(respBody))
}

// getTransactionStatus - GET /api/legacy/transactions/:id
func getTransactionStatus(baseURL string, txnID int) error {
	resp, err := client.Get(fmt.Sprintf("%s/transactions/%d", baseURL, txnID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	return nil
}

// confirmPayment - POST /api/legacy/transactions/:id/confirm
func confirmPayment(baseURL string, txnID int) error {
	resp, err := client.Post(fmt.Sprintf("%s/transactions/%d/confirm", baseURL, txnID), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	return nil
}

// getMerchants - GET /api/legacy/merchants
func getMerchants(baseURL string) error {
	resp, err := client.Get(baseURL + "/merchants")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	return nil
}

// generateQRIS - GET /api/legacy/qris
func generateQRIS(baseURL string, merchantID int, amount int) error {
	resp, err := client.Get(fmt.Sprintf("%s/qris?merchant_id=%d&amount=%d", baseURL, merchantID, amount))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	return nil
}

func runLoadTest(baseURL string, concurrency int, duration time.Duration, mode string) {
	fmt.Printf("\n🚀 Starting Load Test\n")
	fmt.Printf("   Mode        : %s\n", mode)
	fmt.Printf("   Concurrency : %d virtual users\n", concurrency)
	fmt.Printf("   Duration    : %s\n", duration)
	fmt.Printf("   Target      : %s\n", baseURL)
	fmt.Println("   ─────────────────────────────────")

	createStats := &Stats{}
	queryStats := &Stats{}
	confirmStats := &Stats{}
	mixedStats := &Stats{}

	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// Track created transaction IDs for querying/confirming
	var txnIDs []int
	var txnMu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopCh:
					return
				default:
				}

				switch mode {
				case "create":
					// Pure transaction creation flood
					start := time.Now()
					txnID, err := createTransaction(baseURL, rand.Intn(2)+1, int64(rand.Intn(100)*1000+1000))
					latency := time.Since(start)
					createStats.Record(latency, err)
					if err == nil {
						txnMu.Lock()
						txnIDs = append(txnIDs, txnID)
						txnMu.Unlock()
					}

				case "query":
					// Pure read flood
					txnMu.Lock()
					if len(txnIDs) == 0 {
						txnMu.Unlock()
						// Create one first
						txnID, _ := createTransaction(baseURL, 1, 10000)
						if txnID > 0 {
							txnMu.Lock()
							txnIDs = append(txnIDs, txnID)
							txnMu.Unlock()
						}
						continue
					}
					txnID := txnIDs[rand.Intn(len(txnIDs))]
					txnMu.Unlock()

					start := time.Now()
					err := getTransactionStatus(baseURL, txnID)
					latency := time.Since(start)
					queryStats.Record(latency, err)

				case "mixed":
					// Realistic mixed workload
					roll := rand.Intn(100)
					var start time.Time
					var err error

					if roll < 30 {
						// 30% create new transactions
						start = time.Now()
						txnID, e := createTransaction(baseURL, rand.Intn(2)+1, int64(rand.Intn(100)*1000+1000))
						err = e
						if e == nil {
							txnMu.Lock()
							txnIDs = append(txnIDs, txnID)
							txnMu.Unlock()
						}
					} else if roll < 60 {
						// 30% query status
						txnMu.Lock()
						if len(txnIDs) > 0 {
							txnID := txnIDs[rand.Intn(len(txnIDs))]
							txnMu.Unlock()
							start = time.Now()
							err = getTransactionStatus(baseURL, txnID)
						} else {
							txnMu.Unlock()
							start = time.Now()
							err = getMerchants(baseURL)
						}
					} else if roll < 75 {
						// 15% confirm payment
						txnMu.Lock()
						if len(txnIDs) > 0 {
							txnID := txnIDs[rand.Intn(len(txnIDs))]
							txnMu.Unlock()
							start = time.Now()
							err = confirmPayment(baseURL, txnID)
						} else {
							txnMu.Unlock()
							start = time.Now()
							err = getMerchants(baseURL)
						}
					} else if roll < 90 {
						// 15% generate QRIS
						start = time.Now()
						err = generateQRIS(baseURL, rand.Intn(2)+1, rand.Intn(100)*1000+1000)
					} else {
						// 10% get merchants
						start = time.Now()
						err = getMerchants(baseURL)
					}

					latency := time.Since(start)
					mixedStats.Record(latency, err)

				case "confirm":
					// Create then immediately confirm (full flow)
					start := time.Now()
					txnID, err := createTransaction(baseURL, rand.Intn(2)+1, int64(rand.Intn(100)*1000+1000))
					if err == nil {
						err = confirmPayment(baseURL, txnID)
					}
					latency := time.Since(start)
					confirmStats.Record(latency, err)
				}
			}
		}(i)
	}

	// Progress ticker
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				total := atomic.LoadInt64(&createStats.TotalRequests) +
					atomic.LoadInt64(&queryStats.TotalRequests) +
					atomic.LoadInt64(&mixedStats.TotalRequests) +
					atomic.LoadInt64(&confirmStats.TotalRequests)
				errors := atomic.LoadInt64(&createStats.ErrorCount) +
					atomic.LoadInt64(&queryStats.ErrorCount) +
					atomic.LoadInt64(&mixedStats.ErrorCount) +
					atomic.LoadInt64(&confirmStats.ErrorCount)
				fmt.Printf("   ⏳ Progress: %d requests sent (%d errors)...\n", total, errors)
			}
		}
	}()

	// Run for specified duration
	time.Sleep(duration)
	close(stopCh)
	ticker.Stop()
	wg.Wait()

	// Print results
	fmt.Println("\n═══════════════════════════════════════")
	fmt.Println("        LOAD TEST RESULTS")
	fmt.Println("═══════════════════════════════════════")

	switch mode {
	case "create":
		createStats.Report("Transaction Creation")
	case "query":
		queryStats.Report("Transaction Status Query")
	case "mixed":
		mixedStats.Report("Mixed Workload")
	case "confirm":
		confirmStats.Report("Full Flow (Create + Confirm)")
	}

	// Print DB state
	fmt.Println("\n💾 Database Impact:")
	txnMu.Lock()
	fmt.Printf("   Transactions created during test: %d\n", len(txnIDs))
	txnMu.Unlock()
}

func main() {
	concurrency := flag.Int("c", 50, "Number of concurrent virtual users")
	duration := flag.Int("d", 30, "Test duration in seconds")
	mode := flag.String("mode", "mixed", "Test mode: create|query|mixed|confirm")
	host := flag.String("host", "localhost", "Backend host")
	port := flag.String("port", "8081", "Backend port")

	flag.Parse()

	baseURL := fmt.Sprintf("http://%s:%s/api/legacy", *host, *port)

	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║   🔥 QRIS Legacy Load Tester v1.0     ║")
	fmt.Println("║   Simulating real-world traffic        ║")
	fmt.Println("╚═══════════════════════════════════════╝")

	// Verify backend is reachable
	resp, err := http.Get(baseURL + "/merchants")
	if err != nil {
		fmt.Printf("❌ Cannot reach backend at %s: %v\n", baseURL, err)
		os.Exit(1)
	}
	resp.Body.Close()
	fmt.Printf("✅ Backend reachable at %s\n", baseURL)

	runLoadTest(baseURL, *concurrency, time.Duration(*duration)*time.Second, *mode)
}
