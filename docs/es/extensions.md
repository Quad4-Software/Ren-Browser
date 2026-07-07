# Extensiones

Ren Browser admite plugins que añaden esquemas de URL, paneles laterales, comandos, temas, páginas de configuración y pestañas de devtools.

## Instalar extensiones

### Desde Settings

1. Abre **Settings → Extensions**
2. Elige **Install from zip** o **Install from folder**
3. Confirma que el manifiesto carga y que los permisos se ven correctos
4. Activa la extensión

### Instalación manual

Descomprime un plugin en:

```
~/.renbrowser/plugins/<id>/
```

La carpeta debe contener `renbrowser.plugin.json`. El `id` del manifiesto debe coincidir con el nombre de la carpeta.

## Extensión de ejemplo

El repositorio incluye `extensions/hello-extension/`:

- Registra el esquema de URL `hello:`
- Añade un panel lateral **Hello**
- Define un comando **Say hello** con `mod+shift+h`

Úsala como plantilla cuando escribas tu propio plugin.

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

Los campos opcionales incluyen `description`, `author`, `license`, `engines`, `backend` y `contributes`.

### Restricción de motor

```json
"engines": { "renbrowser": ">=0.1.0" }
```

El host rechaza cargar el plugin si la versión de tu aplicación es demasiado antigua.

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

Los plugins deben declarar lo que necesitan. Permisos conocidos:

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

El host aplica los permisos en tiempo de ejecución. Un plugin no puede usar una capacidad que no declaró.

## Script de entrada del frontend

Un `main.js` típico exporta:

- `activate(ctx)` para suscribirse a eventos y registrar UI
- `deactivate()` para limpieza
- `mount(el)` para renderizar HTML del panel lateral
- `handleScheme(url)` para manejadores de esquemas de URL

La extensión hello muestra versiones mínimas de cada una.

## Backend WASM

Los plugins pueden establecer `backend` en una ruta de módulo WASM para lógica más pesada. Los plugins WASM se ejecutan en un runtime restringido con permisos explícitos.

## Notas de seguridad

- Instala plugins solo de fuentes en las que confíes
- Lee la lista de permisos antes de activar
- Trata los plugins como cualquier programa local con acceso a los datos de tu perfil

## Próximos pasos

- Referencia en el código: `internal/plugins/manifest.go` en el repositorio
- [Seguridad](security.md) para el modelo de amenazas de plugins
- [Desarrollo](development.md) para trabajar en el host de plugins
