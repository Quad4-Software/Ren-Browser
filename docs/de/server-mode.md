# Servermodus

Der Servermodus betreibt Ren Browser als Web-App ohne Desktop-Shell. Sie greifen über eine HTTP-URL aus einem anderen Browser zu.

## Wann Servermodus sinnvoll ist

- Homelab oder VPS, der Reticulum bereits betreibt
- Docker-Deployments
- Gemeinsamer Rechner, auf dem Sie die Desktop-App nicht installieren möchten
- Zugriff von Tablets oder Smartphones im LAN

## Schnellstart

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Öffnen Sie `http://localhost:8080` (oder Ihre Host-IP) in Firefox, Chromium oder Safari.

## Docker

Veröffentlichtes Image:

```
ghcr.io/quad4-software/renbrowser:latest
```

Beispielausführung:

```sh
mkdir -p "$HOME/.reticulum-go" "$HOME/.renbrowser"
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

Binden Sie das Reticulum-Verzeichnis nicht schreibgeschützt ein. Details und Podman-Hinweise: [Reticulum-Einrichtung](reticulum-setup.md#server-und-docker).

Lokal bauen:

```sh
task build:docker
task run:docker
```

## Kommandozeilen-Flags

| Flag | Zweck |
|------|-------|
| `--host` | Bind-Adresse (Standard `0.0.0.0` im Server-Build) |
| `--port` | HTTP-Port (Standard `8080`) |
| `--config` | Reticulum-Konfigurationspfad |
| `--trust-proxy` | `X-Forwarded-*` von einem Reverse Proxy vertrauen |
| `--base-path` | URL-Präfix bei Bereitstellung unter einem Unterpfad |
| `--public-mode` | Favoriten, Verlauf und Tabs im Browser-`localStorage` statt serverseitiger SQLite speichern |
| `--profile` | Benannte Profildatenbank |
| `--import-profile` / `--export-profile` | Profil-JSON beim Start |

## Umgebungsvariablen

Der Server liest eine `.env`-Datei im Arbeitsverzeichnis. Bereits in der Umgebung gesetzte Variablen werden nicht überschrieben.

| Variable | Zweck |
|----------|-------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | Bind-Adresse |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | Port |
| `REN_BROWSER_CONFIG` / `RETICULUM_CONFIG` | Reticulum-Konfiguration |
| `REN_BROWSER_TRUST_PROXY` | `true` / `1` / `yes` zum Aktivieren von Trust Proxy |
| `REN_BROWSER_BASE_PATH` | Unterpfad-Präfix |
| `REN_BROWSER_PUBLIC_MODE` | Public-Mode-Schalter |
| `REN_BROWSER_PROFILE` | Profilname |
| `REN_BROWSER_IMPORT_PROFILE` | Importpfad beim Start |
| `REN_BROWSER_EXPORT_PROFILE` | Exportpfad beim Start |

## Public Mode

Ohne `--public-mode` hält der Server Tabs, Verlauf und Favoriten in seiner SQLite-Datenbank auf der Server-Festplatte. Jeder Client, der diese Instanz teilt, sieht dieselben Daten.

Mit `--public-mode` liegen diese Einträge im `localStorage` jedes Browsers. Nutzen Sie das, wenn viele Personen einen Server treffen und kein gemeinsames Profil haben sollen.

## Reverse Proxy

Typisches nginx- oder Caddy-Setup:

1. TLS am Proxy beenden
2. An `127.0.0.1:8080` weiterleiten
3. `X-Forwarded-Proto` und `X-Forwarded-Host` durchreichen
4. Ren Browser mit `--trust-proxy` starten
5. `--base-path` setzen, wenn die App nicht an der Domain-Wurzel liegt

Der Header `X-RenBrowser-Base-Path` wird erkannt, wenn Trust Proxy aktiv ist.

## Keine eingebaute Authentifizierung

Jeder, der den HTTP-Port erreichen kann, kann den Browser nutzen und Reticulum-Traffic auslösen. Stellen Sie Port 8080 nicht dem öffentlichen Internet ohne Folgendes bereit:

- Firewall-Regeln
- VPN
- Reverse Proxy mit Authentifizierung
- Oder alles davon

Lesen Sie [Sicherheit](security.md), bevor Sie einen Server veröffentlichen.

## Asset-Overrides (fortgeschritten)

Für die Entwicklung können Sie Frontend-Dateien von Festplatte oder Zip statt eingebetteter Assets ausliefern:

- `--assets-dir path`
- `--assets-zip path`

Umgebung: `REN_BROWSER_ASSETS_DIR`, `REN_BROWSER_ASSETS_ZIP`.

## Nächste Schritte

- [Daten und Profile](data-and-profiles.md) für SQLite vs. Public Mode
- [Sicherheit](security.md) für Deployment-Härtung
- [Installation](installation.md) für Release-Binärdateien
