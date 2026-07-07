# Seguridad

Ren Browser está pensado para sistemas y redes en las que confías. Esta página resume el uso seguro, la verificación de descargas y cómo informar de problemas.

## Límites de confianza

| Superficie | Riesgo |
|------------|--------|
| Aplicación de escritorio | Webview local con enlaces Go. Sin Node.js en el contenido de página. |
| Modo servidor | Puerto HTTP abierto. Sin inicio de sesión incluido. |
| Plugins | Se ejecutan con permisos declarados. Los plugins WASM están en sandbox con límites. |
| Contenido de la mesh | No confiable como cualquier contenido de red. El HTML Micron se sanitiza salvo que un plugin solicite `render.unsanitized`. |

## Modo servidor

`renbrowser-server` **no tiene autenticación**. Si lo expones en internet arriesgas:

- Escaneo automatizado
- Abuso de tus interfaces de Reticulum
- Sobrecarga de una aplicación de un solo proceso

Si debes exponerlo:

1. Colócalo detrás de un proxy inverso con controles de acceso
2. Usa HTTPS con un certificado válido
3. Restringe el puerto con reglas de firewall o VPN
4. Mantén la aplicación actualizada

Consulta [Modo servidor](server-mode.md) para flags de proxy.

## Plugins de escritorio

Instala extensiones solo de personas o proyectos en los que confíes. Lee la lista de permisos en **Settings → Extensions** antes de activar.

### Comprobaciones al instalar

Al instalar, RenBrowser muestra:

- Interruptores por permiso (los desactivados no se conceden en runtime)
- Endpoints de red escaneados del manifiesto y archivos del paquete
- Insignias de firma (sin firmar, firmada, publicador de confianza, alterada)
- Evaluación heurística de seguridad

Las firmas no válidas bloquean la instalación. Las extensiones sin firmar pueden instalarse si aceptas el riesgo.

### Protección en runtime

- Permisos concedidos aplicados a JS `PluginFetch` y WASM `http_fetch`
- Exportes WASM de red bloqueados sin `network.fetch`
- Límites de peticiones HTTP y trabajo WASM para reducir bloqueos
- Hash de integridad de archivos; manipulación externa desactiva la extensión
- Lista de publicadores de confianza protegida por digest en la base de datos

El tráfico HTTP de plugins aparece en **Developer tools → Network**. Consulta [Extensiones](extensions.md).

## Verificar descargas

Las compilaciones oficiales vienen de [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) y de GitHub Actions CI.

Cada release debería incluir `SHA256SUMS.txt`. Comprueba tu archivo:

```sh
sha256sum -c SHA256SUMS.txt
```

Para Docker, prefiere fijar por digest (`@sha256:...`) después de confiar en una compilación. Las imágenes en GHCR incluyen procedencia de compilación y un SBOM de Docker Buildx.

Si un binario no coincide con las sumas publicadas, trátalo como no confiable.

## Integridad de subrecursos para WASM

El WebAssembly del analizador Micron y su compañero `wasm_exec.js` se comprueban con SRI SHA-384 antes de ejecutarse. Una discrepancia de hash bloquea el código y muestra un error.

## Datos en reposo

- Estado de la aplicación: SQLite bajo `~/.renbrowser/`
- Claves de Reticulum: tu directorio de configuración de Reticulum
- Modo servidor público: parte de los datos solo en `localStorage` del navegador de cada cliente

Cifra los discos a nivel de SO si la máquina es compartida o portátil.

## Informar vulnerabilidades

**No** abras un issue público en GitHub para fallos de seguridad sin corregir.

**Contacto preferido:**

1. LXMF: `f489752fbef161c64d65e385a4e9fc74`

Incluye versión, plataforma, pasos para reproducir e impacto.

Las preguntas legales y de licencias van a [LEGAL.md](../../LEGAL.md) (`legal@quad4.io`), no al canal de seguridad.

## CI y cadena de suministro (resumen)

GitHub Actions ejecuta pruebas, gosec, escaneos Trivy y CodeQL según programación. Las Actions de terceros están fijadas a SHAs de commit en los workflows.

## Próximos pasos

- Lista de permisos en [Extensiones](extensions.md)
- Despliegue en [Modo servidor](server-mode.md)
- [SECURITY.md](../../SECURITY.md) en la raíz del repositorio para la política canónica
