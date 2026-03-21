# Hackathon2026 Backend

Backend платформы ПрофДНК на Go, Gin и PostgreSQL.

## Что делает сервис

Backend закрывает основной сценарий кейса:

- администратор создает психологов и управляет доступом
- психолог работает в личном кабинете
- психолог создает и публикует тесты
- клиент проходит тест по публичной ссылке без регистрации
- ответы, метрики и результаты сохраняются в системе
- система формирует два вида отчетов: для клиента и для психолога
- отчеты собираются в реальном времени в `HTML` и `DOCX` и не хранятся на диске

При запуске приложение автоматически:

- подключается к PostgreSQL
- применяет все миграции из `migrations/*.up.sql`
- создает admin-аккаунты из `ADMIN_ACCOUNTS`
- при `DEMO_DATA_ENABLED=true` сидирует демо-психолога, демо-шаблон отчета и демо-тест
- поднимает HTTP-сервер на порту `8080`

## Стек

- Go
- Gin
- PostgreSQL
- Docker / Docker Compose
- Swagger / OpenAPI

## Быстрый старт без Docker

1. Скопируйте env-файл:

```bash
cp .env.example .env
```

2. Убедитесь, что PostgreSQL уже запущен локально.

3. Запустите backend:

```bash
go run ./cmd
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

## Локальный запуск через Docker Compose

Этот вариант собирает backend-образ локально и поднимает backend вместе с PostgreSQL.

1. Скопируйте env-файл:

```bash
cp .env.example .env
```

2. Поднимите сервисы:

```bash
docker compose up --build -d
```

3. Посмотрите логи:

```bash
docker compose logs -f app
```

4. Остановите сервисы:

```bash
docker compose down
```

Если нужно удалить и volume базы:

```bash
docker compose down -v
```

## Просмотр базы через Adminer

Для локальной разработки можно отдельно поднять PostgreSQL и Adminer:

```bash
docker compose up -d db adminer
```

Далее откройте `http://localhost:8081` и используйте:

- `System`: `PostgreSQL`
- `Server`: `db`
- `Username`: значение `DB_USER` из `.env`
- `Password`: значение `DB_PASSWORD` из `.env`
- `Database`: значение `DB_NAME` из `.env`

## Swagger / OpenAPI

После запуска доступны:

- `http://localhost:8080/swagger/`
- `http://localhost:8080/swagger/openapi.yaml`

## Основные env-переменные

Пример лежит в [.env.example](/home/gr1nder/Coding/Hackathon2026/backend/.env.example).

Ключевые переменные:

- `APP_ENV` - режим приложения
- `APP_IMAGE` - имя Docker-образа для deploy compose
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SSLMODE` - настройки PostgreSQL
- `COOKIE_SECURE` - включать `true` только если backend работает по HTTPS
- `ALLOWED_ORIGINS` - список frontend origin для CORS
- `PUBLIC_BASE_URL` - базовый публичный URL backend, используется при генерации public link и client report URL
- `ADMIN_ACCOUNTS` - сидинг admin-аккаунтов в формате `login:password:Full Name`
- `DEMO_DATA_ENABLED` - включает демо-данные
- `DEMO_PSYCHOLOGIST_EMAIL`, `DEMO_PSYCHOLOGIST_PASSWORD`, `DEMO_PSYCHOLOGIST_FULL_NAME` - данные демо-психолога
- `ADMINER_PORT` - порт Adminer в локальной разработке

## Что входит в демо-данные

При `DEMO_DATA_ENABLED=true` backend автоматически создает:

- демо-аккаунт психолога
- демо-шаблон отчета
- опубликованный демо-тест `ПрофДНК: IT-профориентация`
- пять demo scale-вопросов, подключенных к движку профориентации

Демо-логин:

- email: `demo.psychologist@profdnk.local`
- password: `demo12345`

Это закрывает требование кейса про заранее созданный тест и пример шаблонов отчетов.

## Архитектура

Проект построен по простой слоистой схеме:

- `internal/api` - HTTP-слой, маршруты Gin, cookie auth, CSRF, CORS, валидация и формирование HTTP-ответов
- `internal/service` - бизнес-логика: кабинет психолога, конструктор тестов, публичное прохождение, расчет результатов, генерация отчетов, demo seed
- `internal/repository` - SQL-доступ к PostgreSQL
- `internal/domain` - доменные модели и API-контракты
- `database` - подключение к БД и выполнение миграций
- `cmd` - точка входа приложения

## Ключевые технические решения

- Используется Go + Gin без ORM. SQL остается явным и предсказуемым.
- Аутентификация сделана через cookie sessions + CSRF, потому что проект ориентирован на browser-first frontend.
- Тесты поддерживают разные типы вопросов: `single_choice`, `multiple_choice`, `scale`, `text`, `number`.
- Результаты тестов рассчитываются через `scale_weights`, `option.score` и `formula_rules`.
- `career_result` сохраняется как snapshot в public session, чтобы исторические отчеты не менялись после последующего редактирования теста.
- Шаблоны отчетов хранятся в PostgreSQL как JSON-конфиги.
- Отчеты формируются по запросу: сначала рендерится HTML, затем из него собирается DOCX. Файлы отчетов на диск не сохраняются.
- У теста есть настройка `show_client_report_immediately`: если она включена, клиент после завершения теста может открыть свой client-report по публичной ссылке и `access_token`.

## Что уже реализовано по продукту

- кабинет администратора
- кабинет психолога
- профиль и визитка психолога
- CRUD тестов
- CRUD вопросов
- безопасный reorder вопросов
- публикация теста по публичной ссылке
- публичное прохождение теста без регистрации
- автосохранение прогресса
- хранение прохождений и ответов по каждому клиенту
- формулы расчета результата
- клиентский и профессиональный отчет
- шаблоны отчетов
- публичная выдача клиентского отчета, если это разрешено настройкой теста

## Публичный сценарий клиента

Базовый flow сейчас такой:

1. Психолог публикует тест и получает public link.
2. Клиент открывает `/public/tests/{slug}`.
3. Клиент стартует сессию через `/public/tests/{slug}/start`.
4. Клиент сохраняет прогресс через `/public/tests/{slug}/progress`.
5. Клиент завершает тест через `/public/tests/{slug}/submit`.
6. Если у теста включен `show_client_report_immediately`, backend возвращает:

- `client_report_available`
- `client_report_url`

После этого клиент может открыть:

`GET /public/tests/{slug}/report?access_token=...&format=html|docx`

## Отчеты

Система умеет формировать:

- отчет для клиента
- отчет для психолога

Поддерживаемые форматы:

- `html`
- `docx`

Особенности:

- отчеты генерируются в реальном времени
- отчеты не хранятся на диске
- шаблон один, а содержимое отчета собирается для каждой конкретной сессии отдельно
- шаблоны можно настраивать отдельно для `client` и `psychologist`

## Docker Hub и передача backend фронтендеру

### Сборка и пуш образа

1. Логин в Docker Hub:

```bash
docker login
```

2. Сборка образа:

```bash
docker build -t gr1nder05/hackathon2026-backend:latest .
```

3. Пуш образа:

```bash
docker push gr1nder05/hackathon2026-backend:latest
```

При желании можно использовать version tag:

```bash
docker build -t gr1nder05/hackathon2026-backend:v1 .
docker push gr1nder05/hackathon2026-backend:v1
```

### Что передать фронтендеру

Для запуска через уже собранный образ фронтендеру нужны:

- `docker-compose.deploy.yml`
- `.env.example`
- имя образа в Docker Hub, например `gr1nder05/hackathon2026-backend:latest`

### Как фронтендеру запустить backend

1. Положить `docker-compose.deploy.yml` в отдельную папку.

2. Создать рядом `.env`. Минимальный пример:

```env
APP_IMAGE=gr1nder05/hackathon2026-backend:latest
APP_ENV=development
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=postgres
COOKIE_SECURE=false
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
PUBLIC_BASE_URL=http://localhost:8080
ADMIN_ACCOUNTS=admin:change_me_please:Platform Administrator
```

3. Поднять сервисы:

```bash
docker compose -f docker-compose.deploy.yml up -d
```

4. Если образ обновился:

```bash
docker compose -f docker-compose.deploy.yml pull
docker compose -f docker-compose.deploy.yml up -d
```

5. Посмотреть логи:

```bash
docker compose -f docker-compose.deploy.yml logs -f app
```

6. Остановить сервисы:

```bash
docker compose -f docker-compose.deploy.yml down
```

7. Остановить и удалить данные БД:

```bash
docker compose -f docker-compose.deploy.yml down -v
```

## Структура проекта

- `cmd` - вход в приложение
- `internal/api` - HTTP handlers и middleware
- `internal/domain` - модели и контракты
- `internal/service` - бизнес-логика
- `internal/repository` - SQL и работа с PostgreSQL
- `database` - инициализация БД и миграции
- `migrations` - SQL-миграции
- `docs` - swagger/openapi

## Полезные замечания для разработки

- для frontend-flow backend использует cookie session, а не bearer token
- для защищенных `POST/PUT/DELETE` запросов нужен `X-CSRF-Token`
- public test flow завязан на `slug` и `access_token`
- новые миграции применяются автоматически на старте приложения

## Проверка качества

Базовые команды, которые стоит прогонять перед сборкой образа:

```bash
env GOCACHE=/tmp/gocache go test ./...
env GOCACHE=/tmp/gocache go vet ./...
```
