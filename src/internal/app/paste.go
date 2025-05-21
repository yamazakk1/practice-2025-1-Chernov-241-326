package app

import (
	"errors"
	"math/rand"
	"sync"
	"time"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)



var dbPool *pgxpool.Pool

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

func CreatePaste(text string, expires time.Duration) (string, error) {
	if text == "" {
		return "", errors.New("попытка записи пустой строки")
	}

	slug,_ := CreateSlug(10)
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
	dbPool.Exec(context.Background(),
		"DELETE FROM pastes WHERE slug = $1", slug)
}

// Остальные функции (CreateSlug, StartBackgroundCleanUp и т.д.) остаются без изменений
type Paste struct {
	Text      string
	Slug      string
	ExpiresAt time.Time
}

var alphabet = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
var storage = make(map[string]Paste)
var mu sync.RWMutex
var isrunning bool




// CreateSlug - создание слага
func CreateSlug(n int) (string, error) {
	if n < 3 {
		return "", errors.New("слишком короткий slug")
	}
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = alphabet[(rand.Intn(len(alphabet)))]
	}
	return string(bytes), nil
}





// ExpirienceCheck - проверка паст на проссроченность
func ExpirienceCheck() {
	now := time.Now()
	mu.RLock()
	expired := make([]string, 0)
	for key, value := range storage {
		if now.After(value.ExpiresAt) {
			expired = append(expired, key)
		}
	}
	mu.RUnlock()

	for _, key := range expired {
		deletePaste(key)
	}
}

// Запуск фоновой очистки паст с интервалом в 30 сек
func StartBackgroundCleanUp() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ExpirienceCheck()
			}
		}
	}()
}

// Начало работы программы
func Start() {
	if isrunning {
		return
	}
	isrunning = true

	go StartBackgroundCleanUp()
}

// Конец работы программы к
func Stop() {
	if !isrunning {
		return
	}
	isrunning = false
}

// ВЫвод количества паст
func GetPasteCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(storage)
}
