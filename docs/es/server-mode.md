# Modo servidor

El modo servidor ejecuta Ren Browser como aplicación web sin el shell de escritorio. Accedes desde otro navegador en una URL HTTP.

## Cuándo usar el modo servidor

- Homelab o VPS que ya ejecuta Reticulum
- Despliegues con Docker
- Máquina compartida donde prefieres no instalar la aplicación de escritorio
- Acceso desde tablets o teléfonos en tu LAN

## Inicio rápido

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Abre `http://localhost:8080` (o la IP de tu host) en Firefox, Chromium o Safari.

## Docker

Imagen publicada:

```
ghcr.io/quad4-software/renbrowser:latest
```

Ejemplo de ejecución:

```sh
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

No montes el directorio Reticulum en solo lectura. Detalles y notas de Podman: [Configuración de Reticulum](reticulum-setup.md#servidor-y-docker).

Compilar localmente:

```sh
task build:docker
task run:docker
```

## Flags de línea de comandos

| Flag | Propósito |
|------|-----------|
| `--host` | Dirección de enlace (predeterminado `0.0.0.0` en la compilación de servidor) |
| `--port` | Puerto HTTP (predeterminado `8080`) |
| `--config` | Ruta de configuración de Reticulum |
| `--trust-proxy` | Confiar en `X-Forwarded-*` de un proxy inverso |
| `--base-path` | Prefijo de URL cuando se sirve bajo una subruta |
| `--public-mode` | Guardar favoritos, historial y pestañas en `localStorage` del navegador en lugar de SQLite del servidor |
| `--profile` | Base de datos de perfil con nombre |
| `--import-profile` / `--export-profile` | Perfil JSON al arrancar |

## Variables de entorno

El servidor lee un archivo `.env` en el directorio de trabajo. Las variables ya definidas en el entorno no se sobrescriben.

| Variable | Propósito |
|----------|-----------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | Dirección de enlace |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | Puerto |
| `REN_BROWSER_CONFIG` / `RETICULUM_CONFIG` | Configuración de Reticulum |
| `REN_BROWSER_TRUST_PROXY` | `true` / `1` / `yes` para activar trust proxy |
| `REN_BROWSER_BASE_PATH` | Prefijo de subruta |
| `REN_BROWSER_PUBLIC_MODE` | Interruptor de modo público |
| `REN_BROWSER_PROFILE` | Nombre de perfil |
| `REN_BROWSER_IMPORT_PROFILE` | Ruta de importación al arrancar |
| `REN_BROWSER_EXPORT_PROFILE` | Ruta de exportación al arrancar |

## Modo público

Sin `--public-mode`, el servidor guarda pestañas, historial y favoritos en su base de datos SQLite en el disco del servidor. Cada cliente que comparte esa instancia ve los mismos datos.

Con `--public-mode`, esos elementos viven en el `localStorage` de cada navegador. Úsalo cuando muchas personas accedan a un servidor y no deban compartir un perfil.

## Proxy inverso

Configuración típica de nginx o Caddy:

1. Termina TLS en el proxy
2. Haz proxy a `127.0.0.1:8080`
3. Pasa `X-Forwarded-Proto` y `X-Forwarded-Host`
4. Inicia Ren Browser con `--trust-proxy`
5. Define `--base-path` si la aplicación no está en la raíz del dominio

La cabecera `X-RenBrowser-Base-Path` se reconoce cuando trust proxy está activo.

## Sin autenticación integrada

Cualquiera que pueda alcanzar el puerto HTTP puede usar el navegador y disparar tráfico de Reticulum. No expongas el puerto 8080 a internet público sin:

- Reglas de firewall
- VPN
- Proxy inverso con autenticación
- O todo lo anterior

Lee [Seguridad](security.md) antes de publicar un servidor.

## Sobrescritura de activos (avanzado)

Para desarrollo puedes servir archivos del frontend desde disco o zip en lugar de activos embebidos:

- `--assets-dir path`
- `--assets-zip path`

Entorno: `REN_BROWSER_ASSETS_DIR`, `REN_BROWSER_ASSETS_ZIP`.

## Próximos pasos

- [Datos y perfiles](data-and-profiles.md) para SQLite frente a modo público
- [Seguridad](security.md) para endurecer el despliegue
- [Instalación](installation.md) para binarios de release
