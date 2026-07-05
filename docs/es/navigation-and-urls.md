# Navegación y URLs

Ren Browser acepta varias formas de URL en la barra de direcciones. Esta página las enumera y explica cómo funciona la normalización.

## Destinos NomadNet

Una URL NomadNet completa tiene esta forma:

```
<32-hex-chars>:/page/path.mu
```

Ejemplo:

```
a1b2c3d4e5f6789012345678abcdef01:/page/index.mu
```

El hash es una identidad de destino de Reticulum en hexadecimal. La ruta después de `:/` es el archivo en ese nodo NomadNet.

## Hash abreviado

Si introduces solo una cadena hexadecimal de 32 caracteres, Ren Browser la expande a:

```
<hash>:/page/index.mu
```

Esto coincide con la ruta habitual de la página de inicio de NomadNet.

## Rutas sin hash

Si el contexto actual ya tiene un destino, una ruta que empiece por `/page/` puede resolverse de forma relativa a ese nodo. Para una navegación en frío, prefiere la forma completa `hash:/page/...`.

## Esquemas integrados

Estos esquemas se gestionan dentro de la aplicación. No usan la mesh.

| Esquema | Alias | Descripción |
|---------|-------|-------------|
| `about:` | `about` | Versión, compilación, ruta de configuración de Reticulum, directorio de datos |
| `license:` | `license` | Licencia del proyecto (MIT) |
| `editor:` | `editor` | Editor de código fuente Micron |

La coincidencia no distingue mayúsculas y minúsculas. Los espacios finales se recortan.

## Esquemas de URL de extensiones

Las extensiones instaladas pueden registrar esquemas personalizados en `renbrowser.plugin.json`. Por ejemplo la extensión hello registra `hello:`. Consulta [Extensiones](extensions.md).

## Settings y rutas internas

La UI puede usar rutas internas para paneles. La barra de direcciones se centra en esquemas de la mesh e integrados. Abre **Settings** con el botón de la barra lateral o `Ctrl+,` / `Cmd+,`.

## Títulos de pestaña

Los títulos de pestaña provienen de:

1. Metadatos de la página cuando el nodo proporciona un título
2. Nombres mostrados en Discovery cuando el hash coincide con un nodo conocido
3. Un hash o ruta acortados como respaldo

## Entradas del historial

Cada navegación que carga contenido puede crear una fila de historial con URL, título, hash de destino y marca de tiempo. Las páginas integradas como `about:` se incluyen.

## Resolución de enlaces en páginas Micron

Cuando haces clic en un enlace en una página renderizada:

- `about:`, `license:` y `editor:` se abren localmente
- Las URLs absolutas de la mesh navegan directamente
- Las rutas relativas se combinan con el destino de la página actual

## Reglas de normalización (resumen)

| Escribes | URL normalizada |
|----------|-----------------|
| `about` | `about:` |
| `license` | `license:` |
| `editor` | `editor:` |
| `abcdef...` (32 hex) | `abcdef...:/page/index.mu` |
| `hash:/page/foo.mu` completa | sin cambios |

## Próximos pasos

- [Uso del navegador](using-the-browser.md) para pestañas y paneles
- [Discovery](discovery.md) para elegir nodos sin escribir hashes
