# Desarrollo

Cómo trabajar en Ren Browser de forma local.

## Requisitos

- Go 1.26+
- Node.js 22+ y pnpm 11+
- Task (recomendado)
- Configuración de Reticulum para pruebas en mesh en vivo (opcional)

## Clonar y ejecutar

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task dev
```

`task dev` ejecuta Wails en modo desarrollo con recarga en caliente de Vite en el puerto 9245 por defecto.

## Tareas habituales

| Tarea | Propósito |
|-------|-----------|
| `task dev` | Modo desarrollo de escritorio |
| `task build` | Compilación de producción para el SO actual |
| `task run` | Ejecutar el binario compilado |
| `task check` | Compuerta de calidad completa (Go + frontend) |
| `task test` | Todas las pruebas unitarias |
| `task test:interop` | Pruebas Reticulum en vivo (necesita red) |
| `task build:server` | Binario de servidor sin interfaz gráfica |
| `task run:server` | Ejecutar servidor localmente |
| `task package` | Instalador o paquete de plataforma |

## Compuerta de calidad

Antes de enviar cambios:

```sh
task check
```

`check` ejecuta:

- `gofmt` en fuentes Go
- `go test ./...`
- Escaneo de seguridad gosec
- Comprobación de consistencia de marca
- Typecheck, lint, comprobación de formato, knip, audit y vitest del frontend

Pruebas Go opcionales más exigentes:

```sh
task test:go:race
task test:go:hard
task fuzz:go
```

## Solo frontend

```sh
cd frontend
pnpm install
pnpm check
pnpm test
```

Los bindings bajo `frontend/bindings/` los genera Wails desde servicios Go.

## Estructura Go

| Ruta | Rol |
|------|-----|
| `main_desktop.go` | Entrada de escritorio (ventana Wails) |
| `main_server.go` | Entrada de servidor (solo HTTP) |
| `internal/app/` | Servicio del navegador expuesto a la UI |
| `internal/rns/` | Envoltorio del stack Reticulum |
| `internal/nomadnet/` | Carga de páginas por LXMF. Los nodos NomadNet se detectan por anuncios, no por bibliotecas NomadNet |
| `internal/micron/` | Renderizado Micron |
| `internal/store/` | API de persistencia SQLite |
| `internal/plugins/` | Host de extensiones |
| `frontend/` | UI Svelte 5 |

Consulta [Arquitectura](architecture.md) para un mapa más completo.

## Desarrollo de plugins

```sh
task test:plugins
```

Instala tu plugin de desarrollo en `~/.renbrowser/plugins/` y recarga la aplicación.

## Pruebas de interoperabilidad

```sh
task test:interop
```

Requiere una red Reticulum en vivo y la etiqueta de compilación `interop`. Omítelas para trabajo solo de UI.

## Sonda de activos

Define `REN_BROWSER_ASSET_PROBE=1` al depurar qué cargador de activos (embebido frente a disco) está activo.

## Cabeceras SPDX

Los archivos Go nuevos deben incluir:

```go
// SPDX-License-Identifier: MIT
```

Sigue el estilo existente en archivos vecinos.

## Próximos pasos

- [Arquitectura](architecture.md)
- [Contribuir](contributing.md)
- [Extensiones](extensions.md)
