# Datos y perfiles

Ren Browser guarda marcadores, historial, pestañas, configuración y caché de discovery en disco. Esta página explica dónde viven esos datos y cómo funcionan los perfiles.

## Ubicaciones predeterminadas

| Elemento | Ruta |
|----------|------|
| Directorio de datos | `~/.renbrowser/` |
| Base de datos principal | `~/.renbrowser/renbrowser.db` |
| Plugins | `~/.renbrowser/plugins/<id>/` |
| Perfiles con nombre | `~/.renbrowser/profiles/<name>/renbrowser.db` |
| Estado heredado (migrado) | `~/.renbrowser/state.json` |

En el primer arranque tras una actualización, `state.json` se importa a SQLite automáticamente.

## Qué guarda SQLite

Las tablas y blobs habituales incluyen:

- Pestañas abiertas y estado de restauración de sesión
- Historial de navegación con marcas de tiempo
- Favoritos
- Preferencias del navegador y atajos de teclado
- Entradas de discovery en caché
- Selección de tema y datos de tema personalizado

La corrupción se detecta al abrir. La UI puede ofrecer restablecer la base de datos. El restablecimiento elimina pestañas, historial, favoritos y configuración locales.

## Los datos de Reticulum son independientes

Las claves de identidad y la configuración de interfaces viven en tu directorio de Reticulum (`~/.reticulum-go/` por defecto). Ren Browser lee esa ruta pero no mueve tu identidad de Reticulum a `~/.renbrowser/`.

## Perfiles con nombre

Inicia con `--profile NAME` o `REN_BROWSER_PROFILE=NAME` para usar:

```
~/.renbrowser/profiles/NAME/renbrowser.db
```

Usa perfiles cuando quieras historiales separados en una cuenta (trabajo frente a personal, o pruebas).

## Importación y exportación

Solo al arrancar:

- `--export-profile /path/to/backup.json` escribe los datos del perfil y sale
- `--import-profile /path/to/backup.json` fusiona o reemplaza desde el archivo

El entorno refleja: `REN_BROWSER_EXPORT_PROFILE`, `REN_BROWSER_IMPORT_PROFILE`.

Exporta antes de actualizaciones importantes o al mudarte a una máquina nueva.

## Almacenamiento en modo servidor

| Modo | Pestañas, historial, favoritos |
|------|--------------------------------|
| Servidor predeterminado | SQLite en el servidor en `~/.renbrowser/` del servidor |
| `--public-mode` | `localStorage` del navegador de cada cliente |

Elige el modo público cuando muchos usuarios comparten una instancia de servidor.

## Importación y exportación de temas

Los temas se pueden exportar como JSON desde Settings e importar en otra instalación. Los archivos de tema no son el perfil completo, solo tokens de apariencia.

## Datos de plugins

Las extensiones con permiso `storage.plugin` obtienen almacenamiento aislado identificado por el id del plugin. Desinstalar un plugin no siempre elimina su carpeta. Borra `~/.renbrowser/plugins/<id>/` manualmente si quieres una eliminación limpia.

## Android

Las compilaciones móviles usan el mismo diseño lógico bajo el sandbox de la aplicación. Las rutas difieren según las reglas del SO pero el esquema de la base de datos coincide con escritorio.

## Lista de verificación de copia de seguridad

1. Detén Ren Browser
2. Copia `~/.renbrowser/renbrowser.db` (o la ruta de tu perfil)
3. Copia `~/.reticulum-go/` si también quieres la identidad de la mesh
4. Copia `~/.renbrowser/plugins/` si usas extensiones

Restaura colocando los archivos de vuelta antes del siguiente arranque.

## Próximos pasos

- [Settings](settings.md) para rutas en la UI
- [Modo servidor](server-mode.md) para modo público
- [Solución de problemas](troubleshooting.md) si la base de datos no abre
