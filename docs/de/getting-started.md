# Erste Schritte

Mit Ren Browser öffnen Sie NomadNet-Seiten über Reticulum. Stellen Sie sich einen kleinen Browser vor, der für das Mesh gebaut ist, nicht für das öffentliche Web.

## Was Ren Browser kann

- Öffnet `.mu`-Seiten und andere NomadNet-Inhalte von Mesh-Zielen
- Zeigt Knoten, die Sie über Ihre Reticulum-Interfaces erreichen können
- Speichert Tabs, Verlauf, Lesezeichen und Einstellungen auf Ihrem Rechner (oder im Browser, wenn Sie den öffentlichen Servermodus nutzen)
- Rendert Micron-Markup mit einem WASM-Parser und eingebauten Themes
- Unterstützt Erweiterungen, die Panels, URL-Schemes und Befehle hinzufügen

## Was Sie zuerst brauchen

Bevor Ren Browser eine entfernte Seite laden kann, brauchen Sie eine funktionierende **Reticulum**-Einrichtung auf demselben Rechner (oder eingebunden in einen Docker-Container für den Servermodus).

Ren Browser liest die Reticulum-Konfiguration standardmäßig aus `~/.reticulum-go/`. Mit `--config` oder der Umgebungsvariable `REN_BROWSER_CONFIG` können Sie eine andere Datei angeben.

Sie brauchen **keine** herkömmliche Internetverbindung, um NomadNet-Seiten im Mesh zu durchsuchen. Sie brauchen mindestens ein Interface, das andere Reticulum-Knoten erreichen kann.

## Desktop oder Server

| Modus | Am besten für |
|-------|---------------|
| **Desktop** (Standard) | Tägliche Nutzung unter Linux, Windows oder macOS. Eigenes Fenster, lokale SQLite-Datenbank. |
| **Server** | Homelab, Docker oder ein Rechner, der Reticulum bereits betreibt. Sie öffnen Ren Browser in einem anderen Browser unter `http://host:8080`. |
| **Android** | Mobile Builds, wenn Ihr Release eine APK enthält. Dieselben Kernfunktionen in einem Touch-Layout. |

## Checkliste für den ersten Start

1. Reticulum installieren und eine Konfiguration unter `~/.reticulum-go/` anlegen oder kopieren
2. Ren Browser aus den [Releases](https://github.com/Quad4-Software/Ren-Browser/releases) installieren oder aus dem Quellcode bauen
3. Ren Browser starten und warten, bis Reticulum über Ihre Interfaces verbunden ist
4. **Discovery** öffnen oder einen 32-stelligen Ziel-Hash in die Adressleiste eingeben
5. `about:` aufrufen, um Version, Konfigurationspfad und Datenverzeichnis zu prüfen

## Eingebaute Seiten (ohne Mesh)

Diese funktionieren auch ohne Mesh-Verbindung:

| Adresse | Zweck |
|---------|-------|
| `about:` | App-Version, Build-Infos, Pfade |
| `license:` | MIT-Lizenztext |
| `editor:` | Eingebauter Micron-Editor |

Geben Sie `settings` in die Adressleiste ein oder drücken Sie das Tastenkürzel für Einstellungen, um die Voreinstellungen zu öffnen.

## Nächste Schritte

- [Installation](installation.md), wenn Sie noch nicht installiert haben
- [Reticulum einrichten](reticulum-setup.md), wenn Seiten nicht laden oder Discovery leer bleibt
- [Browser verwenden](using-the-browser.md) für den täglichen Gebrauch
