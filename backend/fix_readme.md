README for Backend

Дата: 2026-03-21

1. Текущее состояние фронта

Фронтенд уже реализован как рабочий MVP и включает три основных блока:

Admin
логин / logout
список психологов
создание психолога
блокировка / разблокировка
продление доступа
карточка психолога
реестр подписок
Psychologist
логин / logout
dashboard
профиль
список тестов
конструктор вопросов
formula rules
report templates
список результатов
деталка сессии
генерация отчётов (HTML / DOCX)
Public client flow
старт по публичной ссылке
сохранение прогресса
завершение теста
экран результата
карточка автора теста

Сборка фронта проходит успешно.

2. Базовые требования к API
Base URL
dev: /api (через Vite proxy)
local backend: http://localhost:8080
Авторизация
cookies + CSRF
для всех non-GET запросов требуется заголовок:
X-CSRF-Token
Формат ошибок (обязательный единый контракт)
{
  "message": "Validation failed",
  "field_errors": {
    "email": "Email already exists"
  }
}

Фронт уже умеет работать с field_errors.

3. Используемые фронтом endpoint’ы
Auth
POST /auth/admin/login
POST /auth/admin/logout
POST /auth/psychologists/login
POST /auth/psychologists/logout
GET /admins/me
GET /psychologists/me
Admin
GET /admins/psychologists
POST /admins/psychologists
GET /admins/psychologists/{id}/workspace
PUT /admins/psychologists/{id}
PUT /admins/psychologists/{id}/access
PUT /admins/psychologists/{id}/profile
PUT /admins/psychologists/{id}/card
PUT /admins/me
Psychologist workspace
GET /psychologists/me/profile
PUT /psychologists/me/profile
GET /psychologists/me/card
PUT /psychologists/me/card
Tests / Questions / Formulas
GET /psychologists/tests
POST /psychologists/tests
GET /psychologists/tests/{id}
PUT /psychologists/tests/{id}
DELETE /psychologists/tests/{id}
POST /psychologists/tests/{id}/publish
GET /psychologists/tests/{id}/questions
POST /psychologists/tests/{id}/questions
PUT /psychologists/tests/{id}/questions/{questionId}
DELETE /psychologists/tests/{id}/questions/{questionId}
GET /psychologists/tests/{id}/formulas
POST /psychologists/tests/{id}/formulas
PUT /psychologists/tests/{id}/formulas/{ruleId}
DELETE /psychologists/tests/{id}/formulas/{ruleId}
POST /psychologists/tests/{id}/formulas/calculate
Results / Reports
GET /psychologists/tests/{id}/results
GET /psychologists/tests/{id}/results/{sessionId}
GET /psychologists/results/{sessionId}/report?audience=client|psychologist&format=html|docx
GET /psychologists/report-templates
POST /psychologists/report-templates
PUT /psychologists/report-templates/{templateId}
DELETE /psychologists/report-templates/{templateId}
Public client flow
GET /public/tests/{slug}
POST /public/tests/{slug}/start
POST /public/tests/{slug}/progress
POST /public/tests/{slug}/submit
4. Критично важные требования для фронта
4.1 Публичная карточка автора теста

Страница:
/session/:slug/author

Данные берутся из:
GET /public/tests/{slug}

Рекомендуемый стабильный контракт:

{
  "id": 12,
  "slug": "base-test",
  "title": "Base test",
  "description": "....",
  "questions": [],
  "psychologist": {
    "user": {
      "id": 5,
      "full_name": "Иванов Иван Иванович",
      "email": "psy@example.com"
    },
    "profile": {
      "specialization": "Профориентация",
      "city": "Москва",
      "about": "....",
      "education": "....",
      "methods": "....",
      "experience_years": 6,
      "timezone": "Europe/Moscow",
      "is_public": true
    },
    "card": {
      "headline": "Психолог-профориентолог",
      "short_bio": "....",
      "contact_email": "psy@example.com",
      "contact_phone": "+7 999 123-45-67",
      "telegram": "@psy",
      "online_available": true,
      "offline_available": false
    }
  }
}

Важно:

допустимы альтернативные структуры, но контракт должен быть стабильным
текущая гибкость фронта — временная
4.2 Универсальные метрики результата

Желательный формат:

{
  "session_id": 77,
  "status": "completed",
  "answers": [],
  "metrics": {
    "stress_resistance": 12,
    "leadership": 7,
    "creativity": 4
  },
  "top_metrics": [
    { "key": "stress_resistance", "value": 12 },
    { "key": "leadership", "value": 7 }
  ],
  "career_result": {
    "scales": [],
    "top_scales": [],
    "top_professions": []
  }
}

Важно:

поддержка legacy career_result пока нужна
но основной контракт — metrics + top_metrics
4.3 Автосохранение сессии

Endpoint:
POST /public/tests/{slug}/progress

Требования:

идемпотентность
устойчивость к частым вызовам
возврат актуальных:
answers
session_id
status
4.4 Перестановка вопросов

Endpoint:
PUT /psychologists/tests/{id}/questions/{questionId}

Проблема: ранее возникал 500 при изменении order_number.

Требования:

стабильный reorder
отсутствие конфликтов
корректная обработка одинаковых временных значений
4.5 Report templates и генерация отчётов

Требования:

GET /psychologists/tests/{id} всегда возвращает report_template_id
GET /psychologists/results/{sessionId}/report:
отдаёт корректный файл
содержит Content-Disposition с filename
5. Где backend должен быть строгим
5.1 Создание психолога

Даже если фронт валидирует, backend обязан проверять:

full_name (3 слова, кириллица)
email
password
contact_phone (+7 999 123-45-67)
city (кириллица)
5.2 Данные респондента (public)

Endpoint:
POST /public/tests/{slug}/start

Ожидания:

respondent_name → минимум имя + фамилия
respondent_phone → нормализуемый формат
respondent_email → optional
respondent_age / gender / education → обязательны только если включены
6. Уже реализовано на фронте
cookies + CSRF flow
обработка field_errors
public author page
report templates
генерация HTML / DOCX
автосохранение
fallback под нестабильные payload’ы
7. Smoke checklist
admin логинится
admin создаёт психолога
психолог логинится
психолог создаёт тест
добавляет вопросы
публикует тест
GET /public/tests/{slug} возвращает тест и автора
клиент стартует тест
сохраняет прогресс
завершает тест
психолог видит результат
открывает сессию
генерируется HTML
генерируется DOCX
8. Приоритет задач для backend
стабилизировать payload автора (GET /public/tests/{slug})
стабилизировать результат (metrics + top_metrics)
исправить reorder вопросов
стабилизировать отчёты и шаблоны
унифицировать field_errors во всех endpoint’ах