# README For Backend

Дата: 2026-03-21

## Контекст

Фронт уже покрывает основной сценарий кейса:

1. админ создаёт психолога
2. психолог создаёт и публикует тест
3. клиент проходит тест по публичной ссылке
4. психолог получает результат и формирует отчёт

Фронт ожидает, что backend стабильно поддерживает этот сценарий end-to-end.

## Что уже использует фронт

### Auth

- `POST /auth/admin/login`
- `POST /auth/admin/logout`
- `POST /auth/psychologists/login`
- `POST /auth/psychologists/logout`
- `GET /admins/me`
- `GET /psychologists/me`

### Admin

- `GET /admins/psychologists`
- `POST /admins/psychologists`
- `GET /admins/psychologists/{id}/workspace`
- `PUT /admins/psychologists/{id}`
- `PUT /admins/psychologists/{id}/access`
- `PUT /admins/psychologists/{id}/profile`
- `PUT /admins/psychologists/{id}/card`

### Psychologist

- `GET /psychologists/me/profile`
- `PUT /psychologists/me/profile`
- `GET /psychologists/me/card`
- `PUT /psychologists/me/card`

### Tests / Questions / Formulas

- `GET /psychologists/tests`
- `POST /psychologists/tests`
- `GET /psychologists/tests/{id}`
- `PUT /psychologists/tests/{id}`
- `DELETE /psychologists/tests/{id}`
- `POST /psychologists/tests/{id}/publish`
- `GET /psychologists/tests/{id}/questions`
- `POST /psychologists/tests/{id}/questions`
- `PUT /psychologists/tests/{id}/questions/{questionId}`
- `DELETE /psychologists/tests/{id}/questions/{questionId}`
- `GET /psychologists/tests/{id}/formulas`
- `POST /psychologists/tests/{id}/formulas`
- `PUT /psychologists/tests/{id}/formulas/{ruleId}`
- `DELETE /psychologists/tests/{id}/formulas/{ruleId}`
- `POST /psychologists/tests/{id}/formulas/calculate`

### Results / Reports

- `GET /psychologists/tests/{id}/results`
- `GET /psychologists/tests/{id}/results/{sessionId}`
- `GET /psychologists/results/{sessionId}/report`
- `GET /psychologists/report-templates`
- `POST /psychologists/report-templates`
- `PUT /psychologists/report-templates/{templateId}`
- `DELETE /psychologists/report-templates/{templateId}`

### Public flow

- `GET /public/tests/{slug}`
- `POST /public/tests/{slug}/start`
- `POST /public/tests/{slug}/progress`
- `POST /public/tests/{slug}/submit`

## Критичные ожидания от backend

### 1. Единый формат ошибок

Фронт ожидает:

```json
{
  "message": "Validation failed",
  "field_errors": {
    "email": "Email already exists"
  }
}
```

`field_errors` уже разбираются на фронте.

### 2. Публичная карточка автора теста

Фронт добавил страницу:

- `/session/:slug/author`

Лучше всего, если `GET /public/tests/{slug}` будет отдавать автора в явном виде:

```json
{
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
```

### 3. Метрики результата

Фронт поддерживает и legacy `career_result`, и универсальные метрики.

Желательный итоговый payload:

```json
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
  ]
}
```

### 4. Автосохранение публичной сессии

Фронт теперь автоматически вызывает:

- `POST /public/tests/{slug}/progress`

Ожидания:

- endpoint идемпотентен
- нормально переживает частые сохранения
- возвращает актуальные `answers`, `session_id`, `status`

### 5. Reorder вопросов

На:

- `PUT /psychologists/tests/{id}/questions/{questionId}`

ранее уже был `500` при обновлении `order_number`.

Нужно, чтобы reorder был стабилен.

### 6. Отчёты

Нужно, чтобы:

- `report_template_id` стабильно возвращался в тесте
- `GET /psychologists/results/{sessionId}/report` отдавал валидный `html/docx`
- `content-disposition` содержал `filename`

## Smoke-checklist

Проверьте совместно с фронтом:

1. админ логинится
2. админ создаёт психолога
3. психолог логинится
4. психолог создаёт тест
5. психолог добавляет вопросы
6. психолог публикует тест
7. `GET /public/tests/{slug}` отдаёт тест и автора
8. клиент стартует тест
9. клиент сохраняет прогресс
10. клиент завершает тест
11. психолог видит сессию в результатах
12. психолог открывает деталку сессии
13. генерируется `html`
14. генерируется `docx`

## Приоритет доработок

1. стабилизировать payload автора в `GET /public/tests/{slug}`
2. стабилизировать итоговый payload результата
3. убрать падения при reorder вопросов
4. стабилизировать `report templates` и `report generation`
5. держать единый контракт ошибок во всех endpoint’ах
