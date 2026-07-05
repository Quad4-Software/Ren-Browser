# Архитектура

Общая схема устройства Ren Browser.

## Обзор

```
┌─────────────────────────────────────────────────────────┐
│  Фронтенд Svelte 5 (webview Wails или браузер на сервере)│
│  Вкладки, chrome, просмотр Micron, панели, настройки    │
└───────────────────────┬─────────────────────────────────┘
                        │ Биндинги Wails / HTTP API
┌───────────────────────▼─────────────────────────────────┐
│  internal/app : BrowserService                          │
│  Навигация, вкладки, история, настройки, мост к плагинам│
└───────┬─────────────┬─────────────┬─────────────────────┘
        │             │             │
        ▼             ▼             ▼
   internal/rns   internal/store  internal/plugins
   Reticulum      SQLite          Расширения
        │
        ▼
   internal/nomadnet : загрузка страниц по LXMF для узлов с анонсом NomadNet
        │
        ▼
   internal/micron : разбор разметки и HTML
```

## Точки входа

| Файл | Тег сборки | Роль |
|------|------------|------|
| `main_desktop.go` | `!server && !android` | Окно Wails, встроенный `frontend/dist` |
| `main_server.go` | `server` | HTTP-сервер, те же ассеты |
| Android main | `android` | Мобильная оболочка |

`internal/bootstrap` связывает конфиг, store, плагины и приложение Wails.

## Фронтенд

- **Фреймворк:** Svelte 5 с Vite
- **Биндинги:** `frontend/bindings/renbrowser/`
- **Главный UI:** `frontend/src/App.svelte`
- **Компоненты:** `frontend/src/lib/components/`
- **Логика браузера:** `frontend/src/lib/browser/`

Рендер Micron может использовать WASM с проверкой SRI (`MicronWasmManager`).

## Сервисы backend

### BrowserService (`internal/app`)

Центральный API для UI:

- навигация и вкладки
- обнаружение, история, избранное
- настройки
- мост к хосту плагинов

### Стек Reticulum (`internal/rns`)

Обёртка `quad4/reticulum-go`:

- старт и останов транспорта
- статистика интерфейсов
- hot reload из Настроек

### Загрузка страниц (`internal/nomadnet`)

Загрузка удалённых `.mu` и связанного контента по LXMF и Reticulum. В Обнаружении узел помечается как NomadNet по анонсу. Библиотеки клиента NomadNet не используются.

### Store (`internal/store` + `internal/db`)

SQLite с миграцией из `state.json`.

### Плагины (`internal/plugins`)

- валидация манифеста
- разрешения
- встроенные схемы (`about:`, `license:`, `editor:`)
- runtime JS и WASM

## Контент и рендер

| Пакет | Роль |
|-------|------|
| `internal/content` | Статические страницы |
| `internal/micron` | Micron в HTML |
| `internal/micronwasm` | WASM-парсер |
| `internal/cache` | Кэш страниц |

## Middleware сервера

`internal/servermw` : base path и URL за прокси в серверном режиме.

## Конфигурация

`internal/config` : флаги, `.env`, `REN_BROWSER_*` в структуру `Runtime` для bootstrap.

## Бренд и пути

`internal/brand` (из `build/brand.yml`):

- каталог `.renbrowser`
- файл `renbrowser.db`
- имя и версия в UI

## Сборка и упаковка

- `Taskfile.yml` : команды разработчика
- `build/` : AppImage, NSIS, macOS, Android, Docker
- `build/config.yml` : конфиг Wails

## CI

GitHub Actions: тесты Go и фронтенда, security-сканы, smoke-сборки, релизные артефакты. См. `.github/workflows/`.

## Точки расширения

1. **Плагины** : манифест, UI и схемы
2. **Темы** : JSON токенов
3. **Сообщественные интерфейсы** : фрагменты конфига Reticulum в Настройках

## Дальше

- [Разработка](development.md)
- [Расширения](extensions.md)
