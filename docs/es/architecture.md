# Arquitectura

Vista de alto nivel de cómo está estructurado Ren Browser.

## Resumen

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 5 frontend (Wails webview or browser in server) │
│  Tabs, chrome, Micron viewer, panels, settings          │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails bindings / HTTP API
┌───────────────────────▼─────────────────────────────────┐
│  internal/app : BrowserService                          │
│  Navigation, tabs, history, settings, plugin bridge     │
└───────┬─────────────┬─────────────┬─────────────────────┘
        │             │             │
        ▼             ▼             ▼
   internal/rns   internal/store  internal/plugins
   Reticulum      SQLite          Extensions
        │
        ▼
   internal/nomadnet : carga de páginas LXMF para nodos anunciados como NomadNet
        │
        ▼
   internal/micron : Markup parse and HTML render
```

## Puntos de entrada

| Archivo | Etiqueta de compilación | Rol |
|---------|------------------------|-----|
| `main_desktop.go` | `!server && !android` | Ventana Wails, `frontend/dist` embebido |
| `main_server.go` | `server` | Servidor HTTP, mismos activos embebidos |
| Android main | `android` | Shell móvil |

`internal/bootstrap` conecta configuración, store, plugins y la aplicación Wails.

## Frontend

- **Framework:** Svelte 5 con Vite
- **Bindings:** Generados bajo `frontend/bindings/renbrowser/`
- **UI principal:** `frontend/src/App.svelte` orquesta chrome y paneles
- **Componentes:** `frontend/src/lib/components/` (barra de pestañas, discovery, settings, etc.)
- **Lógica del navegador:** `frontend/src/lib/browser/` (URLs, atajos, errores)

El renderizado Micron puede usar analizadores WASM gestionados por `MicronWasmManager` con verificación SRI.

## Servicios del backend

### BrowserService (`internal/app`)

API central para la UI:

- Navegar URLs y gestionar estado de pestañas
- Exponer discovery, historial, favoritos
- Cargar y guardar preferencias
- Puente al host de plugins

### Stack Reticulum (`internal/rns`)

Envuelve `quad4/reticulum-go`:

- Iniciar y detener transporte
- Informar estadísticas de interfaces
- Recarga en caliente de configuración desde Settings

### Carga de páginas (`internal/nomadnet`)

Obtiene `.mu` remotos y contenido relacionado por LXMF y Reticulum. Discovery etiqueta nodos como NomadNet cuando el anuncio coincide. Ren Browser no usa bibliotecas del cliente NomadNet.

### Store (`internal/store` + `internal/db`)

Persistencia SQLite con migración desde `state.json` heredado.

### Plugins (`internal/plugins`)

- Validación de manifiesto
- Aplicación de permisos
- Esquemas integrados (`about:`, `license:`, `editor:`)
- Runtimes de plugins JS y WASM

## Contenido y renderizado

| Paquete | Rol |
|---------|-----|
| `internal/content` | Páginas estáticas (about, license) |
| `internal/micron` | Micron a HTML |
| `internal/micronwasm` | Integración del analizador WASM |
| `internal/cache` | Ayudantes de caché de página |

## Middleware del servidor

`internal/servermw` gestiona cabeceras de base path y construcción de URL consciente del proxy en modo servidor.

## Configuración

`internal/config` analiza flags, `.env` y variables `REN_BROWSER_*` en una estructura `Runtime` usada por bootstrap.

## Marca y rutas

`internal/brand` (generado desde `build/brand.yml`) define nombres estables:

- Directorio de datos `.renbrowser`
- Archivo de base de datos `renbrowser.db`
- Nombre mostrado y etiquetas de versión

## Compilación y empaquetado

- `Taskfile.yml` para comandos de desarrollador
- `build/` para empaquetado por SO (Linux AppImage, Windows NSIS, macOS, Android, Docker)
- `build/config.yml` para configuración del proyecto Wails

## CI

GitHub Actions ejecuta pruebas Go, comprobaciones de frontend, escaneos de seguridad, compilaciones de humo de escritorio y servidor, y artefactos de release. Consulta `.github/workflows/`.

## Puntos de extensión

1. **Plugins** con UI y esquemas guiados por manifiesto
2. **Temas** con archivos JSON de tokens
3. **Interfaces comunitarias** con fragmentos de configuración Reticulum en Settings

## Próximos pasos

- [Desarrollo](development.md) para compilar localmente
- [Extensiones](extensions.md) para la superficie de API de plugins
- Árbol de fuentes en la tabla de diseño del `README.md` del repositorio
