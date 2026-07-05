# Разработка

Локальная работа над Ren Browser.

## Требования

- Go 1.26+
- Node.js 22+ и pnpm 11+
- Task (рекомендуется)
- Конфиг Reticulum для живых mesh-тестов (опционально)

## Клонирование и запуск

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task dev
```

`task dev` запускает Wails в режиме разработки с hot reload Vite на порту 9245 по умолчанию.

## Частые задачи

| Задача | Назначение |
|------|------------|
| `task dev` | Настольная разработка |
| `task build` | Продакшен-сборка для текущей ОС |
| `task run` | Запуск собранного бинарника |
| `task check` | Полный quality gate (Go + фронтенд) |
| `task test` | Все юнит-тесты |
| `task test:interop` | Живые тесты Reticulum (нужна сеть) |
| `task build:server` | Серверный бинарник |
| `task run:server` | Локальный сервер |
| `task package` | Установщик или bundle |

## Quality gate

Перед отправкой изменений:

```sh
task check
```

Включает:

- `gofmt` для Go
- `go test ./...`
- gosec
- проверку бренда
- typecheck, lint, format, knip, audit и vitest для фронтенда

Дополнительно:

```sh
task test:go:race
task test:go:hard
task fuzz:go
```

## Только фронтенд

```sh
cd frontend
pnpm install
pnpm check
pnpm test
```

Биндинги в `frontend/bindings/` генерирует Wails из Go-сервисов.

## Структура Go

| Путь | Роль |
|------|------|
| `main_desktop.go` | Точка входа desktop (окно Wails) |
| `main_server.go` | Точка входа сервера (только HTTP) |
| `internal/app/` | BrowserService для UI |
| `internal/rns/` | Обёртка стека Reticulum |
| `internal/nomadnet/` | Загрузка страниц по LXMF. Узлы NomadNet определяются по анонсам, без библиотек NomadNet |
| `internal/micron/` | Рендер Micron |
| `internal/store/` | API SQLite |
| `internal/plugins/` | Хост расширений |
| `frontend/` | UI на Svelte 5 |

Подробнее: [Архитектура](architecture.md).

## Разработка плагинов

```sh
task test:plugins
```

Кладите dev-плагин в `~/.renbrowser/plugins/` и перезапускайте приложение.

## Interop-тесты

```sh
task test:interop
```

Нужна живая сеть Reticulum и тег `interop`. Для чисто UI-правок можно пропустить.

## Отладка ассетов

`REN_BROWSER_ASSET_PROBE=1` : какой загрузчик ассетов активен (встроенный или с диска).

## Заголовки SPDX

Новые Go-файлы:

```go
// SPDX-License-Identifier: MIT
```

Следуйте стилю соседних файлов.

## Дальше

- [Архитектура](architecture.md)
- [Участие в проекте](contributing.md)
- [Расширения](extensions.md)
