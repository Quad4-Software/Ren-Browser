# Discovery

Discovery muestra los nodos NomadNet anunciados en tus interfaces de Reticulum. Úsalo cuando aún no conoces un hash de destino.

## Abrir Discovery

- Haz clic en **Discovery** en la barra lateral
- Usa el atajo de teclado (`Ctrl+Shift+D` / `Cmd+Shift+D` por defecto)

El panel lista nodos con nombres mostrados, hashes y metadatos cuando el anuncio los incluye.

## Qué aparece en la lista

Un nodo aparece cuando:

1. Reticulum recibe un anuncio para ese destino
2. El anuncio coincide con NomadNet o tipos de nodo compatibles que el navegador entiende
3. Tus interfaces pueden alcanzar la ruta que transportó el anuncio

Discovery es en vivo. La lista se actualiza cuando llegan anuncios nuevos y caducan los antiguos.

## Abrir un nodo

Haz clic en una fila para navegar a la página predeterminada de ese nodo (normalmente `index.mu`). También puedes copiar el hash para usarlo en la barra de direcciones.

## Favoritos desde Discovery

Muchas filas ofrecen una forma de añadir el nodo a favoritos. Los favoritos se guardan en la base de datos de tu perfil local.

## Lista vacía

Si Discovery sigue vacío, repasa [Configuración de Reticulum](reticulum-setup.md):

- Las interfaces deben estar activas en Settings
- Necesitas conectividad con pares que transporten anuncios
- Las uniones nuevas pueden tardar un momento en poblarse

## Anuncios frente a alcanzabilidad

Ver un nodo en Discovery no garantiza que cada carga de página tenga éxito. Sigues necesitando una ruta hacia el destino para peticiones de página LXMF por Reticulum.

Si Discovery muestra un nodo pero las páginas fallan, comprueba las rutas de transporte y si el nodo remoto está en línea.

## Interfaces comunitarias

Settings puede listar definiciones de interfaz comunitarias o compartidas cuando están habilitadas. Te ayudan a unirte a segmentos más amplios de la mesh. Aplica los cambios desde la sección de Reticulum en Settings.

## Próximos pasos

- [Navegación y URLs](navigation-and-urls.md) para introducir hashes manualmente
- [Uso del navegador](using-the-browser.md) para favoritos e historial
