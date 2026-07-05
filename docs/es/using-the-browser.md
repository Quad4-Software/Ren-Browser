# Uso del navegador

Esta página cubre el uso diario: pestañas, barra de direcciones, historial, favoritos y visualización de páginas.

## Diseño de la ventana

Las áreas principales son:

- **Barra de pestañas** arriba. Arrastra pestañas, fíjalas o abre vistas divididas.
- **Barra de direcciones** para destinos y esquemas integrados.
- **Área de contenido** donde se renderizan las páginas Micron.
- **Paneles laterales** para Discovery, History, Settings, DevTools y paneles de extensiones.

En pantallas más pequeñas una barra de navegación móvil sustituye parte del chrome de escritorio.

## Pestañas

- **Nueva pestaña**: atajo predeterminado `Ctrl+T` / `Cmd+T` (consulta [Atajos de teclado](keyboard-shortcuts.md))
- **Cerrar pestaña**: `Ctrl+W` / `Cmd+W`
- **Restaurar sesión**: las pestañas abiertas se guardan en la base de datos local y se restauran en el siguiente arranque

Las pestañas fijadas permanecen al frente de la tira de pestañas. La vista dividida te permite mostrar dos pestañas lado a lado.

## Barra de direcciones

Escribe un destino NomadNet o usa un esquema integrado:

| Entrada | Resultado |
|---------|-----------|
| Hash hex de 32 caracteres | Abre `hash:/page/index.mu` |
| URL NomadNet completa | Abre tal como la escribiste |
| `about:` | Página about con versión y rutas |
| `license:` | Texto de la licencia |
| `editor:` | Editor Micron |
| `settings` | Abre Settings (mediante enrutamiento de la UI) |

Pulsa Enter para navegar. La barra también acepta foco con `Ctrl+L` / `Cmd+L`.

## Seguir enlaces

Los enlaces Micron en una página se resuelven de forma relativa al destino actual. Los enlaces internos `about:`, `license:` y `editor:` funcionan como navegación normal.

Los enlaces externos de la mesh usan la sintaxis de destino de Reticulum. Si un enlace falla, comprueba que el nodo destino es alcanzable.

## Historial

Abre el panel History desde la barra lateral o su atajo de teclado. Puedes:

- Buscar por título o URL
- Ver entradas agrupadas por fecha
- Abrir una página pasada en la pestaña actual o en una nueva

El historial se guarda localmente en SQLite (escritorio) salvo que uses el modo servidor público.

## Favoritos

Guarda nodos que visitas a menudo desde el menú contextual de la página o desde Discovery. Los favoritos se sincronizan con el mismo almacén que historial y pestañas.

## Buscar en la página

Pulsa `Ctrl+F` / `Cmd+F` para buscar texto en la página Micron actual. Las coincidencias se resaltan en el visor de contenido.

## Herramientas de desarrollo

Pulsa `Ctrl+Shift+I` / `Cmd+Shift+I` para abrir DevTools. Útil para:

- Inspeccionar tiempos de renderizado
- Ver el código fuente crudo de la página
- Depurar paneles de extensiones que aportan entradas en devtools

## Descargas

Cuando una página o acción ofrece un archivo, los elementos aparecen en el menú de descargas. Las rutas siguen las convenciones de descarga de tu SO en escritorio.

## Errores de página

Si una página no puede cargar, verás un estado de error con un mensaje breve. Causas habituales:

- Destino inalcanzable
- Contenido Micron no válido
- Reticulum no conectado

Usa **Reload** (`Ctrl+R` / `Cmd+R`) después de corregir la conectividad.

## Editor Micron

Abre `editor:` para componer Micron en un editor integrado. Úsalo para borradores locales antes de publicar en un nodo NomadNet.

## Próximos pasos

- [Navegación y URLs](navigation-and-urls.md) para las reglas de URL en detalle
- [Discovery](discovery.md) para explorar la mesh
- [Settings](settings.md) para temas e interfaces
