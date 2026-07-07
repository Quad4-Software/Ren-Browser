# Sicherheit

Ren Browser ist für Systeme und Netzwerke gedacht, denen Sie vertrauen. Diese Seite fasst sichere Nutzung, Download-Prüfung und Meldewege zusammen.

## Vertrauensgrenzen

| Oberfläche | Risiko |
|------------|--------|
| Desktop-App | Lokale Webview mit Go-Bindings. Kein Node.js im Seiteninhalt. |
| Servermodus | Offener HTTP-Port. Kein Login enthalten. |
| Plugins | Laufen mit deklarierten Berechtigungen. WASM-Plugins sind mit Caps sandboxed. |
| Mesh-Inhalt | Unvertrauenswürdig wie jeder Netzwerkinhalt. Micron-HTML wird sanitisiert, außer ein Plugin fordert `render.unsanitized` an. |

## Servermodus

`renbrowser-server` hat **keine Authentifizierung**. Wenn Sie es ins Internet stellen, riskieren Sie:

- Automatisches Scannen
- Missbrauch Ihrer Reticulum-Interfaces
- Überlastung einer Single-Process-App

Wenn Sie es dennoch exponieren müssen:

1. Stellen Sie es hinter einen Reverse Proxy mit Zugriffskontrollen
2. Nutzen Sie HTTPS mit gültigem Zertifikat
3. Schränken Sie den Port per Firewall oder VPN ein
4. Halten Sie die App aktuell

Siehe [Servermodus](server-mode.md) für Proxy-Flags.

## Desktop-Plugins

Installieren Sie Erweiterungen nur von Personen oder Projekten, denen Sie vertrauen. Lesen Sie die Berechtigungsliste unter **Einstellungen → Erweiterungen** vor dem Aktivieren.

### Prüfungen bei der Installation

RenBrowser zeigt beim Installieren:

- Einzelschalter für Berechtigungen (deaktivierte Rechte gelten zur Laufzeit nicht)
- Gescannte Netzwerk-Endpunkte aus Manifest und Paketdateien
- Signatur-Badges (unsigniert, signiert, vertrauenswürdig, manipuliert)
- Heuristische Sicherheitsbewertung

Ungültige Signaturen blockieren die Installation. Unsignierte Erweiterungen können nach eigenem Risiko installiert werden.

### Laufzeitschutz

- Gewährte Berechtigungen für JS `PluginFetch` und WASM `http_fetch`
- WASM-Netzwerkexporte ohne `network.fetch` werden blockiert
- Limits für Plugin-HTTP-Anfragen und WASM-Arbeit gegen Einfrieren
- Integritäts-Hash der Dateien; externe Änderungen deaktivieren die Erweiterung
- Digest-geschützte Benutzer-Vertrauensliste

Plugin-HTTP erscheint in **Entwicklertools → Netzwerk**. Siehe [Erweiterungen](extensions.md).

## Downloads prüfen

Offizielle Builds kommen von [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) und GitHub Actions CI.

Jedes Release sollte `SHA256SUMS.txt` enthalten. Prüfen Sie Ihre Datei:

```sh
sha256sum -c SHA256SUMS.txt
```

Für Docker bevorzugen Sie das Anheften per Digest (`@sha256:...`), nachdem Sie einem Build vertrauen. Images auf GHCR enthalten Build-Provenance und ein SBOM von Docker Buildx.

Wenn eine Binärdatei nicht zu veröffentlichten Prüfsummen passt, behandeln Sie sie als nicht vertrauenswürdig.

## Subresource Integrity für WASM

Micron-Parser-WebAssembly und das Begleitskript `wasm_exec.js` werden vor der Ausführung per SHA-384 SRI geprüft. Ein Hash-Mismatch blockiert den Code und zeigt einen Fehler.

## Daten im Ruhezustand

- Anwendungszustand: SQLite unter `~/.renbrowser/`
- Reticulum-Schlüssel: Ihr Reticulum-Konfigurationsverzeichnis
- Server Public Mode: einige Daten nur im Browser-`localStorage` jedes Clients

Verschlüsseln Sie Festplatten auf OS-Ebene, wenn der Rechner geteilt oder tragbar ist.

## Schwachstellen melden

Öffnen Sie **kein** öffentliches GitHub-Issue für ungepatchte Sicherheitsfehler.

**Bevorzugter Kontakt:**

1. LXMF: `f489752fbef161c64d65e385a4e9fc74`

Geben Sie Version, Plattform, Reproduktionsschritte und Auswirkung an.

Rechts- und Lizenzfragen gehen an [LEGAL.md](../../LEGAL.md) (`legal@quad4.io`), nicht an den Sicherheitskanal.

## CI und Lieferkette (Überblick)

GitHub Actions führt Tests, gosec, Trivy-Scans und CodeQL nach Plan aus. Drittanbieter-Actions sind in Workflows an Commit-SHAs angeheftet.

## Nächste Schritte

- [Erweiterungen](extensions.md) Berechtigungsliste
- [Servermodus](server-mode.md) Deployment
- [SECURITY.md](../../SECURITY.md) im Repository-Stamm für die kanonische Richtlinie
