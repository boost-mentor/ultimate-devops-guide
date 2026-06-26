package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" // драйвер Postgres (pure-Go, работает с CGO_ENABLED=0)
)

// Quote — одна цитата: персонаж + текст.
// Это намеренно простая структура: всё приложение помещается в один файл,
// без слоёв и архитектуры — нам важна демонстрация кубернетизации, а не код.
type Quote struct {
	Character string `json:"character"`
	Quote     string `json:"quote"`
	Pod       string `json:"pod"` // имя пода, который ответил — наглядно видно в k8s
}

// quotes — ВСТРОЕННЫЕ цитаты (fallback). Используются, только если БД не задана
// (нет переменной DB_HOST). Так приложение работает и без базы — удобно для
// первых демо (под, пробы, kubectl apply), а БД подключаем отдельным модулем.
var quotes = []Quote{
	{Character: "скотт", Quote: "Вообще, я против наркотиков, но если ты наркоманка, то я за."},
	{Character: "скотт", Quote: "Я влезбиянился в тебя!"},
	{Character: "скотт", Quote: "Если я надул в штаны, подыграешь, как будто это дождь?"},
	{Character: "уоллес", Quote: "Нельзя расстаться так, чтобы никто не страдал."},
	{Character: "скотт", Quote: "Я с девушками не дерусь, они... мягкие."},
	{Character: "", Quote: "— Скажи честно, мы отстой?\n— Я не знаю, вам видней."},
	{Character: "", Quote: "— Ты в рок-группе?\n— Да, мы ужасные. Приходи, а?"},
}

// db — пул соединений с Postgres. Если БД не сконфигурирована (нет DB_HOST),
// остаётся nil, и приложение отдаёт встроенные цитаты.
var db *sql.DB

// initDB подключается к Postgres, ЕСЛИ задан DB_HOST.
// Пароль приходит из переменной DB_PASSWORD — а её мы кладём в Kubernetes Secret
// и пробрасываем в под через secretKeyRef. Это и есть демонстрация секретов.
func initDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		log.Printf("DB_HOST не задан — работаю на встроенных цитатах (без БД)")
		return
	}
	port := envOr("DB_PORT", "5432")
	user := envOr("DB_USER", "scott")
	name := envOr("DB_NAME", "scott")
	pass := os.Getenv("DB_PASSWORD") // ← из Secret
	ssl := envOr("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl)

	var err error
	db, err = sql.Open("postgres", dsn) // ленивое: реальное соединение — при первом запросе/пинге
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	db.SetMaxOpenConns(5)
	log.Printf("БД сконфигурирована: %s:%s/%s (пароль взят из переменной DB_PASSWORD = Secret)", host, port, name)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	initDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/quote", quoteHandler)    // случайная фраза (из БД или встроенная)
	mux.HandleFunc("/healthz", healthHandler) // базовая health-ручка
	mux.HandleFunc("/livez", liveHandler)     // liveness probe
	mux.HandleFunc("/readyz", readyHandler)   // readiness probe (пингует БД, если она есть)

	addr := ":" + port
	log.Printf("Scott Pilgrim Quotes слушает %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// indexHandler отдаёт одностраничник. Любой неизвестный путь -> 404.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

// quoteHandler — основной API: возвращает случайную цитату в JSON (GET /quote).
// Если БД подключена — берём цитату из неё; иначе из встроенного списка.
func quoteHandler(w http.ResponseWriter, r *http.Request) {
	var q Quote
	var err error
	if db != nil {
		q, err = quoteFromDB()
		if err != nil {
			// БД недоступна/пустая — честно показываем 503, чтобы было видно зависимость.
			http.Error(w, "БД недоступна: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
	} else {
		q = quotes[rand.Intn(len(quotes))]
	}

	// Имя пода = hostname контейнера. В k8s это имя пода —
	// удобно показать, что запрос реально обслуживает конкретный под.
	q.Pod, _ = os.Hostname()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(q)
}

// quoteFromDB достаёт одну случайную цитату из таблицы quotes.
func quoteFromDB() (Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var q Quote
	row := db.QueryRowContext(ctx, `SELECT speaker, body FROM quotes ORDER BY random() LIMIT 1`)
	if err := row.Scan(&q.Character, &q.Quote); err != nil {
		return Quote{}, err
	}
	return q, nil
}

// healthHandler — базовая health-ручка (GET /healthz).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// liveHandler — liveness probe (GET /livez). Жив ли процесс? Не ответил → kubelet ПЕРЕЗАПУСКАЕТ под.
func liveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("alive"))
}

// readyHandler — readiness probe (GET /readyz). Готов ли под ПРИНИМАТЬ трафик?
// Если БД задана — пингуем её: нет БД → под NotReady и выводится из балансировки Service
// (но НЕ перезапускается). Именно этим readiness отличается от liveness.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "БД не готова: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

// indexHTML — простой одностраничник. Лежит прямо в бинаре,
// чтобы образ был самодостаточным: ни сторонних файлов, ни тома, ни CDN на бэке.
const indexHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Scott Pilgrim Quotes</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
<style>
  :root { --pink:#ff2e88; --cyan:#00e5ff; --yellow:#ffd400; --purple:#1a0033; }
  * { box-sizing: border-box; }
  body {
    margin:0; min-height:100vh; display:flex; flex-direction:column;
    align-items:center; justify-content:center; gap:28px;
    font-family:'Press Start 2P', monospace;
    background:
      radial-gradient(circle at 20% 20%, rgba(255,46,136,.25), transparent 40%),
      radial-gradient(circle at 80% 70%, rgba(0,229,255,.25), transparent 40%),
      var(--purple);
    color:#fff; padding:24px; text-align:center;
  }
  h1 {
    font-size:clamp(16px,4vw,30px); line-height:1.4; margin:0;
    color:var(--yellow);
    text-shadow:3px 3px 0 var(--pink), 6px 6px 0 var(--cyan);
  }
  .card {
    background:rgba(0,0,0,.45); border:4px solid var(--cyan);
    border-radius:14px; padding:32px 28px; max-width:680px; width:100%;
    box-shadow:8px 8px 0 var(--pink);
  }
  .character {
    color:var(--pink); font-size:clamp(13px,3vw,20px);
    margin-bottom:18px; text-transform:uppercase;
  }
  .quote {
    font-size:clamp(11px,2.4vw,15px); line-height:1.9; color:#fff;
    white-space:pre-line; /* переносы строк \n => реальные переносы (для диалогов) */
  }
  button {
    font-family:inherit; font-size:clamp(10px,2.2vw,14px);
    color:var(--purple); background:var(--yellow); border:none;
    padding:18px 26px; border-radius:10px; cursor:pointer;
    box-shadow:5px 5px 0 var(--pink); transition:transform .05s ease;
  }
  button:hover { transform:translate(2px,2px); box-shadow:3px 3px 0 var(--pink); }
  button:active { transform:translate(5px,5px); box-shadow:0 0 0 var(--pink); }
  .pod {
    font-size:10px; color:var(--cyan); opacity:.8; line-height:1.6;
  }
</style>
</head>
<body>
  <h1>★ SCOTT PILGRIM ★<br>QUOTES vs THE WORLD</h1>
  <div class="card">
    <div class="character" id="character">нажми кнопку ↓</div>
    <div class="quote" id="quote">Случайная цитата появится здесь.</div>
  </div>
  <button onclick="nextQuote()">▶ СЛЕДУЮЩАЯ ЦИТАТА</button>
  <div class="pod" id="pod"></div>

<script>
async function nextQuote() {
  try {
    const res = await fetch('/quote');
    if (!res.ok) { document.getElementById('quote').textContent = 'Бэкенд: ' + res.status + ' (БД недоступна?)'; return; }
    const data = await res.json();
    const charEl = document.getElementById('character');
    const quoteEl = document.getElementById('quote');
    if (data.character) {
      charEl.textContent = data.character + ':';
      charEl.style.display = 'block';
      quoteEl.textContent = '«' + data.quote + '»';
    } else {
      charEl.style.display = 'none';
      quoteEl.textContent = data.quote;
    }
    document.getElementById('pod').textContent = 'ответил под: ' + data.pod;
  } catch (e) {
    document.getElementById('quote').textContent = 'Бэкенд недоступен :(';
  }
}
nextQuote();
</script>
</body>
</html>`
