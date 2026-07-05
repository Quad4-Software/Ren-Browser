# Entwicklung

So arbeiten Sie lokal an Ren Browser.

## Voraussetzungen

- Go 1.26+
- Node.js 22+ und pnpm 11+
- Task (empfohlen)
- Reticulum-Konfiguration für Live-Mesh-Tests (optional)

## Klonen und starten

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task dev
```

`task dev` startet Wails im Entwicklungsmodus mit Vite Hot Reload standardmäßig auf Port 9245.

## Häufige Tasks

| Task | Zweck |
|------|-------|
| `task dev` | Desktop-Entwicklungsmodus |
| `task build` | Produktions-Build für aktuelles OS |
| `task run` | Gebaute Binärdatei ausführen |
| `task check` | Vollständiges Qualitäts-Gate (Go + Frontend) |
| `task test` | Alle Unit-Tests |
| `task test:interop` | Live-Reticulum-Tests (braucht Netzwerk) |
| `task build:server` | Headless-Server-Binärdatei |
| `task run:server` | Server lokal ausführen |
| `task package` | Plattform-Installer oder Bundle |

## Qualitäts-Gate

Vor dem Senden von Änderungen:

```sh
task check
```

`check` führt aus:

- `gofmt` auf Go-Quellen
- `go test ./...`
- gosec-Sicherheitsscan
- Marken-Konsistenzprüfung
- Frontend-Typecheck, Lint, Format-Check, knip, audit und vitest

Optionale härtere Go-Tests:

```sh
task test:go:race
task test:go:hard
task fuzz:go
```

## Nur Frontend

```sh
cd frontend
pnpm install
pnpm check
pnpm test
```

Bindings unter `frontend/bindings/` werden von Wails aus Go-Services generiert.

## Go-Layout

| Pfad | Rolle |
|------|-------|
| `main_desktop.go` | Desktop-Einstieg (Wails-Fenster) |
| `main_server.go` | Server-Einstieg (nur HTTP) |
| `internal/app/` | An die UI exponierter Browser-Service |
| `internal/rns/` | Reticulum-Stack-Wrapper |
| `internal/nomadnet/` | Seitenabruf über LXMF. NomadNet-Knoten werden über Ankündigungen erkannt, nicht über NomadNet-Bibliotheken |
| `internal/micron/` | Micron-Rendering |
| `internal/store/` | SQLite-Persistenz-API |
| `internal/plugins/` | Erweiterungs-Host |
| `frontend/` | Svelte-5-UI |

Siehe [Architektur](architecture.md) für eine vollständigere Karte.

## Plugin-Entwicklung

```sh
task test:plugins
```

Installieren Sie Ihr Dev-Plugin nach `~/.renbrowser/plugins/` und laden Sie die App neu.

## Interop-Tests

```sh
task test:interop
```

Braucht ein live Reticulum-Netzwerk und das `interop`-Build-Tag. Überspringen für reine UI-Arbeit.

## Asset-Probe

Setzen Sie `REN_BROWSER_ASSET_PROBE=1`, wenn Sie debuggen, welcher Asset-Loader (eingebettet vs. Festplatte) aktiv ist.

## SPDX-Header

Neue Go-Dateien sollten enthalten:

```go
// SPDX-License-Identifier: MIT
```

Passen Sie den Stil an benachbarte Dateien an.

## Nächste Schritte

- [Architektur](architecture.md)
- [Mitwirken](contributing.md)
- [Erweiterungen](extensions.md)
