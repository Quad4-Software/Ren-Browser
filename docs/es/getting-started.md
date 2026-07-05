# Primeros pasos

Ren Browser te permite abrir páginas de NomadNet a través de Reticulum. Piensa en él como un navegador pequeño hecho para la mesh, no para la web pública.

## Qué hace Ren Browser

- Abre páginas `.mu` y otro contenido de NomadNet desde destinos de la mesh
- Muestra los nodos a los que puedes llegar a través de tus interfaces de Reticulum
- Guarda pestañas, historial, marcadores y configuración en tu máquina (o en el navegador cuando usas el modo servidor público)
- Renderiza marcado Micron con un analizador WASM y temas integrados
- Admite extensiones que añaden paneles, esquemas de URL y comandos

## Qué necesitas primero

Antes de que Ren Browser pueda cargar una página remota, necesitas una configuración de **Reticulum** que funcione en la misma máquina (o montada en un contenedor Docker para el modo servidor).

Ren Browser lee la configuración de Reticulum desde `~/.reticulum-go/` por defecto. Puedes apuntar a otro archivo con `--config` o la variable de entorno `REN_BROWSER_CONFIG`.

**No** necesitas una conexión tradicional a internet para navegar páginas de NomadNet en la mesh. Sí necesitas al menos una interfaz que pueda llegar a otros nodos de Reticulum.

## Escritorio o servidor

| Modo | Ideal para |
|------|------------|
| **Desktop** (predeterminado) | Uso diario en Linux, Windows o macOS. Ventana nativa, base de datos SQLite local. |
| **Server** | Homelab, Docker o una máquina que ya ejecuta Reticulum. Abres Ren Browser en otro navegador en `http://host:8080`. |
| **Android** | Compilaciones móviles cuando tu release incluye un APK. Las mismas funciones de navegación en un diseño táctil. |

## Lista de verificación del primer arranque

1. Instala Reticulum y crea o copia una configuración en `~/.reticulum-go/`
2. Instala Ren Browser desde [releases](https://github.com/Quad4-Software/Ren-Browser/releases) o compila desde el código fuente
3. Inicia Ren Browser y espera a que Reticulum se conecte en tus interfaces
4. Abre **Discovery** o escribe un hash de destino de 32 caracteres en la barra de direcciones
5. Visita `about:` para confirmar la versión, la ruta de configuración y el directorio de datos

## Páginas integradas (sin mesh)

Estas funcionan aunque estés desconectado de la mesh:

| Dirección | Propósito |
|-----------|-----------|
| `about:` | Versión de la app, información de compilación, rutas |
| `license:` | Texto de la licencia MIT |
| `editor:` | Editor Micron integrado |

Escribe `settings` en la barra de direcciones o pulsa el atajo de configuración para abrir las preferencias.

## Próximos pasos

- [Instalación](installation.md) si aún no has instalado
- [Configuración de Reticulum](reticulum-setup.md) si las páginas no cargan o Discovery sigue vacío
- [Uso del navegador](using-the-browser.md) para la navegación diaria
