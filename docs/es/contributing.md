# Contribuir

Ren Browser acepta parches enviados a través de Reticulum. Los pull requests de GitHub también pueden ser bienvenidos según la política del proyecto. Esta página sigue [CONTRIBUTING.md](../../CONTRIBUTING.md).

## Flujo de parches

1. Clona o haz fork del repositorio
2. Crea una rama y haz cambios enfocados
3. Ejecuta `task check` cuando toques código Go o frontend
4. Haz commit con un mensaje claro
5. Exporta parches con `git format-patch`
6. Envía el archivo `.patch` por LXMF

## Enviar parches por LXMF

Destino:

```
f489752fbef161c64d65e385a4e9fc74
```

Adjunta el parche usando Sideband, Meshchat, MeshChatX o cualquier cliente LXMF que admita archivos adjuntos. Incluye una descripción breve en el cuerpo del mensaje.

Ten paciencia. La revisión ocurre a tiempo de mesh.

## Comandos de exportación

```sh
# Commit más reciente
git format-patch -1

# Últimos N commits
git format-patch -N

# Todos los commits desde main
git format-patch main..HEAD
```

Cada commit se convierte en un archivo `.patch`.

## Directrices para parches

- Un cambio lógico por serie de parches cuando sea posible
- Prueba antes de enviar
- Sigue el estilo de código existente
- Mantén `// SPDX-License-Identifier: MIT` en archivos Go nuevos
- Declara el uso de IA en el cuerpo del mensaje (ver abajo)

## Licencia

Al enviar un parche aceptas que se licencia bajo la [MIT License](../../LICENSE). Confirmas que tienes derecho a enviar el trabajo.

## Política de IA generativa

Puedes usar herramientas de IA si:

- Tu configuración da al modelo suficiente contexto
- Tu proveedor no entrena con el código que pegas

Lee [Reticulum Zen](https://reticulum.network/manual/zen.html) y la [Reticulum License](https://reticulum.network/manual/license.html).

**Declara** qué herramientas usaste en el mensaje del parche. Si no usaste IA de forma significativa, dilo brevemente.

Se prefieren modelos locales u offline.

Aun así debes leer, entender y probar todo lo que envías. No se acepta salida masiva sin revisar.

## Problemas de seguridad

No envíes detalles de vulnerabilidades como parches casuales por LXMF sin coordinación. Usa el proceso en [Seguridad](security.md).

## Configuración de desarrollo

Consulta [Desarrollo](development.md) para `task dev`, `task check` y el diseño del repositorio.

## Próximos pasos

- [Desarrollo](development.md)
- [Preguntas frecuentes](faq.md)
