# Settings

Settings controla cómo se ve Ren Browser, cómo se conecta a Reticulum y cómo se comporta en tu máquina.

Abre Settings desde la barra lateral o con `Ctrl+,` / `Cmd+,` (personalizable en la configuración de atajos).

## Apariencia

### Temas

Ren Browser incluye temas oscuro y claro. Puedes:

- Cambiar de tema en Settings
- Importar un archivo JSON de tema
- Exportar tu tema actual para copia de seguridad o compartir

Los temas afectan los colores del chrome, los tokens tipográficos y el estilo de páginas Micron cuando corresponde.

### Chrome de la ventana

En escritorio puedes elegir entre una barra de título nativa y una ventana sin marco con controles personalizados de minimizar, maximizar y cerrar. El modo sin marco usa regiones arrastrables en la parte superior de la ventana.

## Reticulum

La sección de Reticulum muestra:

- Interfaces activas y su estado
- Contadores de bytes transmitidos y recibidos
- Editor de configuración con recarga en caliente

Después de editar la configuración, aplica los cambios desde Settings en lugar de reiniciar la aplicación cuando sea posible.

## Atajos de teclado

Cada acción puede tener un acorde como `mod+t` para nueva pestaña. `mod` significa Ctrl en Windows y Linux, Cmd en macOS.

Consulta [Atajos de teclado](keyboard-shortcuts.md) para los valores predeterminados y cómo grabar combinaciones nuevas.

## Extensiones

Gestiona plugins desde **Settings → Extensions**:

- Instalar desde zip, carpeta o módulo `.wasm` empaquetado
- Vista previa: permisos, endpoints de red, insignias de firma, notas de seguridad, idiomas de UI
- Activar o desactivar extensiones instaladas
- Estado de firma y avisos de alteración en cada tarjeta

Consulta [Extensiones](extensions.md) para manifiesto, firma y rutas de instalación.

## Perfil y datos

Settings muestra las rutas de:

- Ubicación de la base de datos SQLite
- Ruta de configuración de Reticulum
- Directorio de plugins bajo `~/.renbrowser/plugins/`

Para perfiles con nombre e importación o exportación, consulta [Datos y perfiles](data-and-profiles.md).

## Preferencias del navegador

Los interruptores adicionales pueden incluir:

- Barra de título nativa frente a ventana sin marco
- Comportamiento predeterminado de paneles
- Opciones que se sincronizan con preferencias del navegador guardadas en SQLite

Los interruptores exactos pueden variar según la versión. En caso de duda, revisa la etiqueta en la UI y este documento para tu etiqueta de release.

## Diseño móvil

En Android, Settings usa los mismos datos pero puede agrupar elementos para navegación táctil. Las opciones principales de Reticulum y tema siguen disponibles.

## Próximos pasos

- [Atajos de teclado](keyboard-shortcuts.md)
- [Extensiones](extensions.md)
- [Datos y perfiles](data-and-profiles.md)
