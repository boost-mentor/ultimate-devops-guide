package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/bits"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/boost-mentor/kubernetes-production-labs/apps/devops-may-cry/internal/store"
)

type quote struct {
	Text string `json:"quote"`
	Who  string `json:"who"`
}

var quotes = []quote{
	{Text: "Прод упал в три ночи. Как обычно — на моей смене.", Who: "дежурный"},
	{Text: "Заказы на демонов принимаем круглосуточно. Пицца — отдельно.", Who: "вывеска"},
	{Text: "Инцидент закрыт. Счёт выставим утром.", Who: "дежурный"},
	{Text: "Не бывает мелких инцидентов. Бывает поздний вызов.", Who: "дежурный"},
	{Text: "Пока ты читал runbook, я закрыл два тикета.", Who: "дежурный"},
	{Text: "Демоны воют, графики красные — обычный вторник.", Who: "дежурный"},
}

var prices = map[string]int{
	"memory-leak":  500,
	"night-outage": 9000,
	"disk-full":    800,
	"ddos-swarm":   15000,
	"cert-expired": 1200,
	"dns-ghost":    3000,
}

var targets = []string{"memory-leak", "night-outage", "disk-full", "ddos-swarm", "cert-expired", "dns-ghost"}

type Server struct {
	store               store.OrderStore
	logger              *slog.Logger
	pod                 string
	mux                 *http.ServeMux
	labEndpointsEnabled bool

	heldMu sync.Mutex
	held   [][]byte

	requests atomic.Uint64
	orders   atomic.Uint64
}

func New(orderStore store.OrderStore, logger *slog.Logger, pod string, labEndpointsEnabled bool) *Server {
	if pod == "" {
		pod = "local"
	}
	s := &Server{
		store:               orderStore,
		logger:              logger,
		pod:                 pod,
		mux:                 http.NewServeMux(),
		labEndpointsEnabled: labEndpointsEnabled,
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.recoverPanic(s.accessLog(s.mux))
}

func (s *Server) routes() {
	// /{$} is an exact root match. Unknown paths must return 404; otherwise a
	// deliberately broken readiness probe such as /nope would look healthy.
	// The explicit "METHOD /path" patterns also make the public route contract
	// visible in one place instead of hiding it across handler implementations.
	s.mux.HandleFunc("GET /{$}", s.handleRoot)
	s.mux.HandleFunc("GET /quote", s.handleQuote)
	s.mux.HandleFunc("GET /status", s.handleStatus)
	s.mux.HandleFunc("GET /order", s.handleOrder)
	s.mux.HandleFunc("POST /orders", s.handleOrder)
	s.mux.HandleFunc("GET /overload", s.handleOverload)
	s.mux.HandleFunc("GET /wound", s.handleWound)
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /livez", s.handleLive)
	s.mux.HandleFunc("GET /readyz", s.handleReady)
	s.mux.HandleFunc("GET /metrics", s.handleMetrics)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
		return
	}
	s.handleQuote(w, r)
}

func (s *Server) handleQuote(w http.ResponseWriter, _ *http.Request) {
	item := quotes[rand.Intn(len(quotes))]
	writeJSON(w, http.StatusOK, map[string]any{
		"quote": item.Text,
		"who":   item.Who,
		"pod":   s.pod,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "devops-may-cry",
		"pod":     s.pod,
		"store":   s.store.Kind(),
		"time":    time.Now().UTC(),
	})
}

func (s *Server) handleOrder(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("on"))
	if r.Method == http.MethodPost {
		var body struct {
			Target string `json:"target"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		target = strings.TrimSpace(body.Target)
	}
	if target == "" {
		target = targets[rand.Intn(len(targets))]
	}
	price := prices[target]
	if price == 0 {
		price = 2500
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	order, err := s.store.Create(ctx, target, price, s.pod)
	if err != nil {
		s.logger.Error("create order", "error", err, "target", target)
		writeError(w, http.StatusServiceUnavailable, "order store unavailable")
		return
	}
	s.orders.Add(1)
	status := http.StatusOK
	if r.Method == http.MethodPost {
		status = http.StatusCreated
	}
	writeJSON(w, status, order)
}

func (s *Server) handleOverload(w http.ResponseWriter, r *http.Request) {
	if !s.labEndpointsEnabled {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	seconds, err := boundedInt(r.URL.Query().Get("sec"), 60, 1, 600)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for worker := 0; worker < 2; worker++ {
		go burnCPU(deadline, uint64(worker+1))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"overload": "ON",
		"seconds":  seconds,
		"pod":      s.pod,
	})
}

func burnCPU(deadline time.Time, seed uint64) {
	value := seed
	for time.Now().Before(deadline) {
		value = bits.RotateLeft64(value*6364136223846793005+1442695040888963407, 17)
	}
	runtime.KeepAlive(value)
}

func (s *Server) handleWound(w http.ResponseWriter, r *http.Request) {
	if !s.labEndpointsEnabled {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	megabytes, err := boundedInt(r.URL.Query().Get("mb"), 100, 1, 4096)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	chunk := make([]byte, megabytes*1024*1024)
	for offset := 0; offset < len(chunk); offset += 4096 {
		chunk[offset] = 1
	}
	s.heldMu.Lock()
	s.held = append(s.held, chunk)
	held := 0
	for _, allocation := range s.held {
		held += len(allocation)
	}
	s.heldMu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"wound_mb": megabytes,
		"held_mb":  held / (1024 * 1024),
		"pod":      s.pod,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "pod": s.pod})
}

func (s *Server) handleLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive", "pod": s.pod})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// Readiness answers "can this instance serve a real request now?". With the
	// PostgreSQL store it includes a bounded database ping; liveness deliberately
	// does not, so a temporary database outage does not restart a healthy process.
	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "order store not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "pod": s.pod})
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintf(w, "# HELP devops_may_cry_http_requests_total Requests received by the API.\n")
	fmt.Fprintf(w, "# TYPE devops_may_cry_http_requests_total counter\n")
	fmt.Fprintf(w, "devops_may_cry_http_requests_total %d\n", s.requests.Load())
	fmt.Fprintf(w, "# HELP devops_may_cry_orders_total Orders accepted by this process.\n")
	fmt.Fprintf(w, "# TYPE devops_may_cry_orders_total counter\n")
	fmt.Fprintf(w, "devops_may_cry_orders_total %d\n", s.orders.Load())
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		s.requests.Add(1)
		next.ServeHTTP(w, r)
		if r.URL.Path != "/livez" && r.URL.Path != "/readyz" {
			s.logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", time.Since(started).Milliseconds(),
				"pod", s.pod,
			)
		}
	})
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("panic recovered", "value", recovered, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func boundedInt(raw string, fallback, minimum, maximum int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum || value > maximum {
		return 0, fmt.Errorf("value must be an integer between %d and %d", minimum, maximum)
	}
	return value, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

const indexHTML = `<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>DevOps May Cry Dispatch</title>
  <style>
    :root { color-scheme: dark; --ink:#e8edf5; --muted:#8f9aab; --red:#f15060; --cyan:#47d7e8; }
    * { box-sizing:border-box } body { margin:0; min-height:100vh; display:grid; place-items:center;
      background:radial-gradient(circle at 20% 10%,#251735 0,transparent 36%),#080b10;
      color:var(--ink); font:16px/1.55 ui-monospace,SFMono-Regular,Menlo,monospace }
    main { width:min(760px,calc(100% - 36px)); border:1px solid #29313d; border-radius:18px;
      background:#10151dcc; box-shadow:0 24px 80px #000a; padding:30px }
    header { display:flex; justify-content:space-between; gap:20px; align-items:center; border-bottom:1px solid #29313d; padding-bottom:20px }
    h1 { margin:0; letter-spacing:.12em; font-size:clamp(24px,5vw,46px) } .badge { color:var(--cyan) }
    p { color:var(--muted) } .panel { margin-top:24px; min-height:150px; border-left:4px solid var(--red); padding:16px 20px; background:#0b0f15 }
    button { border:0; border-radius:9px; background:var(--red); color:white; font:inherit; padding:12px 16px; cursor:pointer }
    code { color:var(--cyan) } #pod { font-size:13px; color:var(--cyan) }
  </style>
</head>
<body><main>
  <header><div><div class="badge">24×7 INCIDENT RESPONSE</div><h1>DEVOPS MAY CRY</h1></div><button id="next">НОВЫЙ ВЫЗОВ</button></header>
  <p>Диспетчерская принимает инциденты, назначает дежурного и оставляет доказательства в логах и метриках.</p>
  <section class="panel"><strong id="who">соединяемся…</strong><p id="quote"></p><div id="pod"></div></section>
  <p><code>/quote · /order · /livez · /readyz · /metrics</code></p>
</main><script>
async function next(){const r=await fetch('/quote');const x=await r.json();who.textContent=x.who;quote.textContent=x.quote;pod.textContent='ответил pod: '+x.pod}
document.querySelector('#next').addEventListener('click',next);next();
</script></body></html>`
