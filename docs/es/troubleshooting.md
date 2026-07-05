# Solución de problemas

Problemas habituales y qué probar primero.

## Reticulum no arranca

**Síntomas:** Línea de log `reticulum start: ...`, Discovery vacío, todas las páginas de la mesh fallan.

**Comprobaciones:**

1. La ruta de configuración en `about:` coincide con donde viven tus archivos
2. `REN_BROWSER_CONFIG` o `--config` apunta a un archivo válido
3. Las definiciones de interfaz son sintácticamente correctas
4. Las claves y rutas de almacenamiento son legibles por el usuario que ejecuta Ren Browser

Corrige la configuración de Reticulum fuera de la aplicación, luego recarga desde Settings o reinicia.

## Discovery está vacío

Consulta [Discovery](discovery.md) y [Configuración de Reticulum](reticulum-setup.md).

Lista breve:

- Espera tras conectar para los anuncios
- Confirma que hay pares en tus interfaces
- Comprueba reglas de firewall para puertos UDP o TCP que configuraste

## Tiempo de espera al cargar página

1. Verifica el hash en la barra de direcciones
2. Abre otro nodo desde Discovery
3. Confirma que Reticulum muestra tráfico en las interfaces en Settings
4. Reintenta después de cambios de radio o ruta en redes mesh

## Base de datos corrupta o no abre

**Síntomas:** Error sobre datos del perfil, oferta de restablecer la base de datos.

**Opciones:**

1. Restaura `renbrowser.db` desde copia de seguridad ([Datos y perfiles](data-and-profiles.md))
2. Restablece desde la UI (destruye pestañas, historial, favoritos y configuración locales)
3. Renombra el archivo dañado y deja que Ren Browser cree una base de datos nueva

La identidad de Reticulum no se ve afectada por un restablecimiento de la base del navegador.

## Error de WASM o del analizador Micron

Si la comprobación SRI falla para el WASM de Micron:

1. No desactives la comprobación
2. Reinstala desde releases oficiales
3. Si compilaste desde el código fuente, ejecuta `task build` de nuevo sin editar a mano `frontend/dist/vendor/`

## Modo servidor: página en blanco o activos incorrectos

1. Comprueba que `--base-path` coincide con el montaje de tu proxy inverso
2. Activa `--trust-proxy` cuando TLS termina aguas arriba
3. Confirma el mapeo de puertos en Docker (`-p 8080:8080`)

## Modo servidor: historial compartido cuando no lo querías

Inicia con `--public-mode` para que cada navegador guarde su propia copia en `localStorage`.

## La extensión no carga

1. El manifiesto debe ser JSON válido en `renbrowser.plugin.json`
2. El `id` debe coincidir con el nombre de carpeta bajo `plugins/`
3. `engines.renbrowser` debe cumplirse con la versión de tu aplicación
4. Cadenas de permiso desconocidas provocan fallo de carga

Comprueba Settings para el mensaje de error.

## Falla la compilación para Android

1. Define `ANDROID_HOME`
2. Ejecuta `task android:install:deps`
3. Usa API 34 y NDK r26+ como se documenta en [Instalación](installation.md)

## Desarrollo: `task check` falla

| Área | Comando |
|------|---------|
| Formato Go | `task fmt:go` |
| Pruebas Go | `task test:go` |
| Frontend | `task frontend:check` |
| Escaneo de seguridad | `task gosec` |

Ejecuta `task check` antes de enviar parches.

## Sigues atascado

1. Anota tu versión desde `about:`
2. Captura logs de la terminal o de Docker
3. Pregunta en tu comunidad de la mesh o envía un informe de error detallado por los canales del proyecto

Consulta [Contribuir](contributing.md) para el envío de parches por LXMF.
