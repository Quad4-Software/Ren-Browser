# Preguntas frecuentes

Respuestas breves a preguntas habituales.

## ¿Qué es Ren Browser?

Un navegador para páginas NomadNet en la mesh Reticulum. No es un navegador web de propósito general para internet público.

## ¿Necesito internet?

Necesitas conectividad Reticulum con otros nodos. Eso puede ser totalmente fuera de internet público (radio, LAN, etc.).

## ¿Dónde obtengo reticulum-go?

Instala y configura [reticulum-go](https://reticulum-go.quad4.io) en la máquina donde ejecutas Ren Browser. La app usa tu configuración de Reticulum (por ejemplo en `~/.reticulum-go/`) pero no crea tu identidad ni tus interfaces.

## ¿Qué es NomadNet?

[Nomad Network](https://github.com/markqvist/NomadNet) es un programa de comunicación mesh fuera de red, construido sobre LXMF y Reticulum. Los nodos conectables pueden alojar páginas y archivos, a menudo en el lenguaje de marcado Micron. Ren Browser no integra el cliente NomadNet. Detecta nodos que anuncian como NomadNet y abre sus páginas alojadas por Reticulum.

## ¿Qué es Micron?

Un lenguaje de marcado eficiente en ancho de banda usado en nodos Nomad Network. Ren Browser lo renderiza a HTML en el visor de contenido.

## ¿Puedo navegar sitios HTTPS normales?

No. Ren Browser apunta al contenido de la mesh Reticulum, incluidas páginas en nodos NomadNet, no a URLs web públicas arbitrarias.

## ¿Escritorio frente a servidor?

El escritorio ejecuta una ventana nativa y guarda datos en SQLite local. El modo servidor sirve la UI por HTTP para usarla en otro navegador. Consulta [Modo servidor](server-mode.md).

## ¿Es seguro el modo servidor en internet?

No por defecto. No hay inicio de sesión. Usa VPN, firewall o proxy inverso con autenticación. Consulta [Seguridad](security.md).

## ¿Dónde están mis datos?

`~/.renbrowser/renbrowser.db` en escritorio por defecto. Consulta [Datos y perfiles](data-and-profiles.md).

## ¿Cómo verifico una descarga de release?

Usa `SHA256SUMS.txt` de la página de release. Consulta [Seguridad](security.md).

## ¿Cómo instalo una extensión?

Settings → Extensions, o descomprime en `~/.renbrowser/plugins/<id>/`. Consulta [Extensiones](extensions.md).

## ¿Cómo escribo la dirección de un nodo?

Pega el hash hexadecimal de 32 caracteres o la URL completa `hash:/page/file.mu`. Consulta [Navegación y URLs](navigation-and-urls.md).

## Discovery no muestra nada

Comprueba las interfaces de Reticulum en Settings y [Configuración de Reticulum](reticulum-setup.md).

## ¿Cómo informo de un error?

Para problemas de seguridad usa LXMF según [Seguridad](security.md). Para correcciones de código consulta [Contribuir](contributing.md).

## ¿Qué licencia tiene el proyecto?

MIT. Escribe `license:` en la barra de direcciones o lee [LICENSE](../../LICENSE).

## ¿Con qué stack está construido?

Go, Wails v3, Svelte 5, SQLite y bibliotecas Reticulum de Quad4. Consulta [Arquitectura](architecture.md).

## ¿Puedo ejecutarlo en Android?

Sí cuando se publique un APK para tu release. Compila desde el código fuente con el Android SDK si hace falta. Consulta [Instalación](installation.md).

## ¿Cómo cambio los atajos de teclado?

Settings → Keybinds. Consulta [Atajos de teclado](keyboard-shortcuts.md).
