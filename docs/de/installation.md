# Installation

Diese Seite behandelt vorgefertigte Downloads, Docker und den Bau aus dem Quellcode.

## Vorgefertigte Downloads (empfohlen)

Laden Sie das neueste Release für Ihr System von [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) herunter.

| Plattform | Datei | Hinweise |
|-----------|-------|----------|
| Linux x86_64 | `renbrowser-linux-amd64.AppImage` | `chmod +x`, dann ausführen. Eine einfache Binärdatei ist ebenfalls enthalten. |
| Linux ARM64 | `renbrowser-linux-arm64.AppImage` | Dieselben Schritte wie bei x86_64. |
| Windows | `renbrowser-windows-amd64.exe` | Direkt ausführen. Kein Installer nötig. |
| macOS | `renbrowser-macos-universal.zip` | Entpacken und `renbrowser.app` öffnen. |
| Server (Linux x86_64) | `renbrowser-server-linux-amd64` | Headless-Binärdatei für Docker oder Self-Hosting. |
| Android | `renbrowser.apk` | Wenn die Release-Pipeline sie enthält. |

Jedes Release enthält `SHA256SUMS.txt`, damit Sie Downloads prüfen können. Siehe [Sicherheit](security.md).

### Download prüfen (Linux oder macOS)

```sh
sha256sum -c SHA256SUMS.txt
```

Prüfen Sie nur die Datei, die Sie heruntergeladen haben, wenn die Summendatei viele Assets auflistet.

### Systemvoraussetzungen

| Paket | Voraussetzungen auf dem Host |
|-------|------------------------------|
| **Linux AppImage** | Bindelt GTK 4, WebKitGTK 6 und weitere Bibliotheken. Kein separates WebKit nötig. Manche Distributionen brauchen FUSE oder `APPIMAGE_EXTRACT_AND_RUN=1`. |
| **Linux Flatpak** | Flatpak plus die Laufzeit `org.gnome.Platform` (GTK 4 und WebKitGTK 6). |
| **Linux-Binärdatei** | GTK 4 und WebKitGTK 6.0 zur Laufzeit (z. B. Debian/Ubuntu 24.04+, Fedora, Arch). |
| **Windows `.exe`** | [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/). Meist auf Windows 10/11 vorhanden. Der NSIS-Installer kann sie nachinstallieren; die portable `.exe` nicht. |
| **macOS `.app`** | Aktuelles macOS mit systemeigenem WebKit (kein Extra-Runtime). |
| **Android APK** | Android 5.0+ (API 21+). |
| **Server-Binärdatei / Docker** | Kein Desktop-GUI-Stack. UI im Browser des Hosts öffnen. |

## Docker oder Podman (Servermodus)

Offizielles Image: `ghcr.io/quad4-software/renbrowser`

Binden Sie Reticulum-Konfiguration und Profildaten ein. Das Image läuft nicht als root; geben Sie UID/GID des Hosts an:

```sh
mkdir -p "$HOME/.reticulum-go" "$HOME/.renbrowser"
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

Dieselben Flags gelten für `podman run`. Unter Podman können Sie `--userns=keep-id` statt `--user "$(id -u):$(id -g)"` verwenden. Bei SELinux `:Z` an die Volume-Flags anhängen.

Öffnen Sie `http://localhost:8080` in einem beliebigen Browser auf demselben Rechner.

Aus diesem Repository bauen und starten:

```sh
task build:docker
task run:docker
```

Das Server-Image hat **keinen Anmeldebildschirm**. Stellen Sie es nur in Netzwerken bereit, denen Sie vertrauen. Siehe [Servermodus](server-mode.md) und [Sicherheit](security.md).

## Bau aus dem Quellcode

Für Mitwirkende oder Plattformen ohne vorgefertigtes Paket.

### Voraussetzungen

- [Go](https://go.dev/) 1.26 oder neuer
- [Node.js](https://nodejs.org/) 22+ und [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/) (empfohlen)
- Reticulum-Konfiguration unter `~/.reticulum-go/` (oder `REN_BROWSER_CONFIG` setzen)

### Einfacher Build

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Go-Module laden Quad4-Abhängigkeiten automatisch von GitHub.

### Plattformspezifische Builds

```sh
task build:windows
task build:darwin
task build:android      # physisches Gerät (arm64)
task build:android:emu  # Emulator (Host-ABI)
```

### Installer und Pakete

```sh
task package                  # aktuelles Betriebssystem
task package:linux:appimage   # Linux AppImage
task package:darwin:universal # macOS universal
task package:windows          # Windows NSIS-Installer
```

### Android SDK

Android-Builds benötigen das [Android SDK](https://developer.android.com/studio) (API 34, NDK r26+). Setzen Sie `ANDROID_HOME` und führen Sie `task android:install:deps` aus, wenn der Build fehlende Tools meldet.

## Server-Binärdatei aus dem Quellcode

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Siehe [Servermodus](server-mode.md) für Umgebungsvariablen und Hinweise zum Deployment.

## Nach der Installation

1. Prüfen Sie, ob die Reticulum-Konfiguration vorhanden ist ([Reticulum einrichten](reticulum-setup.md))
2. Starten Sie die App und öffnen Sie `about:`
3. Lesen Sie [Browser verwenden](using-the-browser.md)
