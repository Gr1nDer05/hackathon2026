# Hackathon2026

Платформа ПрофДНК для хакатона: backend и frontend в одном репозитории.

Проект закрывает основной сценарий кейса:

1. администратор создаёт психологов и управляет доступом
2. психолог создаёт и публикует тесты
3. клиент проходит тест по публичной ссылке без регистрации
4. психолог получает результаты, метрики и формирует отчёты

## Структура репозитория

```text
backend/   Go + Gin + PostgreSQL + Swagger
frontend/  React + Vite + SWR + dnd-kit
```

## Что реализовано

### Backend

- cookie auth + CSRF для администратора и психолога
- кабинет психолога и кабинет администратора
- CRUD тестов, вопросов и formula rules
- публикация тестов по публичной ссылке
- публичное прохождение теста без регистрации
- автосохранение прогресса
- расчёт результатов и метрик
- шаблоны отчётов
- генерация отчётов `HTML / DOCX`
- AI-генерация draft шаблонов отчётов для `pro`-подписки
- заглушка-заявка на покупку подписки

### Frontend

- админка: психологи, карточка психолога, подписки
- кабинет психолога: dashboard, профиль, тесты, конструктор, результаты
- drag-and-drop конструктор вопросов
- formula rules и preview расчёта
- шаблоны отчётов и AI-draft flow
- клиентский публичный сценарий
- экран результата клиента
- карточка автора теста
- адаптация под мобильные экраны

## Документация по частям

- backend: [backend/README.md](backend/README.md)
- frontend: [frontend/README.md](frontend/README.md)
- backend для фронтенда: [backend/README_FOR_FRONTEND.md](backend/README_FOR_FRONTEND.md)
- frontend для бэкенда: [frontend/README_FOR_BACKEND.md](frontend/README_FOR_BACKEND.md)

Корневой README нужен как обзор репозитория. Подробности по env, API и внутренней архитектуре вынесены в README соответствующей части проекта.

## Стек

### Backend

- Go
- Gin
- PostgreSQL
- Docker / Docker Compose
- Swagger / OpenAPI

### Frontend

- React 19
- Vite
- SWR
- react-router-dom
- lucide-react
- `@dnd-kit/*`
- motion

## Варианты запуска

### Вариант 1. Локальная разработка

Backend поднимается из `backend/`, frontend запускается локально из `frontend/`.

1. Подготовить backend:

```bash
cd backend
cp .env.example .env
docker compose up -d db adminer
go run ./cmd
```

2. Подготовить frontend:

```bash
cd frontend
npm install
npm run dev
```

После этого:

- frontend: `http://localhost:5173`
- backend: `http://localhost:8080`
- swagger: `http://localhost:8080/swagger/`
- adminer: `http://localhost:8081`

Если backend удобнее запускать полностью в Docker, смотри [backend/README.md](backend/README.md).

### Вариант 2. Контейнерный demo-стенд

Контейнерный frontend лежит в `frontend/` и проксирует `/api` в backend.

```bash
cd frontend
cp .env.example .env
docker compose -f docker-compose.deploy.yml up -d --build frontend app db adminer
```

После этого:

- frontend: `http://localhost:3000`
- backend: `http://localhost:8080`
- swagger: `http://localhost:8080/swagger/`
- adminer: `http://localhost:8081`

Важно:

- этот сценарий использует backend image через `APP_IMAGE`
- если нужен запуск backend из исходников, используй `backend/`

## Основной сценарий проверки

1. Админ логинится
2. Создаёт психолога или использует demo-аккаунт
3. Психолог логинится
4. Заполняет профиль
5. Создаёт тест
6. Добавляет вопросы и formula rules
7. Привязывает шаблон отчёта
8. Публикует тест
9. Клиент проходит тест по публичной ссылке
10. Психолог открывает результаты
11. Генерирует `HTML` и `DOCX` отчёт

## Демо-данные

Backend умеет поднимать демо-данные при `DEMO_DATA_ENABLED=true`:

- демо-психолога
- демо-психолога с истекшей подпиской
- демо-шаблон отчёта
- демо-тест профориентации

Конкретные demo-логины и env-переменные описаны в [backend/README.md](backend/README.md).

## Что важно знать

- backend и frontend используют cookie-based auth
- для non-`GET` запросов нужен `X-CSRF-Token`
- frontend уже работает по контракту ошибок вида:

```json
{
  "message": "Validation failed",
  "field_errors": {
    "email": "Email already exists"
  }
}
```

- публичный клиентский отчёт доступен только для завершённых сессий и только если у теста включён `show_client_report_immediately`
- AI-генерация draft шаблона отчёта доступна для `pro`-подписки

## Что смотреть дальше

- если работаешь над API: [backend/README.md](backend/README.md)
- если работаешь над интерфейсом: [frontend/README.md](frontend/README.md)
- если синхронизируешь frontend и backend по контракту: смотри оба интеграционных README
