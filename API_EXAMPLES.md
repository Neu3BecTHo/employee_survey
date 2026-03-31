# API Examples - Employee Survey Web App

## 📋 Базовая информация

**Base URL:** `http://localhost:8080`

**Headers для всех запросов:**
- `X-User-Id: <user_id>` - ID пользователя (1=admin, 2=employee, 3=employee)
- `Content-Type: application/json` - для POST/PUT запросов

## � Быстрый старт

### 1. Запуск приложения (Docker)
```bash
cp .env.example .env
docker-compose up -d
```

### 2. Проверка работы API
```bash
curl http://localhost:8080/users
```

### 3. Минификация файлов (для продакшена)
```bash
# JS файлы
for f in static/js/*.js; do
  if [[ ! $f =~ \.min\.js$ ]]; then
    npx terser "$f" -o "${f%.js}.min.js" -c -m
  fi
done

# CSS файлы
for f in static/css/*.css; do
  if [[ ! $f =~ \.min\.css$ ]]; then
    npx clean-css-cli -o "${f%.css}.min.css" "$f"
  fi
done
```

## �👥 Пользователи по умолчанию:

| ID | Имя | Роль |
|----|-----|------|
| 1 | Иван Иванов | admin |
| 2 | Петр Петров | employee |
| 3 | Анна Сидорова | employee |

---

## 🔐 Публичные эндпоинты

### 1. Получение списка пользователей
```bash
curl -H "X-User-Id: 1" http://localhost:8080/users
```

### 2. Получение списка опросов
```bash
curl -H "X-User-Id: 1" http://localhost:8080/surveys
```

### 3. Получение деталей опроса
```bash
curl -H "X-User-Id: 1" http://localhost:8080/surveys/1
```

### 4. Получение результатов опроса
```bash
curl -H "X-User-Id: 1" http://localhost:8080/surveys/1/results
```

### 5. Получение своих ответов (employee)
```bash
curl -H "X-User-Id: 2" http://localhost:8080/surveys/my
```

---

## 🛡️ Администраторские эндпоинты

### 1. Создание опроса (только admin)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 1" \
  -d '{
    "title": "Новый опрос о работе",
    "description": "Как вам работаеться в нашей компании?",
    "status": "draft"
  }' \
  http://localhost:8080/surveys
```

### 2. Обновление опроса (только admin)
```bash
curl -X PUT \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 1" \
  -d '{
    "title": "Обновленный опрос",
    "description": "Обновленное описание",
    "status": "open"
  }' \
  http://localhost:8080/surveys/1
```

### 3. Добавление вопроса (только admin)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 1" \
  -d '{
    "text": "Насколько вы довольны рабочей обстановкой?",
    "type": "single_choice",
    "is_required": true,
    "options": ["Очень доволен", "Доволен", "Нейтрально", "Недоволен", "Очень недоволен"]
  }' \
  http://localhost:8080/surveys/1/questions
```

### 4. Добавление текстового вопроса (только admin)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 1" \
  -d '{
    "text": "Что можно улучшить в работе?",
    "type": "text",
    "is_required": false,
    "options": []
  }' \
  http://localhost:8080/surveys/1/questions
```

### 5. Открытие опроса (только admin)
```bash
curl -X POST \
  -H "X-User-Id: 1" \
  http://localhost:8080/surveys/1/open
```

### 6. Закрытие опроса (только admin)
```bash
curl -X POST \
  -H "X-User-Id: 1" \
  http://localhost:8080/surveys/1/close
```

---

## 📝 Отправка ответов

### 1. Отправка ответов на опрос
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 2" \
  -d '{
    "answers": [
      {
        "question_id": 1,
        "value": "Отлично"
      },
      {
        "question_id": 2,
        "value": "Все хорошо, спасибо!"
      }
    ]
  }' \
  http://localhost:8080/surveys/1/responses
```

---

## 🚫 Тестирование ошибок

### 1. Попытка employee создать опрос (должно быть Forbidden)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 2" \
  -d '{"title":"Test","description":"Test"}' \
  http://localhost:8080/surveys
# Ожидаемый результат: Forbidden
```

### 2. Отправка пустых ответов (должна быть ошибка валидации)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 2" \
  -d '{"answers":[]}' \
  http://localhost:8080/surveys/1/responses
# Ожидаемый результат: required question X is not answered
```

### 3. Отправка ответов на закрытый опрос
```bash
# Сначала закрываем опрос
curl -X POST -H "X-User-Id: 1" http://localhost:8080/surveys/1/close

# Пытаемся отправить ответ
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 3" \
  -d '{"answers":[{"question_id":1,"value":"Test"}]}' \
  http://localhost:8080/surveys/1/responses
# Ожидаемый результат: Survey is not open for responses
```

### 4. Дублирование ответов
```bash
# Отправляем ответ первый раз
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 2" \
  -d '{"answers":[{"question_id":1,"value":"Test"}]}' \
  http://localhost:8080/surveys/1/responses

# Пытаемся отправить второй раз
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 2" \
  -d '{"answers":[{"question_id":1,"value":"Test"}]}' \
  http://localhost:8080/surveys/1/responses
# Ожидаемый результат: User has already responded to this survey
```

---

## 📊 Примеры ответов API

### Успешное создание опроса:
```json
{
  "id": 5,
  "title": "Новый опрос о работе",
  "description": "Как вам работаеться в нашей компании?",
  "status": "draft",
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Список опросов:
```json
[
  {
    "id": 1,
    "title": "Обратная связь по проекту",
    "description": "Как вы оцениваете коммуникацию в команде?",
    "status": "open",
    "created_at": "2024-01-01T10:00:00Z",
    "questions": [
      {
        "id": 1,
        "survey_id": 1,
        "text": "Как вы оцениваете коммуникацию в команде?",
        "type": "single_choice",
        "is_required": true,
        "options": ["Отлично", "Хорошо", "Нормально", "Плохо"],
        "created_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
]
```

### Результаты опроса:
```json
{
  "survey": {
    "id": 1,
    "title": "Обратная связь по проекту",
    "description": "Как вы оцениваете коммуникацию в команде?",
    "status": "open",
    "created_at": "2024-01-01T10:00:00Z"
  },
  "total_responses": 5,
  "question_results": [
    {
      "question": {
        "id": 1,
        "text": "Как вы оцениваете коммуникацию в команде?",
        "type": "single_choice",
        "is_required": true,
        "options": ["Отлично", "Хорошо", "Нормально", "Плохо"]
      },
      "answers": [
        {"value": "Отлично", "count": 2},
        {"value": "Хорошо", "count": 2},
        {"value": "Нормально", "count": 1}
      ]
    }
  ]
}
```

---

## 🔧 Отладка

### Проверка работы сервера:
```bash
curl http://localhost:8080/users
```

### Проверка ролевой системы:
```bash
# Admin запрос (успешный)
curl -H "X-User-Id: 1" http://localhost:8080/surveys

# Employee запрос (успешный)
curl -H "X-User-Id: 2" http://localhost:8080/surveys

# Admin действие (успешное)
curl -X POST -H "X-User-Id: 1" -d '{"title":"Test"}' http://localhost:8080/surveys

# Employee действие (ошибка)
curl -X POST -H "X-User-Id: 2" -d '{"title":"Test"}' http://localhost:8080/surveys
```

---

## 📱 Frontend страницы

- Главная страница: `http://localhost:8080/`
- Вход: `http://localhost:8080/login`
- Прохождение опроса: `http://localhost:8080/surveys/1/take`
- Мои ответы: `http://localhost:8080/surveys/my`
- Управление опросами: `http://localhost:8080/admin/surveys`
- Редактирование опроса: `http://localhost:8080/admin/surveys/1`
- Результаты опроса: `http://localhost:8080/surveys/1/results`

---

**💡 Совет:** Используйте эти примеры для тестирования API и интеграции с другими системами.
