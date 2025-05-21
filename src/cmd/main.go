package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/yamazakk1/go-pastebin/internal/app"
	"github.com/yamazakk1/go-pastebin/server"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Инициализация БД
	connStr := "postgres://postgres:123@localhost:5432/pastebin?sslmode=disable"

	if err := app.InitDB(connStr); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer app.CloseDB()

	// Запуск фоновой очистки
	app.StartBackgroundCleanUp()

	// Запуск сервера
	srv := server.NewServer()
	if err := srv.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
