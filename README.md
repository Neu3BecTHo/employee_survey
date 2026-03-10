# Employee Survey Web App

Веб-приложение для проведения внутренних опросов сотрудников на Go + JavaScript.

## Описание проекта

Приложение позволяет администраторам создавать опросы, а сотрудникам - проходить их. Реализована система ролей, валидация ответов и сбор результатов.

## Основные возможности

- **Ролевая система**: Администратор и сотрудник
- **Создание опросов**: Текстовые и выборные вопросы
- **Прохождение опросов**: Валидация обязательных полей
- **Результаты**: Агрегированные данные по ответам
- **Современный интерфейс**: Адаптивный дизайн

## Технологии

- **Backend**: Go + PostgreSQL + Gorilla Mux
- **Frontend**: HTML5 + CSS3 + JavaScript (ES6+)
- **База данных**: PostgreSQL с миграциями
- **Контейнеризация**: Docker Compose

## Структура проекта

```
survey-app/
├── docker-compose.yml    # Конфигурация сервисов
├── Dockerfile           # Docker образ для Go
├── main.go             # Основной сервер
├── main_test.go        # Unit тесты
├── handlers.go         # HTTP handlers
├── internal/
│   └── models.go       # Структуры данных
├── migrations/         # Миграции БД
│   ├── 000001_create_tables.up.sql
│   └── 000001_create_tables.down.sql
└── static/             # Frontend файлы
    ├── index.html      # Главная страница
    ├── login.html      # Страница входа
    ├── take-survey.html # Прохождение опроса
    ├── survey-results.html # Результаты
    ├── css/
    │   └── style.css   # Стили
    └── js/
        ├── app.js      # Общая логика
        ├── login.js    # Логика входа
        ├── surveys.js  # Управление опросами
        ├── take-survey.js # Прохождение опроса
        └── survey-results.js # Просмотр результатов
```

## Запуск проекта

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+ (для локального запуска)

### 1. Клонирование и запуск

```bash
# Запуск с Docker Compose (рекомендуется)
docker-compose up --build

# Или локально (нужен PostgreSQL)
go mod tidy
go run main.go
```

### 2. Доступ к приложению

- **Frontend**: http://localhost:8080
- **API**: http://localhost:8080/api/*
- **База данных**: localhost:5432

### 3. Тестовые пользователи

При первом запуске создаются тестовые пользователи:

| ID | Имя | Роль |
|----|-----|------|
| 1  | Иван Иванов | admin |
| 2  | Петр Петров | employee |
| 3  | Анна Сидорова | employee |

## API Документация

### Аутентификация

Все запросы требуют заголовка `X-User-Id` с ID пользователя.

### Эндпоинты

#### Пользователи
- `GET /users` - Получить список всех пользователей

#### Опросы
- `GET /surveys` - Получить список опросов (фильтруется по роли)
- `POST /surveys` - Создать опрос (admin)
- `GET /surveys/{id}` - Получить опрос с вопросами
- `PUT /surveys/{id}` - Обновить опрос (admin)
- `POST /surveys/{id}/questions` - Добавить вопрос (admin)
- `POST /surveys/{id}/open` - Открыть опрос (admin)
- `POST /surveys/{id}/close` - Закрыть опрос (admin)

#### Ответы
- `POST /surveys/{id}/responses` - Отправить ответы (employee)
- `GET /surveys/{id}/results` - Получить результаты (admin)

### Примеры запросов

#### Создание опроса
```bash
curl -X POST http://localhost:8080/surveys \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 1" \
  -d '{
    "title": "Обратная связь по проекту",
    "description": "Помогите нам улучшить наши процессы"
  }'
```

#### Добавление вопроса
```bash
curl -X POST http://localhost:8080/surveys/1/questions \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 1" \
  -d '{
    "text": "Как бы вы оценили коммуникацию в команде?",
    "type": "single_choice",
    "is_required": true,
    "options": ["Отлично", "Хорошо", "Удовлетворительно", "Плохо"]
  }'
```

#### Отправка ответа
```bash
curl -X POST http://localhost:8080/surveys/1/responses \
  -H "Content-Type: application/json" \
  -H "X-User-Id: 2" \
  -d '{
    "answers": [
      {
        "question_id": 1,
        "value": "Хорошо"
      }
    ]
  }'
```

## Тестирование

### Unit тесты

```bash
go test -v
```

### Ручное тестирование

#### Сценарий 1: Создание опроса (Администратор)
1. Зайти на http://localhost:8080/login
2. Выбрать "Иван Иванов (Администратор)"
3. Нажать "Создать опрос"
4. Заполнить название и описание
5. Добавить вопросы (текстовый и выборный)
6. Открыть опрос

#### Сценарий 2: Прохождение опроса (Сотрудник)
1. Войти как "Петр Петров (Сотрудник)"
2. Выбрать открытый опрос
3. Заполнить все обязательные поля
4. Отправить ответы
5. Проверить сообщение об успехе

#### Сценарий 3: Просмотр результатов (Администратор)
1. Войти как администратор
2. Открыть страницу результатов опроса
3. Проверить статистику и ответы

### Проверки валидации

- ✅ Нельзя отправить ответ на закрытый опрос
- ✅ Один пользователь не может отвечать дважды
- ✅ Обязательные вопросы должны быть заполнены
- ✅ Выборные вопросы проверяют допустимые варианты

## Разработка

### Добавление новых пользователей

В базе данных:
```sql
INSERT INTO users (name, role) VALUES ('Новый Пользователь', 'employee');
```

### Структура базы данных

```sql
-- Пользователи
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
);

-- Опросы
CREATE TABLE surveys (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'draft'
);

-- Вопросы
CREATE TABLE survey_questions (
    id SERIAL PRIMARY KEY,
    survey_id INTEGER REFERENCES surveys(id),
    text TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_required BOOLEAN DEFAULT false,
    options JSONB
);

-- Ответы
CREATE TABLE survey_responses (
    id SERIAL PRIMARY KEY,
    survey_id INTEGER REFERENCES surveys(id),
    user_id INTEGER REFERENCES users(id),
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE survey_answers (
    id SERIAL PRIMARY KEY,
    response_id INTEGER REFERENCES survey_responses(id),
    question_id INTEGER REFERENCES survey_questions(id),
    value TEXT
);
```

## Производительность и безопасность

- **SQL-инъекции**: Защита через параметризованные запросы
- **XSS**: Экранирование данных на frontend
- **CSRF**: Не реализован (можно добавить gorilla/csrf)
- **Валидация**: Серверная и клиентская валидация
- **Индексы**: Оптимизированы запросы к БД

## Будущие улучшения

- [ ] Добавление CSRF защиты
- [ ] Аутентификация через JWT
- [ ] Экспорт результатов в Excel/PDF
- [ ] Шаблоны опросов
- [ ] Уведомления по email
- [ ] Многоязычность
- [ ] API документация (Swagger)

## Лицензия

MIT License

## Контакты

Для вопросов и предложений создавайте issues в репозитории.
