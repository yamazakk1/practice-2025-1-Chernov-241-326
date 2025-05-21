package server

import (
	"bufio"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yamazakk1/go-pastebin/internal/app"
)

// Структура сервера
type Server struct {
	HTTPListener net.Listener
	ControlChan  chan struct{}
}

// Конструктор сервера
func NewServer() *Server {
	return &Server{
		ControlChan: make(chan struct{}),
	}
}

// Начало работы HTTP сервера
func (s *Server) Start(httpAddr string) error {
	var err error

	// Запуск HTTP сервера
	s.HTTPListener, err = net.Listen("tcp", httpAddr)
	if err != nil {
		return errors.New("ошибка запуска HTTP сервера")
	}
	defer s.HTTPListener.Close()

	fmt.Println("HTTP сервер запустился на порте ", httpAddr)

	// Инициализация HTTP обработчиков
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/paste/", s.handlePaste)
	mux.HandleFunc("/create", s.handleCreate)

	httpServer := &http.Server{
		Handler: mux,
	}

	// Запуск фоновой очистки паст
	app.Start()

	// Фоновая работа консоли
	go s.cmdReceiver()

	// Запуск HTTP сервера
	err = httpServer.Serve(s.HTTPListener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// HTTP обработчики

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Go Pastebin</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
		textarea { width: 100%%; height: 200px; margin: 10px 0; }
		button { padding: 10px 15px; background: #4CAF50; color: white; border: none; cursor: pointer; }
		.paste { background: #f5f5f5; padding: 15px; margin: 10px 0; border-left: 3px solid #4CAF50; white-space: pre-wrap; }
	</style>
</head>
<body>
	<h1>Go Pastebin</h1>
	<form action="/create" method="post" accept-charset="utf-8">
		<textarea name="text" placeholder="Введите текст здесь..." required></textarea><br>
		<label>Срок действия: 
			<select name="expires">
				<option value="1h">1 час</option>
				<option value="24h" selected>1 день</option>
				<option value="168h">1 неделя</option>
			</select>
		</label>
		<button type="submit">Создать</button>
	</form>
</body>
</html>`

	fmt.Fprint(w, html)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	text := r.PostFormValue("text")
	if text == "" {
		http.Error(w, "Текст не может быть пустым", http.StatusBadRequest)
		return
	}

	expires, _ := time.ParseDuration(r.PostFormValue("expires"))
	if expires <= 0 {
		expires = 24 * time.Hour
	}

	slug, err := app.CreatePaste(text, expires)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Паста создана</title>
</head>
<body>
	<h1>Паста успешно создана!</h1>
	<div class="paste">
		<p>Ссылка: <a href="/paste/%s">/paste/%s</a></p>
		<p><a href="/">Создать новую</a></p>
	</div>
</body>
</html>`, slug, slug)

	fmt.Fprint(w, html)
}

func (s *Server) handlePaste(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	slug := strings.TrimPrefix(r.URL.Path, "/paste/")
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text, err := app.GetPaste(slug)
	if err != nil {
		// ... обработка ошибки
		return
	}

	// Экранируем HTML-символы и заменяем переносы строк на <br>
	escapedText := html.EscapeString(text)
	formattedText := strings.ReplaceAll(escapedText, "\n", "<br>")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Паста %s</title>
	<style>
		.paste-content {
			white-space: pre-wrap; /* Сохраняем переносы строк */
			font-family: monospace;
			background: #f5f5f5;
			padding: 15px;
			border-radius: 5px;
			overflow-wrap: break-word;
		}
	</style>
</head>
<body>
	<h1>Просмотр пасты</h1>
	<div class="paste-content">%s</div>
	<p><a href="/">Создать новую</a></p>
</body>
</html>`, slug, formattedText)

	fmt.Fprint(w, html)
}

// Работа консоли на фоне
func (s *Server) cmdReceiver() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		cmd := scanner.Text()
		switch cmd {
		case "stop":
			close(s.ControlChan)
			fmt.Println("Shutting down server...")
			os.Exit(0)
		case "count":
			fmt.Printf("Total pastes: %d\n", app.GetPasteCount())
		default:
			fmt.Println("Available commands: stop, count")
		}
		fmt.Print("> ")
	}
}
