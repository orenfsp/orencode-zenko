# Frontend

Frontend личного кабинета Федерации спортивного программирования на Next.js 16, React 19 и Tailwind CSS 4.

## Требования

- Node.js 22+
- npm
- запущенный backend; по умолчанию ожидается `http://localhost:8080`

## Локальный запуск

Перейдите в папку frontend:

```bash
cd frontend
```

Создайте локальный файл окружения:

```bash
cp .env.example .env
```

Минимальная настройка `.env`:

```dotenv
BACKEND_URL=http://localhost:8080
```

Установите зависимости и запустите dev-сервер:

```bash
npm ci
npm run dev
```

Приложение будет доступно по адресу [http://localhost:3000](http://localhost:3000). Если порт занят, Next.js выберет следующий свободный порт.

Запросы frontend к `/api/v1/*` проксируются на `BACKEND_URL` через rewrites из `next.config.ts`, поэтому отдельная настройка CORS для локального запуска не требуется.

## Переменные окружения

| Переменная | Назначение | Значение по умолчанию |
| --- | --- | --- |
| `BACKEND_URL` | Адрес backend для проксирования `/api/v1/*` | `http://localhost:8080` |
| `NEXT_PUBLIC_PERSONAL_CONSENT_TEMPLATE_URL` | Ссылка на шаблон согласия на обработку данных | пусто |
| `NEXT_PUBLIC_PARENTAL_CONSENT_TEMPLATE_URL` | Ссылка на шаблон согласия законного представителя | пусто |

Переменные с префиксом `NEXT_PUBLIC_` доступны в браузере, поэтому в них нельзя хранить секреты.

## Production-сборка

```bash
npm run build
npm run start
```

По умолчанию production-сервер также запускается на `http://localhost:3000`.

## Docker

Из папки frontend:

```bash
docker build --build-arg BACKEND_URL=http://server:8080 -t orencode-zenko-frontend .
docker run --rm -p 3000:3000 orencode-zenko-frontend
```

Для обращения к backend на хостовой машине укажите доступный из контейнера адрес вместо `http://server:8080`.

## Проверки

```bash
npm run lint
npm run build
```

Доступные npm-скрипты:

- `npm run dev` — локальная разработка;
- `npm run build` — production-сборка и проверка TypeScript;
- `npm run start` — запуск собранного приложения;
- `npm run lint` — проверка ESLint.

