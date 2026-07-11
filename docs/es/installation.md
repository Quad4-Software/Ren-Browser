# Instalación

Esta página cubre descargas precompiladas, Docker y compilación desde el código fuente.

## Descargas precompiladas (recomendado)

Obtén el último release para tu sistema en [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases).

| Plataforma | Archivo | Notas |
|------------|---------|-------|
| Linux x86_64 | `renbrowser-linux-amd64.AppImage` | `chmod +x` y ejecuta. También se incluye un binario simple. |
| Linux ARM64 | `renbrowser-linux-arm64.AppImage` | Los mismos pasos que en x86_64. |
| Windows | `renbrowser-windows-amd64.exe` | Ejecuta directamente. No hace falta instalador. |
| macOS | `renbrowser-macos-universal.zip` | Descomprime y abre `renbrowser.app`. |
| Server (Linux x86_64) | `renbrowser-server-linux-amd64` | Binario sin interfaz gráfica para Docker o autoalojamiento. |
| Android | `renbrowser.apk` | Cuando el pipeline de release lo incluya. |

Cada release incluye `SHA256SUMS.txt` para verificar las descargas. Consulta [Seguridad](security.md).

### Verificar una descarga (Linux o macOS)

```sh
sha256sum -c SHA256SUMS.txt
```

Comprueba solo el archivo que descargaste si el archivo de sumas lista muchos activos.

### Requisitos del sistema

| Paquete | Requisitos en el host |
|---------|------------------------|
| **Linux AppImage** | Incluye GTK 4, WebKitGTK 6 y otras bibliotecas. No hace falta instalar WebKit aparte. Algunas distros necesitan FUSE o `APPIMAGE_EXTRACT_AND_RUN=1`. |
| **Linux Flatpak** | Flatpak más el runtime `org.gnome.Platform` (GTK 4 y WebKitGTK 6). |
| **Binario Linux** | GTK 4 y WebKitGTK 6.0 en tiempo de ejecución (p. ej. Debian/Ubuntu 24.04+, Fedora o Arch). |
| **Windows `.exe`** | [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/). Suele estar en Windows 10/11. El instalador NSIS puede instalarlo; el `.exe` portable no. |
| **macOS `.app`** | macOS reciente con WebKit del sistema (sin runtime extra). |
| **Android APK** | Android 5.0+ (API 21+). |
| **Binario servidor / Docker** | Sin stack gráfico de escritorio. Abre la UI en un navegador del host. |

## Docker o Podman (modo servidor)

Imagen oficial: `ghcr.io/quad4-software/renbrowser`

Monta la configuración de Reticulum y los datos de perfil. La imagen no corre como root; pasa el UID/GID del host:

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

Los mismos flags sirven con `podman run`. En Podman puedes usar `--userns=keep-id` en lugar de `--user "$(id -u):$(id -g)"`. Si SELinux bloquea el montaje, añade `:Z` a los volúmenes.

Abre `http://localhost:8080` en cualquier navegador de la misma máquina.

Compila y ejecuta desde este repositorio:

```sh
task build:docker
task run:docker
```

La imagen del servidor **no tiene pantalla de inicio de sesión**. Expónela solo en redes en las que confíes. Consulta [Modo servidor](server-mode.md) y [Seguridad](security.md).

## Compilar desde el código fuente

Para colaboradores o plataformas sin paquete precompilado.

### Requisitos

- [Go](https://go.dev/) 1.26 o más reciente
- [Node.js](https://nodejs.org/) 22+ y [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/) (recomendado)
- Configuración de Reticulum en `~/.reticulum-go/` (o define `REN_BROWSER_CONFIG`)

### Compilación básica

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Los módulos de Go descargan las dependencias de Quad4 desde GitHub automáticamente.

### Compilaciones por plataforma

```sh
task build:windows
task build:darwin
task build:android      # dispositivo físico (arm64)
task build:android:emu  # emulador (ABI del host)
```

### Instaladores y paquetes

```sh
task package                  # SO actual
task package:linux:appimage   # Linux AppImage
task package:darwin:universal # macOS universal
task package:windows          # instalador Windows NSIS
```

### Android SDK

Las compilaciones para Android necesitan el [Android SDK](https://developer.android.com/studio) (API 34, NDK r26+). Define `ANDROID_HOME` y ejecuta `task android:install:deps` si la compilación informa de herramientas faltantes.

## Binario de servidor desde el código fuente

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Consulta [Modo servidor](server-mode.md) para variables de entorno y notas de despliegue.

## Después de instalar

1. Confirma que la configuración de Reticulum está en su sitio ([Configuración de Reticulum](reticulum-setup.md))
2. Inicia la aplicación y abre `about:`
3. Lee [Uso del navegador](using-the-browser.md)
