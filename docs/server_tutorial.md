# Техническое руководство: Создание Pastebin-сервера

Этот туториал поможет вам шаг за шагом создать pastebin-сервер на языке Go. Мы начнём с простой версии, где пасты хранятся в памяти компьютера, затем подключим базу данных PostgreSQL для надёжного хранения, добавим веб-интерфейс для работы через браузер, настроим автоматическую очистку устаревших паст и добавим управление через консоль. Руководство написано максимально понятно, чтобы даже новички могли разобраться. Каждый шаг объясняет, что мы делаем, зачем это нужно и как избежать ошибок.

## Что такое pastebin-сервер?
Pastebin — это веб-приложение, где можно сохранить текст (например, код, заметки или логи), получить уникальную ссылку (например, `/paste/aBc123`) и поделиться ею. Наш сервер будет:
- Принимать текст через форму и сохранять его с выбранным сроком действия (1 час, 1 день, 1 неделя).
- Показывать текст по уникальному URL.
- Удалять пасты, когда их срок действия истекает.
- Поддерживать команды в консоли: `stop` (остановить сервер) и `count` (показать количество паст).

## Шаг 1: Настройка окружения
**Зачем**: Нам нужно подготовить инструменты для разработки — установить Go, создать проект и настроить зависимости.

1. **Установите Go**:
   - Скачайте Go версии 1.18 или выше с [go.dev](https://go.dev/dl/).
   - Убедитесь, что Go работает: выполните в терминале `go version`. Вы должны увидеть что-то вроде `go version go1.20.5`.
   - Ошибка? Проверьте, добавлен ли Go в переменную PATH (инструкции на сайте Go).

2. **Создайте проект**:
   - Создайте папку `pastebin` и внутри неё папку `src/`.
   - В корне проекта выполните:
     ```bash
     go mod init github.com/yourusername/pastebin
     ```
     Это создаст файл `go.mod`, который управляет зависимостями. Замените `yourusername` на ваш ник (или оставьте любое имя, например, `pastebin`).

3. **Добавьте зависимости**:
   - Установите библиотеку `godotenv` для работы с файлом конфигурации `.env`:
     ```bash
     go get github.com/joho/godotenv
     ```
   - Позже мы добавим ещё зависимости, но пока этого достаточно.
   - Ошибка? Убедитесь, что вы в корне проекта (`pastebin`), и проверьте интернет-соединение.

**Результат**: У вас есть настроенный проект с Go и базовой структурой.

## Шаг 2: Реализация in-memory хранилища
**Зачем**: Сначала создадим простую версию сервера, где пасты хранятся в памяти (в переменной `map`). Это поможет понять, как работает логика паст, прежде чем подключать сложные вещи вроде базы данных.

1. **Создайте файл для логики паст**:
   - В папке `src/` создайте файл `paste.go` в пакете `app`. Это будет место для функций, которые создают и получают пасты.
   - Добавьте код:
     ```go
     package app

     import (
         "errors"
         "math/rand"
         "sync"
         "time"
     )

     type Paste struct {
         Text      string
         Slug      string
         ExpiresAt time.Time
     }

     var alphabet = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
     var storage = make(map[string]Paste)
     var mu sync.RWMutex

     func CreateSlug(n int) (string, error) {
         if n < 3 {
             return "", errors.New("слишком короткий slug")
         }
         bytes := make([]byte, n)
         for i := 0; i < n; i++ {
             bytes[i] = alphabet[rand.Intn(len(alphabet))]
         }
         return string(bytes), nil
     }

     func CreatePaste(text string, expires time.Duration) (string, error) {
         if text == "" {
             return "", errors.New("попытка записи пустой строки")
         }
         slug, _ := CreateSlug(10)
         mu.Lock()
         storage[slug] = Paste{
             Text:      text,
             Slug:      slug,
             ExpiresAt: time.Now().Add(expires),
         }
         mu.Unlock()
         return slug, nil
     }

     func GetPaste(slug string) (string, error) {
         mu.RLock()
         paste, exists := storage[slug]
         mu.RUnlock()
         if !exists {
             return "", errors.New("паста не найдена")
         }
         if time.Now().After(paste.ExpiresAt) {
             mu.Lock()
             delete(storage, slug)
             mu.Unlock()
             return "", errors.New("паста удалена")
         }
         return paste.Text, nil
     }
     ```
   - **Что делает код**:
     - `Paste` — структура для пасты (текст, slug, срок действия).
     - `storage` — карта (`map`), где ключ — slug, значение — паста.
     - `CreateSlug` — генерирует случайный slug (например, "xYz123") длиной 10 символов.
     - `CreatePaste` — сохраняет текст в `map`, возвращает slug.
     - `GetPaste` — возвращает текст пасты или ошибку, если паста не найдена или устарела.
     - `mu` (mutex) — защищает `map` от одновременного доступа, чтобы избежать ошибок в многопоточной среде.

2. **Протестируйте in-memory хранилище**:
   - Создайте `src/main.go` для проверки:
     ```go
     package main

     import (
         "fmt"
         "time"
         "github.com/yamazakk1/go-pastebin/internal/app"
     )

     func main() {
         slug, err := app.CreatePaste("Привет, это тестовая паста!", 1*time.Hour)
         if err != nil {
             fmt.Println("Ошибка создания пасты:", err)
             return
         }
         fmt.Println("Создана паста с slug:", slug)
         text, err := app.GetPaste(slug)
         if err != nil {
             fmt.Println("Ошибка получения пасты:", err)
             return
         }
         fmt.Println("Текст пасты:", text)
     }
     ```
   - Запустите:
     ```bash
     go run src/main.go
     ```
   - **Ожидаемый результат**: Вы увидите slug (например, "aBc123") и текст пасты. Если вы подождёте час, паста "удалится" (GetPaste вернёт ошибку).
   - **Ошибки**:
     - Если `go run` не работает, проверьте, что вы в корне проекта и `go.mod` существует.
     - Если паста не создаётся, убедитесь, что текст не пустой.

**Результат**: У вас есть рабочая логика для создания и получения паст, которая хранит данные в памяти. Это временное решение, так как данные теряются при перезапуске программы.

## Шаг 3: Интеграция с PostgreSQL
**Зачем**: Хранить пасты в памяти ненадёжно — при перезапуске сервера все данные пропадают. PostgreSQL сохраняет пасты на диске, что делает сервер более практичным.

1. **Установите PostgreSQL**:
   - Скачайте и установите PostgreSQL с [postgresql.org](https://www.postgresql.org/download/).
   - Запустите сервер PostgreSQL (инструкции зависят от вашей ОС).
   - Создайте базу данных:
     ```bash
     createdb pastebin
     ```
   - **Ошибки**:
     - Если `createdb` не работает, убедитесь, что PostgreSQL запущен и вы вошли под правильным пользователем (по умолчанию `postgres`).

2. **Настройте подключение**:
   - В корне проекта создайте файл `.env`:
     ```plaintext
     DB_HOST=localhost
     DB_PORT=5432
     DB_USER=postgres
     DB_PASSWORD=123
     DB_NAME=pastebin
     ```
     Замените `DB_PASSWORD` на ваш пароль для PostgreSQL.
   - Установите библиотеку `pgx/v5` для работы с PostgreSQL:
     ```bash
     go get github.com/jackc/pgx/v5
     ```

3. **Создайте таблицу**:
   - Подключитесь к базе `pastebin`:
     ```bash
     psql -U postgres -d pastebin
     ```
   - Выполните SQL-запрос:
     ```sql
     CREATE TABLE pastes (
         slug TEXT PRIMARY KEY,
         text TEXT NOT NULL,
         expires_at TIMESTAMP NOT NULL
     );
     ```
   - **Зачем**: Таблица `pastes` хранит slug (уникальный ключ), текст пасты и время истечения срока действия.
   - **Ошибки**: Если таблица не создаётся, проверьте права доступа и правильность команды.

4. **Обновите paste.go для работы с PostgreSQL**:
   - Замените in-memory логику на работу с БД:
     ```go
     package app

     import (
         "context"
         "errors"
         "math/rand"
         "time"
         "github.com/jackc/pgx/v5/pgxpool"
     )

     var dbPool *pgxpool.Pool
     var alphabet = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

     func InitDB(connString string) error {
         pool, err := pgxpool.New(context.Background(), connString)
         if err != nil {
             return err
         }
         dbPool = pool
         return nil
     }

     func CloseDB() {
         if dbPool != nil {
             dbPool.Close()
         }
     }

     func CreateSlug(n int) (string, error) {
         if n < 3 {
             return "", errors.New("слишком короткий slug")
         }
         bytes := make([]byte, n)
         for i := 0; i < n; i++ {
             bytes[i] = alphabet[rand.Intn(len(alphabet))]
         }
         return string(bytes), nil
     }

     func CreatePaste(text string, expires time.Duration) (string, error) {
         if text == "" {
             return "", errors.New("попытка записи пустой строки")
         }
         slug, _ := CreateSlug(10)
         _, err := dbPool.Exec(context.Background(),
             "INSERT INTO pastes (slug, text, expires_at) VALUES ($1, $2, $3)",
             slug, text, time.Now().Add(expires))
         return slug, err
     }

     func GetPaste(slug string) (string, error) {
         var text string
         var expiresAt time.Time
         err := dbPool.QueryRow(context.Background(),
             "SELECT text, expires_at FROM pastes WHERE slug = $1",
             slug).Scan(&text, &expiresAt)
         if err != nil {
             return "", err
         }
         if time.Now().After(expiresAt) {
             go deletePaste(slug)
             return "", errors.New("паста удалена")
         }
         return text, nil
     }

     func deletePaste(slug string) {
         dbPool.Exec(context.Background(), "DELETE FROM pastes WHERE slug = $1", slug)
     }
     ```
   - **Что изменилось**:
     - `InitDB` подключается к PostgreSQL через `pgxpool`.
     - `CreatePaste` вставляет пасту в таблицу `pastes`.
     - `GetPaste` получает пасту из БД и проверяет срок действия.
     - `deletePaste` удаляет устаревшую пасту.
   - **Ошибки**:
     - Если подключение не работает, проверьте `.env` и убедитесь, что PostgreSQL запущен.
     - Если запросы не выполняются, проверьте, создана ли таблица `pastes`.

5. **Протестируйте работу с БД**:
   - Обновите `main.go`:
     ```go
     package main

     import (
         "log"
         "github.com/joho/godotenv"
         "github.com/yamazakk1/go-pastebin/internal/app"
     )

     func main() {
         if err := godotenv.Load(); err != nil {
             log.Println("No .env file found")
         }
         connStr := "postgres://postgres:123@localhost:5432/pastebin?sslmode=disable"
         if err := app.InitDB(connStr); err != nil {
             log.Fatal("Ошибка подключения к БД:", err)
         }
         defer app.CloseDB()
         slug, err := app.CreatePaste("Тестовая паста в БД", 1*time.Hour)
         if err != nil {
             log.Fatal("Ошибка создания пасты:", err)
         }
         log.Println("Создана паста с slug:", slug)
         text, err := app.GetPaste(slug)
         if err != nil {
             log.Fatal("Ошибка получения пасты:", err)
         }
         log.Println("Текст пасты:", text)
     }
     ```
   - Запустите:
     ```bash
     go run src/main.go
     ```
   - **Ожидаемый результат**: Паста сохраняется в БД, и вы видите её slug и текст.
   - Проверьте в PostgreSQL:
     ```sql
     SELECT * FROM pastes;
     ```
   - **Ошибки**: Если паста не сохраняется, проверьте строку подключения (`connStr`) и права доступа к БД.

**Результат**: Теперь пасты хранятся в PostgreSQL, и данные сохраняются даже после перезапуска сервера.

## Шаг 4: Добавление HTTP-сервера
**Зачем**: Чтобы пользователи могли создавать и просматривать пасты через браузер, нам нужен HTTP-сервер с веб-интерфейсом.

1. **Создайте файл для HTTP-обработчиков**:
   - В `src/` создайте `httpserver.go` в пакете `server`:
     ```go
     package server

     import (
         "fmt"
         "html"
         "net"
         "net/http"
         "strings"
         "github.com/yamazakk1/go-pastebin/internal/app"
     )

     type Server struct {
         HTTPListener net.Listener
         ControlChan  chan struct{}
     }

     func NewServer() *Server {
         return &Server{ControlChan: make(chan struct{})}
     }

     func (s *Server) Start(httpAddr string) error {
         var err error
         s.HTTPListener, err = net.Listen("tcp", httpAddr)
         if err != nil {
             return fmt.Errorf("ошибка запуска HTTP сервера: %v", err)
         }
         mux := http.NewServeMux()
         mux.HandleFunc("/", s.handleRoot)
         mux.HandleFunc("/paste/", s.handlePaste)
         mux.HandleFunc("/create", s.handleCreate)
         httpServer := &http.Server{Handler: mux}
         return httpServer.Serve(s.HTTPListener)
     }

     func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
         w.Header().Set("Content-Type", "text/html; charset=utf-8")
         html := `
         <h1>Go Pastebin</h1>
         <form method="post" action="/create">
             <textarea name="text" rows="10" cols="50"></textarea><br>
             <label>Срок действия: </label>
             <select name="expires">
                 <option value="1h">1 час</option>
                 <option value="24h">1 день</option>
                 <option value="168h">1 неделя</option>
             </select><br>
             <input type="submit" value="Создать">
         </form>
         `
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
         slug, err := app.CreatePaste(text, expires)
         if err != nil {
             http.Error(w, "Ошибка создания пасты", http.StatusInternalServerError)
             return
         }
         html := fmt.Sprintf(`
         <h1>Паста успешно создана!</h1>
         <p>Ссылка: <a href="/paste/%s">/paste/%s</a></p>
         <a href="/">Создать новую</a>
         `, slug, slug)
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
             http.Error(w, "Паста не найдена или удалена", http.StatusNotFound)
             return
         }
         escapedText := html.EscapeString(text)
         formattedText := strings.ReplaceAll(escapedText, "\n", "<br>")
         html := fmt.Sprintf(`
         <h1>Просмотр пасты</h1>
         <p>%s</p>
         <a href="/">Создать новую</a>
         `, formattedText)
         fmt.Fprint(w, html)
     }
     ```
   - **Что делает код**:
     - `Server` — структура для сервера с каналом управления.
     - `Start` — запускает HTTP-сервер на порту (например, `:8080`).
     - `handleRoot` — отображает форму для создания пасты.
     - `handleCreate` — обрабатывает POST-запрос, создаёт пасту и показывает ссылку.
     - `handlePaste` — показывает текст пасты по slug, экранируя HTML для безопасности.
   - **Зачем**: Эти обработчики создают простой веб-интерфейс, чтобы пользователи могли работать с пастами через браузер.

2. **Обновите main.go для запуска сервера**:
   - Замените тестовый код:
     ```go
     package main

     import (
         "log"
         "github.com/joho/godotenv"
         "github.com/yamazakk1/go-pastebin/internal/app"
         "github.com/yamazakk1/go-pastebin/server"
     )

     func main() {
         if err := godotenv.Load(); err != nil {
             log.Println("No .env file found")
         }
         connStr := "postgres://postgres:123@localhost:5432/pastebin?sslmode=disable"
         if err := app.InitDB(connStr); err != nil {
             log.Fatal("Ошибка подключения к БД:", err)
         }
         defer app.CloseDB()
         srv := server.NewServer()
         if err := srv.Start(":8080"); err != nil {
             log.Fatal(err)
         }
     }
     ```
   - **Что делает**:
     - Загружает конфигурацию из `.env`.
     - Подключается к PostgreSQL.
     - Запускает HTTP-сервер на порту 8080.

3. **Протестируйте веб-интерфейс**:
   - Запустите:
     ```bash
     go run src/main.go
     ```
   - Откройте в браузере: `http://localhost:8080`.
   - Заполните форму, выберите срок действия (например, 1 час), нажмите "Создать".
   - Перейдите по ссылке `/paste/<slug>` и проверьте, что текст отображается.
   - **Ошибки**:
     - Если страница не открывается, проверьте, что порт 8080 свободен (`netstat -an | grep 8080`).
     - Если форма не работает, убедитесь, что `text` и `expires` передаются корректно.

**Результат**: Теперь у вас есть полноценный веб-сервер, где можно создавать и просматривать пасты через браузер.

![Главная страница сервера](../resources/screenshots/server_home.png)  


![Просмотр пасты](../resources/screenshots/paste_view.png)  


## Шаг 5: Добавление фоновой очистки
**Зачем**: Чтобы сервер не заполнялся устаревшими пастами, мы настроим автоматическую очистку каждые 30 секунд.

1. **Добавьте функцию очистки в paste.go**:
   - В конец `paste.go` добавьте:
     ```go
     func StartBackgroundCleanUp() {
         go func() {
             ticker := time.NewTicker(time.Second * 30)
             defer ticker.Stop()
             for {
                 select {
                 case <-ticker.C:
                     ExpirienceCheck()
                 }
             }
         }()
     }

     func ExpirienceCheck() {
         now := time.Now()
         rows, _ := dbPool.Query(context.Background(), "SELECT slug, expires_at FROM pastes")
         defer rows.Close()
         var slugs []string
         for rows.Next() {
             var slug string
             var expiresAt time.Time
             rows.Scan(&slug, &expiresAt)
             if now.After(expiresAt) {
                 slugs = append(slugs, slug)
             }
         }
         for _, slug := range slugs {
             deletePaste(slug)
         }
     }
     ```
   - **Что делает**:
     - `StartBackgroundCleanUp` запускает фоновую задачу, проверяющую пасты каждые 30 секунд.
     - `ExpirienceCheck` находит устаревшие пасты и удаляет их.
   - **Зачем**: Это экономит место в БД и поддерживает порядок.

2. **Запустите очистку**:
   - В `main.go` перед запуском сервера добавьте:
     ```go
     app.StartBackgroundCleanUp()
     ```
   - **Полный код main.go**:
     ```go
     package main

     import (
         "log"
         "github.com/joho/godotenv"
         "github.com/yamazakk1/go-pastebin/internal/app"
         "github.com/yamazakk1/go-pastebin/server"
     )

     func main() {
         if err := godotenv.Load(); err != nil {
             log.Println("No .env file found")
         }
         connStr := "postgres://postgres:123@localhost:5432/pastebin?sslmode=disable"
         if err := app.InitDB(connStr); err != nil {
             log.Fatal("Ошибка подключения к БД:", err)
         }
         defer app.CloseDB()
         app.StartBackgroundCleanUp()
         srv := server.NewServer()
         if err := srv.Start(":8080"); err != nil {
             log.Fatal(err)
         }
     }
     ```

3. **Протестируйте очистку**:
   - Создайте пасту с коротким сроком действия (например, 1 минута).
   - Подождите 1–2 минуты и попробуйте открыть пасту по `/paste/<slug>`.
   - **Ожидаемый результат**: Паста удаляется, и вы видите сообщение "Паста не найдена или удалена".
   - Проверьте в БД:
     ```sql
     SELECT * FROM pastes;
     ```
   - **Ошибки**: Если пасты не удаляются, проверьте логи сервера (`log.Println` в `ExpirienceCheck`).

**Результат**: Сервер теперь автоматически очищает устаревшие пасты, что делает его более надёжным.

## Шаг 6: Консольное управление
**Зачем**: Добавим команды `stop` и `count`, чтобы управлять сервером из консоли, например, узнать, сколько паст активно, или безопасно остановить сервер.

1. **Добавьте поддержку команд в httpserver.go**:
   - В конец `httpserver.go` добавьте:
     ```go
     import (
         "bufio"
         "fmt"
         "os"
     )

     func (s *Server) cmdReceiver() {
         scanner := bufio.NewScanner(os.Stdin)
         fmt.Print("> ")
         for scanner.Scan() {
             cmd := scanner.Text()
             switch cmd {
             case "stop":
                 close(s.ControlChan)
                 fmt.Println("Останавливаю сервер...")
                 os.Exit(0)
             case "count":
                 fmt.Printf("Всего паст: %d\n", app.GetPasteCount())
             default:
                 fmt.Println("Доступные команды: stop, count")
             }
             fmt.Print("> ")
         }
     }
     ```
   - В `paste.go` добавьте:
     ```go
     func GetPasteCount() int {
         var count int
         dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM pastes").Scan(&count)
         return count
     }
     ```

2. **Запустите обработку команд**:
   - В `httpserver.go` в метод `Start` добавьте:
     ```go
     go s.cmdReceiver()
     ```
   - **Полный код Start**:
     ```go
     func (s *Server) Start(httpAddr string) error {
         var err error
         s.HTTPListener, err = net.Listen("tcp", httpAddr)
         if err != nil {
             return fmt.Errorf("ошибка запуска HTTP сервера: %v", err)
         }
         mux := http.NewServeMux()
         mux.HandleFunc("/", s.handleRoot)
         mux.HandleFunc("/paste/", s.handlePaste)
         mux.HandleFunc("/create", s.handleCreate)
         httpServer := &http.Server{Handler: mux}
         go s.cmdReceiver()
         return httpServer.Serve(s.HTTPListener)
     }
     ```

3. **Протестируйте команды**:
   - Запустите сервер:
     ```bash
     go run src/main.go
     ```
   - В консоли введите:
     - `count` — должно показать количество паст в БД.
     - `stop` — сервер должен завершиться.
   - **Ошибки**:
     - Если команды не работают, проверьте, что `cmdReceiver` запускается (`go s.cmdReceiver()`).
     - Если `count` возвращает 0, убедитесь, что в БД есть пасты.

**Результат**: Теперь вы можете управлять сервером через консоль, что упрощает отладку и администрирование.

![Консоль с командами](../resources/screenshots/console.png)  


## Диаграммы
- **Архитектура сервера**:
  ![Архитектура](../resources/diagrams/architecture.png)  
 
- **Последовательность создания пасты**:
  ![Последовательность](../resources/diagrams/sequence_create.png)  

## Полезные советы
- **Безопасность**: Не публикуйте `.env` в Git (добавьте его в `.gitignore`).
- **Масштабирование**: Для продакшена используйте `golang-migrate` для управления схемой БД.
- **Тестирование**: Создавайте пасты с разными сроками действия и проверяйте их удаление.
- **Логирование**: Добавьте `log.Println` в критические места (например, `CreatePaste`), чтобы отслеживать ошибки.

## Проблемы и их решение
- **Сервер не запускается**: Проверьте, свободен ли порт 8080 (`netstat -an | grep 8080`).
- **БД не подключается**: Убедитесь, что PostgreSQL запущен, и проверьте `.env`.
- **Пасты не сохраняются**: Проверьте SQL-запросы и таблицу `pastes`.
- **Форма не работает**: Убедитесь, что имена полей в HTML (`text`, `expires`) совпадают с кодом.
