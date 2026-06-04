package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ── Configuration ──────────────────────────────────────────────────────────

var baseURL = "http://localhost:8080"

type loadProfile struct {
	Name       string
	Users      int
	Duration   time.Duration
	MinDelayMs int64
	MaxDelayMs int64
}

var profiles = []loadProfile{
	{"🟢 Light Load", 10, 30 * time.Second, 200, 800},
	{"🟡 Medium Load", 50, 30 * time.Second, 500, 2000},
	{"🔴 Heavy Load", 100, 60 * time.Second, 1000, 4000},
	{"💀 Extreme Load", 200, 60 * time.Second, 2000, 8000},
	{"📊 Quick Benchmark", 50, 15 * time.Second, 0, 0},
}

// ── Metrics ────────────────────────────────────────────────────────────────

type metrics struct {
	totalRequests  atomic.Int64
	successCount   atomic.Int64
	errorCount     atomic.Int64
	totalLatencyNs atomic.Int64
	minLatencyNs   atomic.Int64
	maxLatencyNs   atomic.Int64
	mu             sync.Mutex
	latencies      []int64
}

func newMetrics() *metrics {
	m := &metrics{}
	m.minLatencyNs.Store(1<<63 - 1)
	return m
}

func (m *metrics) record(d time.Duration, ok bool) {
	ns := d.Nanoseconds()
	m.totalRequests.Add(1)
	m.totalLatencyNs.Add(ns)
	if ok {
		m.successCount.Add(1)
	} else {
		m.errorCount.Add(1)
	}
	for {
		old := m.minLatencyNs.Load()
		if ns >= old || m.minLatencyNs.CompareAndSwap(old, ns) {
			break
		}
	}
	for {
		old := m.maxLatencyNs.Load()
		if ns <= old || m.maxLatencyNs.CompareAndSwap(old, ns) {
			break
		}
	}
	m.mu.Lock()
	m.latencies = append(m.latencies, ns)
	m.mu.Unlock()
}

func (m *metrics) p95() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.latencies)
	if n == 0 {
		return 0
	}
	sorted := make([]int64, n)
	copy(sorted, m.latencies)
	// simple insertion sort is fine for report
	for i := 1; i < n; i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	idx := int(float64(n) * 0.95)
	if idx >= n {
		idx = n - 1
	}
	return time.Duration(sorted[idx])
}

// ── HTTP helpers ───────────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 30 * time.Second}

func apiGet(path string) ([]byte, int, error) {
	resp, err := httpClient.Get(baseURL + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return b, resp.StatusCode, nil
}

func apiPost(path string, body interface{}) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		j, _ := json.Marshal(body)
		reader = bytes.NewReader(j)
	}
	resp, err := httpClient.Post(baseURL+path, "application/json", reader)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return b, resp.StatusCode, nil
}

// ── Merchant discovery ─────────────────────────────────────────────────────

type merchantInfo struct {
	ID   string
	QRID string
	Name string
}

func fetchMerchant() (*merchantInfo, error) {
	b, code, err := apiGet("/api/merchants")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", code, string(b))
	}
	var resp map[string]interface{}
	json.Unmarshal(b, &resp)
	var list []interface{}
	if v, ok := resp["merchants"]; ok {
		list, _ = v.([]interface{})
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no merchants found — is backend seeded?")
	}
	m := list[0].(map[string]interface{})
	id := strVal(m, "ID", "id")
	qrid := strVal(m, "QRID", "qr_id")
	name := strVal(m, "MerchantName", "merchant_name")
	return &merchantInfo{ID: id, QRID: qrid, Name: name}, nil
}

func strVal(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// ── Generate QRIS payload ──────────────────────────────────────────────────

func generateQRIS(merchantID string, amount int) (string, error) {
	b, code, err := apiGet(fmt.Sprintf("/api/qris?merchant_id=%s&amount=%d", merchantID, amount))
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("HTTP %d", code)
	}
	var resp map[string]interface{}
	json.Unmarshal(b, &resp)
	p, _ := resp["qris_payload"].(string)
	return p, nil
}

// ── Stress mode control ────────────────────────────────────────────────────

func enableStress(minMs, maxMs int64) error {
	_, code, err := apiPost("/api/stress/enable", map[string]int64{
		"min_delay_ms": minMs, "max_delay_ms": maxMs,
	})
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("HTTP %d", code)
	}
	return nil
}

func disableStress() error {
	_, code, err := apiPost("/api/stress/disable", nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("HTTP %d", code)
	}
	return nil
}

// ── Load test runner ───────────────────────────────────────────────────────

func runLoadTest(p loadProfile) {
	fmt.Printf("\n%s\n", repeat("═", 60))
	fmt.Printf("  %s\n", p.Name)
	fmt.Printf("  Users: %d | Duration: %s\n", p.Users, p.Duration)
	if p.MinDelayMs > 0 {
		fmt.Printf("  Stress Delay: %d–%d ms\n", p.MinDelayMs, p.MaxDelayMs)
	}
	fmt.Printf("%s\n\n", repeat("═", 60))

	// Discover merchant
	merchant, err := fetchMerchant()
	if err != nil {
		fmt.Printf("❌ Cannot discover merchant: %v\n", err)
		return
	}
	fmt.Printf("✅ Merchant: %s (%s)\n", merchant.Name, merchant.ID)

	// Generate a QRIS payload
	qrPayload, err := generateQRIS(merchant.ID, 10000)
	if err != nil {
		fmt.Printf("❌ Cannot generate QRIS: %v\n", err)
		return
	}
	fmt.Printf("✅ QRIS payload ready\n")

	// Enable stress if profile requires it
	if p.MinDelayMs > 0 {
		if err := enableStress(p.MinDelayMs, p.MaxDelayMs); err != nil {
			fmt.Printf("⚠️  Failed to enable stress: %v (continuing anyway)\n", err)
		} else {
			fmt.Printf("✅ Stress mode enabled (%d–%d ms)\n", p.MinDelayMs, p.MaxDelayMs)
		}
	}

	m := newMetrics()
	ctx_done := make(chan struct{})
	var wg sync.WaitGroup

	fmt.Printf("\n🚀 Starting %d concurrent users for %s...\n\n", p.Users, p.Duration)
	start := time.Now()

	for i := 0; i < p.Users; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx_done:
					return
				default:
				}
				doTransaction(merchant, qrPayload, m)
			}
		}()
	}

	// Progress ticker
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ctx_done:
				ticker.Stop()
				return
			case <-ticker.C:
				elapsed := time.Since(start).Seconds()
				total := m.totalRequests.Load()
				errs := m.errorCount.Load()
				rps := float64(total) / elapsed
				fmt.Printf("  ⏱  %.0fs | Requests: %d | Errors: %d | RPS: %.1f\n",
					elapsed, total, errs, rps)
			}
		}
	}()

	time.Sleep(p.Duration)
	close(ctx_done)
	wg.Wait()

	// Disable stress after test
	if p.MinDelayMs > 0 {
		disableStress()
	}

	printReport(m, p, time.Since(start))
}

func doTransaction(merchant *merchantInfo, qrPayload string, m *metrics) {
	// Use the same fixed amount that was used to generate the QRIS payload
	amount := 10000
	payload := map[string]interface{}{
		"qr_payload":  qrPayload,
		"merchant_id": merchant.QRID,
		"amount":      amount,
	}

	// Step 1: Scan
	t0 := time.Now()
	scanBody, scanCode, err := apiPost("/api/transactions/scan", payload)
	scanDur := time.Since(t0)

	if err != nil || scanCode != 201 {
		m.record(scanDur, false)
		return
	}

	var scanResp map[string]interface{}
	json.Unmarshal(scanBody, &scanResp)
	data, _ := scanResp["data"].(map[string]interface{})
	txID, _ := data["transaction_id"].(string)
	if txID == "" {
		m.record(scanDur, false)
		return
	}

	// Step 2: Confirm (alternate async/sync)
	var confirmPath string
	if rand.Intn(2) == 0 {
		confirmPath = fmt.Sprintf("/api/transactions/%s/confirm", txID)
	} else {
		confirmPath = fmt.Sprintf("/api/transactions/%s/confirm-sync", txID)
	}

	t1 := time.Now()
	_, confirmCode, err := apiPost(confirmPath, nil)
	confirmDur := time.Since(t1)

	totalDur := scanDur + confirmDur
	ok := err == nil && confirmCode == 200
	m.record(totalDur, ok)
}

// ── Report ─────────────────────────────────────────────────────────────────

func printReport(m *metrics, p loadProfile, elapsed time.Duration) {
	total := m.totalRequests.Load()
	success := m.successCount.Load()
	errors := m.errorCount.Load()
	avgNs := int64(0)
	if total > 0 {
		avgNs = m.totalLatencyNs.Load() / total
	}

	fmt.Printf("\n%s\n", repeat("═", 60))
	fmt.Printf("  📊 LOAD TEST REPORT\n")
	fmt.Printf("%s\n", repeat("─", 60))
	fmt.Printf("  Profile:       %s\n", p.Name)
	fmt.Printf("  Duration:      %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Concurrency:   %d users\n", p.Users)
	fmt.Printf("%s\n", repeat("─", 60))
	fmt.Printf("  Total Requests:  %d\n", total)
	fmt.Printf("  ✅ Success:      %d (%.1f%%)\n", success, pct(success, total))
	fmt.Printf("  ❌ Errors:       %d (%.1f%%)\n", errors, pct(errors, total))
	fmt.Printf("%s\n", repeat("─", 60))
	fmt.Printf("  Avg Latency:   %s\n", time.Duration(avgNs).Round(time.Millisecond))
	fmt.Printf("  Min Latency:   %s\n", time.Duration(m.minLatencyNs.Load()).Round(time.Millisecond))
	fmt.Printf("  Max Latency:   %s\n", time.Duration(m.maxLatencyNs.Load()).Round(time.Millisecond))
	fmt.Printf("  P95 Latency:   %s\n", m.p95().Round(time.Millisecond))
	rps := float64(total) / elapsed.Seconds()
	fmt.Printf("  Throughput:    %.1f req/s\n", rps)
	fmt.Printf("%s\n\n", repeat("═", 60))
}

func pct(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// ── Menu ───────────────────────────────────────────────────────────────────

func printMenu() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║         QRIS Latency Optimizer — Load Test CLI         ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║  1) 🟢 Light Load    — 10 VUs × 30s  (200–800ms)      ║")
	fmt.Println("║  2) 🟡 Medium Load   — 50 VUs × 30s  (0.5–2s)        ║")
	fmt.Println("║  3) 🔴 Heavy Load    — 100 VUs × 60s (1–4s)          ║")
	fmt.Println("║  4) 💀 Extreme Load  — 200 VUs × 60s (2–8s)          ║")
	fmt.Println("║  5) 📊 Quick Bench   — 50 VUs × 15s  (no stress)     ║")
	fmt.Println("║  6) 🔧 Enable Stress Mode only                        ║")
	fmt.Println("║  7) ✅ Disable Stress Mode                            ║")
	fmt.Println("║  8) 🚪 Exit                                           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Print("\n  Select option [1-8]: ")
}

func main() {
	if v := os.Getenv("BASE_URL"); v != "" {
		baseURL = v
	}

	fmt.Println("\n🔌 Target backend:", baseURL)

	// Quick health check
	if _, code, err := apiGet("/api/ping"); err != nil || code != 200 {
		fmt.Println("⚠️  Backend may not be reachable. Ensure it is running.")
	} else {
		fmt.Println("✅ Backend is healthy")
	}

	for {
		printMenu()
		var choice int
		if _, err := fmt.Scan(&choice); err != nil {
			fmt.Println("Invalid input, try again.")
			continue
		}

		switch choice {
		case 1, 2, 3, 4, 5:
			runLoadTest(profiles[choice-1])
		case 6:
			fmt.Print("  Min delay (ms): ")
			var minD, maxD int64
			fmt.Scan(&minD)
			fmt.Print("  Max delay (ms): ")
			fmt.Scan(&maxD)
			if err := enableStress(minD, maxD); err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
			} else {
				fmt.Printf("✅ Stress mode enabled (%d–%d ms)\n", minD, maxD)
			}
		case 7:
			if err := disableStress(); err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
			} else {
				fmt.Println("✅ Stress mode disabled")
			}
		case 8:
			fmt.Println("\n👋 Goodbye!")
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}
