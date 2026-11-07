// Нагрузочный тест HTTP API сервиса order. Только стандартная библиотека.
//
// Запуск (после task up-all): task load-test
// или вручную: go run ./scripts/loadtest/main.go -session ... -user ... -part ...
//
// Сценарий:
//  1. create — POST /api/v1/orders в N воркеров в течение -duration
//  2. get    — GET  /api/v1/orders/{uuid} по созданным заказам
//  3. pay    — оплата -pay заказов (каждая оплата запускает сборку в assembly: 5-15 с на заказ,
//     поэтому большое число оплат создаёт очередь в Kafka)
//  4. race   — -race параллельных оплат ОДНОГО заказа: корректно, если успешна ровно одна
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
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type config struct {
	base, session, user, part string
	workers, pay, race        int
	duration                  time.Duration
}

type result struct {
	latency time.Duration
	status  int
	body    []byte
}

var client = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        512,
		MaxIdleConnsPerHost: 512,
	},
}

func (c config) do(method, path string, payload any) result {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return result{status: -1}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-Uuid", c.session)

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return result{latency: time.Since(start), status: -1}
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	return result{latency: time.Since(start), status: resp.StatusCode, body: data}
}

func (c config) createOrder() (string, result) {
	r := c.do(http.MethodPost, "/api/v1/orders", map[string]any{
		"user_uuid": c.user, "part_uuids": []string{c.part},
	})

	var out struct {
		OrderUUID string `json:"order_uuid"`
	}
	_ = json.Unmarshal(r.body, &out)

	return out.OrderUUID, r
}

// runFor запускает fn в workers горутин до истечения duration и собирает результаты.
func runFor(workers int, d time.Duration, fn func() result) ([]result, time.Duration) {
	var mu sync.Mutex
	var all []result
	var wg sync.WaitGroup
	deadline := time.Now().Add(d)
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []result
			for time.Now().Before(deadline) {
				local = append(local, fn())
			}
			mu.Lock()
			all = append(all, local...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	return all, time.Since(start)
}

func report(name string, rs []result, elapsed time.Duration) {
	if len(rs) == 0 {
		fmt.Printf("| %s | 0 | - | - | - | - | - |\n", name)
		return
	}

	lat := make([]time.Duration, 0, len(rs))
	codes := map[int]int{}
	for _, r := range rs {
		lat = append(lat, r.latency)
		codes[r.status]++
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	pct := func(p float64) time.Duration { return lat[int(float64(len(lat)-1)*p)] }

	var cs []string
	for code, n := range codes {
		cs = append(cs, fmt.Sprintf("%d×%d", code, n))
	}
	sort.Strings(cs)

	fmt.Printf("| %s | %d | %.0f | %v | %v | %v | %s |\n", name, len(rs),
		float64(len(rs))/elapsed.Seconds(),
		pct(0.50).Round(time.Microsecond*100), pct(0.95).Round(time.Microsecond*100), pct(0.99).Round(time.Microsecond*100),
		strings.Join(cs, " "))
}

func main() {
	var c config
	flag.StringVar(&c.base, "base", "http://localhost:8080", "адрес order")
	flag.StringVar(&c.session, "session", "", "session uuid (из iam Login)")
	flag.StringVar(&c.user, "user", "", "uuid пользователя этой сессии")
	flag.StringVar(&c.part, "part", "", "uuid детали из inventory")
	flag.IntVar(&c.workers, "workers", 20, "параллельных воркеров")
	flag.DurationVar(&c.duration, "duration", 15*time.Second, "длительность фаз create и get")
	flag.IntVar(&c.pay, "pay", 30, "сколько заказов оплатить (каждый = 5-15 с сборки в assembly)")
	flag.IntVar(&c.race, "race", 20, "параллельных оплат одного заказа")
	flag.Parse()

	if c.session == "" || c.user == "" || c.part == "" {
		fmt.Fprintln(os.Stderr, "нужны -session, -user, -part (см. scripts/loadtest/run.sh)")
		os.Exit(2)
	}

	fmt.Printf("workers=%d duration=%v pay=%d race=%d\n\n", c.workers, c.duration, c.pay, c.race)
	fmt.Println("| Фаза | Запросов | RPS | p50 | p95 | p99 | Коды ответа |")
	fmt.Println("|---|---:|---:|---|---|---|---|")

	// 1. create
	var mu sync.Mutex
	var orders []string
	rs, el := runFor(c.workers, c.duration, func() result {
		id, r := c.createOrder()
		if id != "" {
			mu.Lock()
			orders = append(orders, id)
			mu.Unlock()
		}
		return r
	})
	report("create order", rs, el)

	if len(orders) == 0 {
		fmt.Println("\nне создано ни одного заказа, дальше идти нет смысла")
		os.Exit(1)
	}

	// 2. get
	rs, el = runFor(c.workers, c.duration, func() result {
		mu.Lock()
		id := orders[rand.Intn(len(orders))]
		mu.Unlock()
		return c.do(http.MethodGet, "/api/v1/orders/"+id, nil)
	})
	report("get order", rs, el)

	// 3. pay (ограниченное число, чтобы не забить очередь сборки)
	n := min(c.pay, len(orders))
	var idx atomic.Int64
	var payRes []result
	var pmu sync.Mutex
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				k := int(idx.Add(1)) - 1
				if k >= n {
					return
				}
				r := c.do(http.MethodPost, "/api/v1/orders/"+orders[k]+"/pay", map[string]string{"payment_method": "CARD"})
				pmu.Lock()
				payRes = append(payRes, r)
				pmu.Unlock()
			}
		}()
	}
	wg.Wait()
	report(fmt.Sprintf("pay order (%d шт.)", n), payRes, time.Since(start))

	// 4. race: параллельные оплаты одного заказа
	if c.race > 0 {
		id, _ := c.createOrder()
		var raceRes []result
		var rmu sync.Mutex
		gate := make(chan struct{})
		var rwg sync.WaitGroup
		for i := 0; i < c.race; i++ {
			rwg.Add(1)
			go func() {
				defer rwg.Done()
				<-gate
				r := c.do(http.MethodPost, "/api/v1/orders/"+id+"/pay", map[string]string{"payment_method": "CARD"})
				rmu.Lock()
				raceRes = append(raceRes, r)
				rmu.Unlock()
			}()
		}
		start = time.Now()
		close(gate)
		rwg.Wait()
		report(fmt.Sprintf("race: %d оплат одного заказа", c.race), raceRes, time.Since(start))

		succ := 0
		for _, r := range raceRes {
			if r.status >= 200 && r.status < 300 {
				succ++
			}
		}
		fmt.Printf("\nГонка: успешных оплат одного заказа = %d из %d (корректно: 1)\n", succ, c.race)
	}
}
