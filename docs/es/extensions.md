# Extensiones

Ren Browser admite plugins que añaden esquemas de URL, paneles laterales, comandos, temas, páginas de configuración y pestañas de devtools.

## Instalar extensiones

### Desde Settings

1. Abre **Settings → Extensions**
2. Elige **Install extension**, luego **.zip**, **carpeta** o **módulo .wasm empaquetado**
3. Revisa la vista previa de instalación:
   - Permisos solicitados (puedes desactivar permisos individuales antes de instalar)
   - URLs externas que la extensión puede contactar (escaneadas del manifiesto y archivos del paquete)
   - Estado de la firma del publicador (sin firmar, firmada, publicador de confianza, no válida)
   - Advertencias de evaluación de seguridad
   - Idiomas de interfaz incluidos (cuando hay `locales/*.json`)
4. Confirma y activa la extensión

Las extensiones con `network.fetch` muestran un diálogo con los endpoints detectados. La lista de URLs sigue visible aunque desactives `network.fetch` antes de instalar.

### Instalación manual

Descomprime un plugin en:

```
~/.renbrowser/plugins/<id>/
```

La carpeta debe contener `renbrowser.plugin.json`. El `id` del manifiesto debe coincidir con el nombre de la carpeta.

## Extensiones de ejemplo

El repositorio incluye `extensions/hello-extension/`:

- Registra el esquema de URL `hello:`
- Añade un panel lateral **Hello**
- Define un comando **Say hello** con `mod+shift+h`

`extensions/micron-translator/` traduce páginas Micron (`.mu`) con Google Translate o LibreTranslate. Comandos: **Translate Micron page** (`mod+shift+t`) y **Restore original** (`mod+shift+r`).

## Archivo de manifiesto

Nombre del archivo: `renbrowser.plugin.json`

Campos obligatorios:

| Campo | Propósito |
|-------|-----------|
| `manifestVersion` | Actualmente `1` |
| `id` | Id único (`a-z`, `A-Z`, `0-9`, `.`, `-`, de 3 a 128 caracteres) |
| `name` | Nombre mostrado |
| `version` | Cadena semver |
| `main` | Script de entrada del frontend (opcional si solo hay backend) |
| `permissions` | Lista de capacidades (ver abajo) |

Campos opcionales: `description`, `author`, `license`, `engines`, `backend`, `network`, `contributes`.

### Restricción de motor

```json
"engines": { "renbrowser": ">=0.1.0" }
```

El host rechaza cargar el plugin si la versión de tu aplicación es demasiado antigua.

### Endpoints de red

Las extensiones con `network.fetch` deben declarar hosts o URLs contactados:

```json
"network": {
  "endpoints": [
    "https://api.example.com/",
    "URL de servicio configurada por el usuario"
  ]
}
```

Al instalar, RenBrowser también escanea `.js`, `.go`, `.wasm` y otros archivos del paquete en busca de URLs `http`/`https`.

### Contribuciones

| Tipo | Propósito |
|------|-----------|
| `urlSchemes` | Gestionar esquemas personalizados |
| `panels` | Ranuras de barra lateral u otros paneles |
| `commands` | Entradas de paleta de comandos y atajos |
| `themes` | Archivos JSON de tema adicionales |
| `settings` | Subpáginas de Settings |
| `devtools` | Pestañas de DevTools |
| `renderers` | Renderizadores personalizados para tipos MIME o extensiones |

## Permisos

| Permiso | Permite |
|---------|---------|
| `storage.plugin` | Almacenamiento clave-valor privado del plugin |
| `navigation.read` | Leer URL actual e información de pestaña |
| `navigation.write` | Disparar navegación |
| `network.fetch` | Fetch a través de APIs de red permitidas |
| `events.emit` | Emitir eventos del host |
| `events.subscribe` | Escuchar eventos del host |
| `devtools.network` | Detalle de red extra en DevTools |
| `render.unsanitized` | Omitir parte de la sanitización HTML (peligroso) |

El host aplica los permisos en tiempo de ejecución. Los permisos que desactives al instalar no se conceden a `ctx.network.fetch` ni a WASM `http_fetch`.

## Firmas del publicador

Las extensiones pueden incluir una firma Ed25519 en `renbrowser.plugin.rsg` (compatible con `rnid` de Reticulum). Las firmas no válidas bloquean la instalación.

Insignias en la vista previa y la lista:

| Insignia | Significado |
|----------|-------------|
| Sin firmar | Sin archivo de firma |
| Firmada | Identidad Reticulum válida |
| De confianza | Publicador de la lista de confianza |
| Alterada | Archivos modificados fuera de RenBrowser (desactivada hasta volver a activar) |

Durante la instalación puedes elegir **Confiar en esta identidad del publicador**. La lista de usuario está en `~/.renbrowser/trusted_publishers.json` y está protegida por un digest en la base de datos del perfil.

Firma con `build/scripts/sign-extension.sh` (requiere Python `rnid`).

## Traducciones de UI del plugin

Las extensiones pueden incluir cadenas en `locales/<code>.json`. Los títulos del manifiesto pueden usar `%clave.ruta%`; el host carga catálogos desde `/_plugins/<id>/locales/<code>.json`.

La vista previa de instalación lista los códigos de locale incluidos.

## Script de entrada del frontend

Exportes típicos en `main.js`: `activate(ctx)`, `deactivate()`, `mount(el)`, `handleScheme(url)`.

Con `network.fetch` concedido: `ctx.network.fetch()`. Comprueba `ctx.capabilities.networkFetch` antes de trabajo que use red.

Con backend WASM: `ctx.wasm.call("export", input)`. Cadenas con `ctx.i18n.t("key")`.

## Módulos WASM empaquetados

Un archivo `.wasm` puede llevar manifiesto (`renbrowser.plugin`), archivos (`renbrowser.files`) y firma opcional (`renbrowser.signature`).

Instala desde **Settings → Extensions → Choose .wasm module**.

`extensions/micron-translator/` usa TinyGo (`build-wasm.sh`, `go run ./extensions/micron-translator/bundle`).

## Backend WASM

`backend` apunta a un módulo WASM. El host ofrece `renhost.http_fetch` cuando `network.fetch` fue concedido en la instalación. Hay límites de peticiones, timeouts y tamaño.

## DevTools

En **Developer tools → Network**, las peticiones HTTP salientes de extensiones aparecen como **Extension fetch** con código de estado y duración.

## Integridad

Tras instalar, RenBrowser guarda un hash de los archivos del paquete. Cambios en disco fuera de la app desactivan la extensión (**Alterada**). Volver a activar acepta el estado actual.

## Notas de seguridad

- Instala solo de fuentes de confianza
- Lee permisos y endpoints antes de confirmar
- Prefiere extensiones firmadas de publicadores conocidos
- Trata los plugins como programas locales con acceso a tu perfil

## Próximos pasos

- Referencia: `internal/plugins/manifest.go`
- [Seguridad](security.md)
- [Desarrollo](development.md)
