# Серверный режим

Ren Browser как веб-приложение без настольной оболочки. Доступ по HTTP из другого браузера.

## Когда использовать

- Домашний сервер или VPS с Reticulum
- Docker
- Общая машина без установки desktop-приложения
- Планшеты и телефоны в локальной сети

## Быстрый старт

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Откройте `http://localhost:8080` или IP хоста в Firefox, Chromium или Safari.

## Docker

Образ:

```
ghcr.io/quad4-software/renbrowser:latest
```

Пример:

```sh
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

Не монтируйте каталог Reticulum только для чтения. Подробности и Podman: [Настройка Reticulum](reticulum-setup.md#сервер-и-docker).

Локальная сборка:

```sh
task build:docker
task run:docker
```

## Флаги командной строки

| Флаг | Назначение |
|------|------------|
| `--host` | Адрес привязки (в server-сборке по умолчанию `0.0.0.0`) |
| `--port` | Порт HTTP (по умолчанию `8080`) |
| `--config` | Путь к конфигу Reticulum |
| `--trust-proxy` | Доверять `X-Forwarded-*` от прокси |
| `--base-path` | Префикс URL при размещении в подпути |
| `--public-mode` | Вкладки, история и избранное в `localStorage` браузера, а не в SQLite сервера |
| `--profile` | Именованный профиль |
| `--import-profile` / `--export-profile` | JSON профиля при старте |

## Переменные окружения

Читается файл `.env` в рабочем каталоге. Уже заданные в окружении переменные не перезаписываются.

| Переменная | Назначение |
|----------|------------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | Адрес привязки |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | Порт |
| `REN_BROWSER_CONFIG` / `RETICULUM_CONFIG` | Конфиг Reticulum |
| `REN_BROWSER_TRUST_PROXY` | `true` / `1` / `yes` |
| `REN_BROWSER_BASE_PATH` | Подпуть |
| `REN_BROWSER_PUBLIC_MODE` | Публичный режим |
| `REN_BROWSER_PROFILE` | Имя профиля |
| `REN_BROWSER_IMPORT_PROFILE` | Импорт при старте |
| `REN_BROWSER_EXPORT_PROFILE` | Экспорт при старте |

## Публичный режим

Без `--public-mode` вкладки, история и избранное в SQLite на диске сервера. Все клиенты одного инстанса видят одни данные.

С `--public-mode` эти данные в `localStorage` каждого браузера. Удобно, когда много людей ходят на один сервер и не должны делить один профиль.

## Обратный прокси

Типичная схема с nginx или Caddy:

1. TLS на прокси
2. Проксирование на `127.0.0.1:8080`
3. Заголовки `X-Forwarded-Proto` и `X-Forwarded-Host`
4. Запуск с `--trust-proxy`
5. `--base-path`, если приложение не в корне домена

При включённом trust proxy учитывается заголовок `X-RenBrowser-Base-Path`.

## Нет встроенной аутентификации

Любой, кто достучится до HTTP-порта, может пользоваться браузером и генерировать трафик Reticulum. Не выставляйте 8080 в интернет без:

- правил файрвола
- VPN
- прокси с авторизацией
- или всего перечисленного

Перед публикацией прочитайте [Безопасность](security.md).

## Переопределение ассетов (для разработки)

- `--assets-dir path`
- `--assets-zip path`

Переменные: `REN_BROWSER_ASSETS_DIR`, `REN_BROWSER_ASSETS_ZIP`.

## Дальше

- [Данные и профили](data-and-profiles.md)
- [Безопасность](security.md)
- [Установка](installation.md)
