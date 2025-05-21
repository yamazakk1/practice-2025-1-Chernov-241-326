# Этапы работы над проектом

## 1 марта – 15 марта 2025: Изучение технологий
- Изучил основы Go: синтаксис, пакеты `net/http`, `time`.
- Ознакомился с PostgreSQL и библиотекой `pgx/v5`.
- Прочитал материалы [codecrafters-io/build-your-own-x](https://github.com/codecrafters-io/build-your-own-x).
- Настроил окружение: установил Go, PostgreSQL, VS Code.

## 16 марта – 31 марта 2025: Разработка in-memory версии сервера
- Создал базовую структуру проекта: `src/main.go`, `src/paste.go`.
- Реализовал in-memory хранилище с использованием `map` для паст.
- Добавил функции создания и получения паст (`CreatePaste`, `GetPaste`).
- Протестировал базовую логику в консоли.

## 1 апреля – 15 апреля 2025: Интеграция с PostgreSQL
- Настроил PostgreSQL: создал базу `pastebin` и таблицу `pastes`.
- Обновил `paste.go` для работы с `pgx/v5`.
- Перенёс хранение паст из `map` в БД.
- Проверил создание и получение паст через PostgreSQL.

## 15 апреля – 30 апреля 2025: Создание сайта "Киберполигон"
- Создал статический сайт с 5 страницами: `index.html`, `about.html`, `participation.html`, `blog.html`, `resources.html`.
- Настроил стили в `style.css` с адаптивным дизайном.
- Добавил контент о платформе "Киберполигон" (Django, SQLite, Vagrant).
- Проверил отображение сайта в браузере.
- Скриншоты: `resources/screenshots/index_screenshot.png`, `resources/screenshots/blog_screenshot.png`.

## 16 апреля – 6 мая 2025: Разработка и тестирование HTTP-сервера
- Добавил `httpserver.go` с обработчиками `/`, `/create`, `/paste/<slug>`.
- Реализовал форму для создания паст и отображение паст в браузере.
- Добавил фоновую очистку паст (`StartBackgroundCleanUp`).
- Реализовал консольные команды `stop` и `count`.
- Провёл тестирование: создание паст, проверка сроков действия, удаление.
- Скриншоты: `resources/screenshots/server_home.png`, `resources/screenshots/paste_view.png`, `resources/screenshots/console.png`.

## 7 мая – 20 мая 2025: Документация и финализация
- Создал страницу `site/server.html` для описания pastebin-сервера.
- Написал документацию: `docs/server_tutorial.md`, `docs/project_stages.md`.
- Подготовил UML-диаграммы: `resources/diagrams/architecture.png`, `resources/diagrams/sequence_create.png`.
- Сделал скриншоты 
- Подготовил отчёт по практике (`reports/report.docx`, `reports/report.pdf`).