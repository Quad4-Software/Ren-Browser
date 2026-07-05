# Установка

Здесь описаны готовые сборки, Docker и сборка из исходников.

## Готовые сборки (рекомендуется)

Скачайте последний релиз для вашей системы на [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases).

| Платформа | Файл | Примечания |
|----------|------|------------|
| Linux x86_64 | `renbrowser-linux-amd64.AppImage` | `chmod +x` и запуск. Есть и обычный бинарник. |
| Linux ARM64 | `renbrowser-linux-arm64.AppImage` | Как для x86_64. |
| Windows | `renbrowser-windows-amd64.exe` | Запуск без установщика. |
| macOS | `renbrowser-macos-universal.zip` | Распакуйте и откройте `renbrowser.app`. |
| Сервер (Linux x86_64) | `renbrowser-server-linux-amd64` | Headless для Docker или своего хостинга. |
| Android | `renbrowser.apk` | Если пайплайн релиза его публикует. |

К каждому релизу прилагается `SHA256SUMS.txt` для проверки. См. [Безопасность](security.md).

### Проверка загрузки (Linux или macOS)

```sh
sha256sum -c SHA256SUMS.txt
```

Можно проверить только нужный файл, если в списке много артефактов.

## Docker или Podman (серверный режим)

Официальный образ: `ghcr.io/quad4-software/renbrowser-server`

Смонтируйте конфиг Reticulum, чтобы контейнер мог войти в mesh:

```sh
docker run --rm -p 8080:8080 \
  -v "$HOME/.reticulum-go:/root/.reticulum-go:ro" \
  ghcr.io/quad4-software/renbrowser-server:latest
```

Откройте `http://localhost:8080` в любом браузере на той же машине.

Сборка и запуск из этого репозитория:

```sh
task build:docker
task run:docker
```

У серверного образа **нет экрана входа**. Открывайте порт только в сетях, которым доверяете. См. [Серверный режим](server-mode.md) и [Безопасность](security.md).

## Сборка из исходников

Для разработчиков или платформ без готового пакета.

### Требования

- [Go](https://go.dev/) 1.26 или новее
- [Node.js](https://nodejs.org/) 22+ и [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/) (рекомендуется)
- Конфиг Reticulum в `~/.reticulum-go/` (или `REN_BROWSER_CONFIG`)

### Базовая сборка

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Go-модули подтягивают зависимости Quad4 с GitHub автоматически.

### Сборка под платформы

```sh
task build:windows
task build:darwin
task build:android      # физическое устройство (arm64)
task build:android:emu  # эмулятор (ABI хоста)
```

### Установщики и пакеты

```sh
task package                  # текущая ОС
task package:linux:appimage   # Linux AppImage
task package:darwin:universal # macOS universal
task package:windows          # установщик NSIS для Windows
```

### Android SDK

Для Android нужен [Android SDK](https://developer.android.com/studio) (API 34, NDK r26+). Укажите `ANDROID_HOME` и при ошибках выполните `task android:install:deps`.

## Серверный бинарник из исходников

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Переменные окружения и развёртывание: [Серверный режим](server-mode.md).

## После установки

1. Проверьте конфиг Reticulum ([Настройка Reticulum](reticulum-setup.md))
2. Запустите приложение и откройте `about:`
3. Прочитайте [Работа с браузером](using-the-browser.md)
