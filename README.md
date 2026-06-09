# 🧠 DevBrain

<p align="center">
  <!-- Замени example.com на URL к изображению, если хочешь добавить логотип -->
  <!-- <img src="https://example.com/logo.png" alt="DevBrain Logo" width="200"/> -->
  <br>
  <strong>Ваш интеллектуальный менеджер ссылок</strong>
</p>

<p align="center">
  <!-- Примеры бейджей -->
  <a href="#"><img src="https://img.shields.io/badge/Language-Go-blue.svg" alt="Go"></a>
  <a href="#"><img src="https://img.shields.io/badge/Database-SQLite%20%7C%20PostgreSQL-green.svg" alt="Database"></a>
  <a href="#"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="#"><img src="https://img.shields.io/badge/Status-Stable-brightgreen.svg" alt="Status"></a>
</p>

<p align="center">
  DevBrain — это мощный и интуитивно понятный инструмент для организации ваших закладок. Сортируйте, ищите и анализируйте ваши любимые ресурсы в одном месте.
</p>

## ✨ Особенности

- **🏷️ Теги:** Назначайте теги ссылкам для удобной категоризации.
- **🔍 Поиск:** Быстро находите нужные ссылки по заголовкам, URL и тегам.
- **⚙️ Управление:** Добавляйте, редактируйте и удаляйте закладки.
- **📊 Статистика:** Отслеживайте количество ваших ссылок и популярные теги.
- **🔒 Безопасность:**
  - Аутентификация с использованием **JWT**.
  - Защита от распространённых атак (Rate Limiting, CORS, безопасные заголовки).
  - Надёжное **хеширование паролей** (Argon2id).
- **⚡ Технологии:** Go (backend), HTML/CSS/JavaScript (frontend), SQLite (или PostgreSQL).
- **☁️ Деплой:** Простой деплой на Render.com (с поддержкой Auto-Deploy).

## 🚀 Быстрый старт (локально)

1.  **Убедитесь, что у вас установлен [Go](https://go.dev/) (версия >= 1.21).**
2.  **Клонируйте репозиторий:**
    ```bash
    git clone https://github.com/skywaJlker192/DevBrain.git
    cd DevBrain
    ```
3.  **Создайте файл `.env`** в корне проекта, основываясь на `.env.example`. Укажите настройки базы данных (SQLite или PostgreSQL) и секретные ключи.
4.  **Запустите сервер:**
    ```bash
    go run ./cmd/server
    ```
    Откройте `http://localhost:8080` в вашем браузере.

## ☁️ Деплой на Render.com

1.  Создайте аккаунт на [Render.com](https://render.com).
2.  Создайте новый **Web Service**.
3.  Подключите репозиторий GitHub (`skywaJlker192/DevBrain`).
4.  Укажите:
    -   **Build Command:** `go build -o server ./cmd/server`
    -   **Start Command:** `./server`
5.  В разделе **Environment Variables** укажите переменные из вашего `.env` файла.
6.  Нажмите **Create Web Service**.
7.  Убедитесь, что включён **Auto-Deploy** для автоматического обновления при пуше в `main`.

## 📁 Структура проекта
