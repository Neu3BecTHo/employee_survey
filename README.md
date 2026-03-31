# Employee Survey Web App

Веб-приложение для проведения внутренних опросов сотрудников с ролевой системой доступа.

## 📋 Описание

Система позволяет администраторам создавать опросы, а сотрудникам проходить их и просматривать свои ответы. Приложение реализовано на Go (backend) и PostgreSQL (база данных) с современным frontend на JavaScript.

## 🚀 Быстрый старт (Docker)

### 1. Настройка окружения

```bash
# Создайте .env файл из примера
cp .env.example .env

# Отредактируйте при необходимости
nano .env
```

### 2. Запуск приложения

```bash
# Запуск через Docker Compose
docker-compose up -d

# Или с пересборкой (после изменений)
docker-compose up -d --build
```

### 3. Откройте в браузере
```
http://localhost:8080
```

**Готово!** Приложение доступно по адресу http://localhost:8080

## 🛠 Технологический стек

### Backend:
- **Go 1.24** - основной язык программирования
- **PostgreSQL 15** - база данных
- **Gorilla Mux** - HTTP роутер
- **database/sql** - работа с БД
- **golang-migrate** - миграции
- **godotenv** - загрузка .env файлов

### Frontend:
- **HTML5** - семантическая разметка
- **CSS3** - стилизация
- **JavaScript ES6+** - логика интерфейса
- **Fetch API** - работа с backend

### Инфраструктура:
- **Docker** - контейнеризация
- **Docker Compose** - оркестрация
- **JSONB** - хранение вариантов ответа
- **Terser** - минификация JS
- **CleanCSS** - минификация CSS

## 📁 Структура проекта

```
employee_survey/
├── main.go                      # Главный файл приложения
├── handlers.go                  # HTTP обработчики
├── internal/
│   └── models.go               # Модели данных
├── migrations/                  # Миграции БД
├── static/
│   ├── css/                    # Стили (base.css, components.css)
│   ├── js/                     # JavaScript файлы
│   └── templates/              # HTML шаблоны
├── docker-compose.yml          # Docker конфигурация
├── Dockerfile                  # Docker образ
├── .env.example                # Шаблон переменных окружения
├── .env                        # Локальные переменные (не в git)
├── DEV_MODE.md                 # Документация DEV_MODE
├── API_EXAMPLES.md             # Примеры API запросов
├── CHECKLIST.md                # Чек-лист тестирования
└── README.md                   # Этот файл
```

## ⚙️ Конфигурация (.env)

Все настройки в одном файле `.env`:

```bash
# Режим разработки (true = обычные файлы, false = минифицированные)
DEV_MODE=true

# Настройки базы данных
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=survey_user
DB_PASSWORD=survey_pass
DB_NAME=survey_app

# Переменные для postgres контейнера
POSTGRES_USER=${DB_USER}
POSTGRES_PASSWORD=${DB_PASSWORD}
POSTGRES_DB=${DB_NAME}
```

## 🗜 Минификация файлов

### Установка инструментов

```bash
# Минификация JS
npm install -g terser

# Минификация CSS  
npm install -g clean-css-cli
```

### Минификация всех файлов

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

### Режимы работы

**Разработка (DEV_MODE=true):**
- Загружаются обычные файлы: `app.js`, `base.css`
- Удобно для отладки

**Продакшен (DEV_MODE=false):**
- Загружаются минифицированные: `app.min.js`, `base.min.css`
- Экономия трафика ~30-40%

Подробнее в [DEV_MODE.md](DEV_MODE.md)

## 🚀 Разработка без Docker

### Требования:
- Go 1.24+
- PostgreSQL 15+
- Node.js (для минификации)

### Шаги:

```bash
# 1. Установите зависимости Go
go mod download

# 2. Запустите PostgreSQL локально
# или используйте docker-compose up postgres

# 3. Создайте .env файл
cp .env.example .env

# 4. Запустите приложение
go run .
```

Приложение автоматически выполнит миграции БД при первом запуске.

## 📊 API Эндпоинты

### Публичные эндпоинты:
- `GET /users` - список пользователей
- `GET /surveys` - список опросов
- `GET /surveys/{id}` - детали опроса
- `POST /surveys/{id}/responses` - отправка ответов
- `GET /surveys/{id}/results` - результаты опроса
- `GET /surveys/my` - мои ответы
- `GET /surveys/responses/{id}` - детали ответа

### Администраторские эндпоинты:
- `POST /surveys` - создание опроса
- `PUT /surveys/{id}` - обновление опроса
- `POST /surveys/{id}/questions` - добавление вопроса
- `PUT /surveys/{id}/questions/{qid}` - обновление вопроса
- `DELETE /surveys/{id}/questions/{qid}` - удаление вопроса
- `POST /surveys/{id}/open` - открытие опроса
- `POST /surveys/{id}/close` - закрытие опроса

Подробные примеры запросов в [API_EXAMPLES.md](API_EXAMPLES.md)

## 🔐 Ролевая система

### Администратор (admin):
- Полный доступ ко всем функциям
- Создание и управление опросами
- Просмотр всех результатов

### Сотрудник (employee):
- Просмотр открытых опросов
- Прохождение опросов
- Просмотр только своих ответов

## 📱 Frontend страницы

- `/` - главная страница со списком опросов
- `/login` - выбор пользователя для входа
- `/surveys/{id}/take` - прохождение опроса
- `/surveys/{id}/already-responded` - сообщение о повторном прохождении
- `/surveys/my` - мои ответы (сотрудник)
- `/surveys/responses/{id}` - детали моего ответа
- `/admin/surveys` - управление опросами (администратор)
- `/admin/surveys/{id}` - редактирование опроса
- `/surveys/{id}/results` - результаты опроса

## 🧪 Тестирование

### Чек-лист ручной проверки:
См. [CHECKLIST.md](CHECKLIST.md) - полный чек-лист для тестирования всех функций.

### API тесты через curl:
См. [API_EXAMPLES.md](API_EXAMPLES.md) - примеры запросов и ответов.

### Быстрая проверка:
```bash
# Проверка API
curl http://localhost:8080/users

# Проверка с авторизацией
curl -H "X-User-Id: 1" http://localhost:8080/surveys
```

## 📝 База данных

### Основные таблицы:
- `users` - пользователи
- `surveys` - опросы
- `survey_questions` - вопросы опросов
- `survey_responses` - ответы пользователей
- `survey_answers` - конкретные ответы

### Индексы:
- `survey_questions(survey_id)`
- `survey_responses(survey_id, user_id)` - уникальный
- `survey_answers(response_id, question_id)`

## 🚨 Валидация и ограничения

### Обязательные поля:
- Название опроса
- Текст вопроса
- Ответы на обязательные вопросы

### Ограничения:
- Один пользователь не может отвечать дважды на один опрос
- Нельзя отправить пустые ответы
- Вопросы можно добавлять только в статусе "draft"
- Только admin может создавать/редактировать опросы

## 🐛 Отладка

### Логи Docker:
```bash
# Логи приложения
docker-compose logs -f app

# Логи базы данных
docker-compose logs -f postgres
```

### Пересборка:
```bash
# Полная пересборка
docker-compose down
docker-compose up -d --build

# Удаление данных БД (осторожно!)
docker-compose down -v
```

## 🤝 Вклад в проект

1. Fork проекта
2. Создайте feature branch
3. Внесите изменения
4. Минифицируйте файлы если нужно
5. Обновите документацию
6. Push в branch
7. Создайте Pull Request

## 📄 Лицензия

MIT License

## 📞 Поддержка

Для вопросов и предложений создайте Issue в репозитории.

---

**Разработано с ❤️ для внутренних опросов сотрудников**
