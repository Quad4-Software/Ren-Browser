# Atajos de teclado

Atajos predeterminados en escritorio. Puedes cambiarlos en **Settings → Keybinds**.

`mod` significa **Ctrl** en Windows y Linux, **Cmd** en macOS.

## Atajos predeterminados

| Acción | Acorde | Descripción |
|--------|--------|-------------|
| Enfocar barra de direcciones | `mod+l` | Mover el foco del teclado al campo de URL |
| Recargar página | `mod+r` | Recargar la pestaña activa |
| Herramientas de desarrollo | `mod+shift+i` | Abrir o cerrar DevTools |
| Buscar en la página | `mod+f` | Abrir la barra de búsqueda en la página actual |
| Panel Discovery | `mod+shift+d` | Abrir la barra lateral Discovery |
| Panel Settings | `mod+,` | Abrir Settings |
| Nueva pestaña | `mod+t` | Abrir una pestaña en blanco |
| Nueva ventana | `mod+shift+n` | Abrir otra ventana (cuando esté soportado) |
| Cerrar pestaña | `mod+w` | Cerrar la pestaña activa |
| Pantalla completa | `f11` | Alternar pantalla completa |

## Grabar un atajo nuevo

1. Abre **Settings → Keybinds**
2. Haz clic en la acción que quieres cambiar
3. Pulsa la nueva combinación de teclas
4. Los conflictos con otra acción se muestran en la UI. Resuélvelos antes de guardar.

Mientras grabas, los demás atajos se pausan para que el grabador solo vea tu acorde nuevo.

## Sintaxis de acordes

Los atajos se guardan como acordes en minúsculas unidos por `+`:

- `mod` es Ctrl o Cmd
- `shift` es Shift
- `alt` es Alt
- El segmento final es el nombre de la tecla (`l`, `r`, `,`, `f11`, etc.)

Ejemplo: `mod+shift+d` es Ctrl+Shift+D en Linux.

## Atajos de extensiones

Las extensiones pueden aportar comandos con campos `keybind` opcionales en `renbrowser.plugin.json`. Las extensiones activadas fusionan sus comandos en la paleta de comandos del host y en el manejo de atajos.

## Android

Los teclados físicos en Android siguen las mismas reglas de acordes cuando están conectados. La UI táctil no requiere atajos.

## Próximos pasos

- [Settings](settings.md) para otras preferencias
- [Uso del navegador](using-the-browser.md) para la vista general de paneles
