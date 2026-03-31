# DEV_MODE - Режим разработки

Для удобной разработки реализована поддержка переключения между минифицированными и обычными файлами через `.env` файл.

## Как это работает

При `DEV_MODE=true` в `.env` файле загружаются обычные (неминифицированные) JS/CSS файлы:
- `app.js`, `login.js`, `surveys.js` и т.д.
- `base.css`, `components.css` и т.д.

При `DEV_MODE=false` или без `.env` файла:
- `app.min.js`, `login.min.js`, `surveys.min.js` и т.д.
- `base.min.css`, `components.min.css` и т.д.

## Настройка

### 1. Создайте .env файл из примера:
```bash
cp .env.example .env
```

### 2. Отредактируйте `.env`:
```bash
# Для разработки (обычные файлы)
DEV_MODE=true

# Для продакшена (минифицированные файлы)
DEV_MODE=false
```

## Использование

Просто запустите приложение — `.env` файл загрузится автоматически:
```bash
go run .
```

### Docker Compose
```yaml
services:
  app:
    environment:
      - DEV_MODE=true  # или false для продакшена
```

## Минификация файлов

### Установка инструментов

```bash
# Установка terser для JS
npm install -g terser

# Установка clean-css для CSS
npm install -g clean-css-cli
```

### Минификация JS файлов

```bash
cd static/js

# Минификация одного файла
npx terser app.js -o app.min.js -c -m

# Минификация всех файлов
for f in *.js; do
  if [[ ! $f =~ \.min\.js$ ]]; then
    npx terser "$f" -o "${f%.js}.min.js" -c -m
  fi
done
```

### Минификация CSS файлов

```bash
cd static/css

# Минификация одного файла
npx clean-css-cli -o base.min.css base.css

# Минификация всех файлов
for f in *.css; do
  if [[ ! $f =~ \.min\.css$ ]]; then
    npx clean-css-cli -o "${f%.css}.min.css" "$f"
  fi
done
```

### Результаты минификации

| Файл | Оригинал | Минифицированный | Экономия |
|------|----------|------------------|----------|
| `app.js` | ~9 KB | ~5 KB | ~45% |
| `surveys.js` | ~7 KB | ~5 KB | ~30% |
| `base.css` | ~34 KB | ~25 KB | ~25% |
| `components.css` | ~9 KB | ~7 KB | ~20% |

## Структура шаблонов

Все HTML шаблоны используют условную загрузку:

```html
{{if .DevMode}}
<link rel="stylesheet" href="/static/css/base.css">
<link rel="stylesheet" href="/static/css/components.css">
<script src="/static/js/app.js"></script>
{{else}}
<link rel="stylesheet" href="/static/css/base.min.css">
<link rel="stylesheet" href="/static/css/components.min.css">
<script src="/static/js/app.min.js"></script>
{{end}}
```

## Переменные окружения

### DEV_MODE
- **true** - режим разработки (обычные файлы)
- **false** - продакшен (минифицированные файлы)

### DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
Настройки подключения к PostgreSQL.

### POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
Настройки для инициализации postgres контейнера.

## Рекомендации

### Разработка
- Используйте `DEV_MODE=true`
- Работайте с обычными файлами
- Удобно для отладки в DevTools

### Продакшен
- Установите `DEV_MODE=false`
- Убедитесь, что все `.min.js` и `.min.css` файлы созданы
- Проверьте загрузку в DevTools → Network

## Безопасность

`.env` файл добавлен в `.gitignore` и не будет коммититься в git. Для команды используйте `.env.example` как шаблон.

## Проверка работы

1. Откройте DevTools (F12)
2. Перейдите во вкладку Network
3. Обновите страницу
4. Проверьте загружаемые файлы:
   - При `DEV_MODE=true`: `app.js`, `base.css`
   - При `DEV_MODE=false`: `app.min.js`, `base.min.css`
