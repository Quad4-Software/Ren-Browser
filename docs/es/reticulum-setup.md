# Configuración de Reticulum

Ren Browser usa [Reticulum](https://reticulum.network/) a través del stack `quad4/reticulum-go`. Esta página explica qué espera la aplicación y cómo corregir problemas habituales de la mesh.

## Ubicación predeterminada de la configuración

| Elemento | Ruta predeterminada |
|----------|---------------------|
| Directorio de configuración de Reticulum | `~/.reticulum-go/` |
| Flag de sobrescritura | `--config /path/to/config` |
| Variable de entorno de sobrescritura | `REN_BROWSER_CONFIG` o `RETICULUM_CONFIG` |

Los archivos exactos dentro del directorio dependen de tu configuración de Reticulum o reticulum-go. Ren Browser inicia el stack al arrancar y recarga los cambios de interfaz desde **Settings**.

## Qué ocurre al iniciar

1. Ren Browser carga tu configuración de Reticulum
2. Las interfaces entran en línea (UDP, TCP, RNode y otras que hayas configurado)
3. Los anuncios de nodos NomadNet aparecen en **Discovery**
4. Las peticiones de página salen por LXMF y Reticulum hacia nodos que alojan páginas Micron

Si el arranque falla, revisa el log de la terminal (escritorio) o los logs del contenedor (servidor). La aplicación sigue ejecutándose para que puedas abrir `about:` y **Settings**.

## Interfaces en Settings

Abre **Settings** y busca la sección de Reticulum. Puedes:

- Ver qué interfaces están activas
- Ver estadísticas de transmisión y recepción
- Editar la configuración y aplicar recarga en caliente sin reiniciar toda la aplicación

Úsalo cuando añadas una interfaz nueva o cambies claves y quieras que el navegador recoja los cambios rápido.

## Unirse a la mesh

Necesitas al menos una ruta hacia otros nodos de Reticulum. Opciones habituales:

- **UDP o TCP local** en una LAN con otros pares de Reticulum
- **RNode** u otro hardware de radio similar
- **Definiciones de interfaz** que apunten a pares o hubs conocidos

Reticulum queda fuera del alcance de este manual. Lee el [manual de Reticulum](https://reticulum.network/manual/) para la sintaxis de interfaces y la gestión de identidades.

## Destinos NomadNet

Las páginas de NomadNet viven en destinos de Reticulum. En la barra de direcciones puedes usar:

- Una ruta completa como `abcdef0123456789abcdef0123456789:/page/index.mu`
- Un hash hexadecimal de 32 caracteres (Ren Browser añade `:/page/index.mu`)

Las páginas usan el formato de marcado Micron. Ren Browser las renderiza con el pipeline Micron integrado.

## Cuando Discovery está vacío

Recorre esta lista:

1. Confirma que Reticulum se ejecuta dentro de Ren Browser (Settings muestra las interfaces)
2. Comprueba que tus interfaces coinciden con cómo están configurados los pares en la mesh
3. Espera un momento tras conectar. Los anuncios no son instantáneos
4. Verifica que estás en la misma red lógica que los nodos que esperas ver

## Cuando las páginas agotan el tiempo o fallan

1. Confirma que el hash de destino es correcto
2. Comprueba que tienes ruta hacia ese destino (no solo visibilidad en Discovery)
3. Prueba otro nodo conocido desde Discovery
4. Revisa devtools o logs en busca de errores LXMF o de transporte

## Servidor y Docker

Cuando ejecutas `renbrowser-server` en Docker, monta la configuración del host en solo lectura:

```sh
-v "$HOME/.reticulum-go:/root/.reticulum-go:ro"
```

El usuario del contenedor debe poder leer claves y definiciones de interfaz dentro de ese directorio.

## Próximos pasos

- [Discovery](discovery.md) para explorar nodos anunciados
- [Navegación y URLs](navigation-and-urls.md) para formatos de la barra de direcciones
- [Solución de problemas](troubleshooting.md) para mensajes de error
