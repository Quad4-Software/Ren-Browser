# Architektur

Überblick über den Aufbau von Ren Browser.

## Überblick

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 5 frontend (Wails webview or browser in server) │
│  Tabs, chrome, Micron viewer, panels, settings          │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails bindings / HTTP API
┌───────────────────────▼─────────────────────────────────┐
│  internal/app : BrowserService                          │
│  Navigation, tabs, history, settings, plugin bridge     │
└───────┬─────────────┬─────────────┬─────────────────────┘
        │             │             │
        ▼             ▼             ▼
   internal/rns   internal/store  internal/plugins
   Reticulum      SQLite          Extensions
        │
        ▼
   internal/nomadnet : LXMF-Seitenabruf für als NomadNet angekündigte Knoten
        │
        ▼
   internal/micron : Markup parse and HTML render
```

## Einstiegspunkte

| Datei | Build-Tag | Rolle |
|-------|-----------|-------|
| `main_desktop.go` | `!server && !android` | Wails-Fenster, eingebettetes `frontend/dist` |
| `main_server.go` | `server` | HTTP-Server, dieselben eingebetteten Assets |
| Android main | `android` | Mobile Shell |

`internal/bootstrap` verdrahtet Konfiguration, Store, Plugins und Wails-App zusammen.

## Frontend

- **Framework:** Svelte 5 mit Vite
- **Bindings:** Generiert unter `frontend/bindings/renbrowser/`
- **Haupt-UI:** `frontend/src/App.svelte` orchestriert Chrome und Panels
- **Komponenten:** `frontend/src/lib/components/` (Tab-Leiste, Discovery, Einstellungen usw.)
- **Browser-Logik:** `frontend/src/lib/browser/` (URLs, Tastenkürzel, Fehler)

Micron-Rendering kann WASM-Parser nutzen, verwaltet von `MicronWasmManager` mit SRI-Verifikation.

## Backend-Services

### BrowserService (`internal/app`)

Zentrale API für die UI:

- URLs navigieren und Tab-Zustand verwalten
- Discovery, Verlauf, Favoriten bereitstellen
- Voreinstellungen laden und speichern
- Brücke zum Plugin-Host

### Reticulum-Stack (`internal/rns`)

Umschließt `quad4/reticulum-go`:

- Transport starten und stoppen
- Interface-Statistiken melden
- Konfiguration aus Einstellungen per Hot Reload neu laden

### Seitenabruf (`internal/nomadnet`)

Lädt entfernte `.mu`- und verwandte Inhalte über LXMF und Reticulum. Discovery kennzeichnet Knoten als NomadNet, wenn die Ankündigung passt. Ren Browser nutzt keine NomadNet-Client-Bibliotheken.

### Store (`internal/store` + `internal/db`)

SQLite-Persistenz mit Migration vom Legacy `state.json`.

### Plugins (`internal/plugins`)

- Manifest-Validierung
- Berechtigungserzwingung
- Eingebaute Schemes (`about:`, `license:`, `editor:`)
- JS- und WASM-Plugin-Laufzeiten

## Inhalt und Rendering

| Paket | Rolle |
|-------|-------|
| `internal/content` | Statische Seiten (about, license) |
| `internal/micron` | Micron zu HTML |
| `internal/micronwasm` | WASM-Parser-Integration |
| `internal/cache` | Seiten-Cache-Helfer |

## Server-Middleware

`internal/servermw` behandelt Base-Path-Header und proxy-bewusstes URL-Bauen im Servermodus.

## Konfiguration

`internal/config` parst Flags, `.env` und `REN_BROWSER_*`-Variablen in eine `Runtime`-Struktur, die bootstrap nutzt.

## Marke und Pfade

`internal/brand` (generiert aus `build/brand.yml`) definiert stabile Namen:

- Datenverzeichnis `.renbrowser`
- DB-Datei `renbrowser.db`
- Anzeigename und Versionslabels

## Build und Packaging

- `Taskfile.yml`: Entwicklerbefehle
- `build/`: plattformspezifisches Packaging (Linux AppImage, Windows NSIS, macOS, Android, Docker)
- `build/config.yml`: Wails-Projektkonfiguration

## CI

GitHub Actions führt Go-Tests, Frontend-Checks, Sicherheitsscans, Desktop- und Server-Smoke-Builds und Release-Artefakte aus. Siehe `.github/workflows/`.

## Erweiterungspunkte

1. **Plugins**: manifestgesteuerte UI und Schemes
2. **Themes**: JSON-Token-Dateien
3. **Community-Interfaces**: Reticulum-Konfigurations-Snippets in Einstellungen

## Nächste Schritte

- [Entwicklung](development.md) zum lokalen Bauen
- [Erweiterungen](extensions.md) für Plugin-API-Oberfläche
- Quellbaum in der Layout-Tabelle der `README.md` im Repository
