# Форум Проект

Это проект форума, состоящий из следующих компонентов:
- Сервис аутентификации (gRPC)
- Сервис форума (REST API)
- База данных PostgreSQL
- Фронтенд (React)

## Требования

- Go 1.21 или выше
- PostgreSQL 14 или выше
- Node.js 16 или выше
- npm или yarn

## Запуск проекта

### 1. Настройка базы данных

1. Установите PostgreSQL, если еще не установлен
2. Создайте базу данных:
```sql
CREATE DATABASE forum_service;
```

3. Настройте переменные окружения для подключения к базе данных:
```bash
# Windows (PowerShell)
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="your_password"
$env:DB_NAME="forum_service"

# Linux/MacOS
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=forum_service
```

### 2. Запуск сервиса аутентификации

1. Перейдите в директорию сервиса аутентификации:
```bash
cd backend/auth-service
```

2. Установите зависимости:
```bash
go mod download
```

3. Запустите сервис:
```bash
go run cmd/main.go
```

Сервис будет доступен по адресу: `localhost:50051`

### 3. Запуск сервиса форума

1. Перейдите в директорию сервиса форума:
```bash
cd backend/forum-service
```

2. Установите зависимости:
```bash
go mod download
```

3. Запустите сервис:
```bash
go run cmd/main.go
```

Сервис будет доступен по адресу: `localhost:8080`

### 4. Запуск фронтенда

1. Перейдите в директорию фронтенда:
```bash
cd forum-frontend
```

2. Установите зависимости:
```bash
npm install
```

3. Запустите приложение:
```bash
npm start
```

Фронтенд будет доступен по адресу: `http://localhost:3000`

## API Документация

После запуска сервиса форума, вы можете получить доступ к Swagger документации по адресу:
```
http://localhost:8080/swagger/index.html
```

## Структура проекта

### Сервис аутентификации
- gRPC сервис на порту 50051
- Обрабатывает аутентификацию и авторизацию пользователей
- Управляет профилями пользователей

### Сервис форума
- REST API сервис на порту 8080
- Обрабатывает посты и комментарии
- Интегрируется с сервисом аутентификации

### База данных
- PostgreSQL на порту 5432
- Хранит данные пользователей, постов и комментариев 