# Reticulum einrichten

Ren Browser nutzt [Reticulum](https://reticulum.network/) über den `quad4/reticulum-go`-Stack. Diese Seite erklärt, was die App erwartet und wie Sie häufige Mesh-Probleme beheben.

## Standardpfad der Konfiguration

| Element | Standardpfad |
|---------|--------------|
| Reticulum-Konfigurationsverzeichnis | `~/.reticulum-go/` |
| Überschreiben per Flag | `--config /path/to/config` |
| Überschreiben per Umgebungsvariable | `REN_BROWSER_CONFIG` oder `RETICULUM_CONFIG` |

Die genauen Dateien im Verzeichnis hängen von Ihrer Reticulum- oder reticulum-go-Einrichtung ab. Ren Browser startet den Stack beim Start und lädt Interface-Änderungen aus **Einstellungen** neu.

## Was beim Start passiert

1. Ren Browser lädt Ihre Reticulum-Konfiguration
2. Interfaces gehen online (UDP, TCP, RNode und andere, die Sie konfiguriert haben)
3. Ankündigungen von NomadNet-Knoten erscheinen in **Discovery**
4. Seitenanfragen laufen über LXMF und Reticulum zu Knoten, die Micron-Seiten hosten

Wenn der Start fehlschlägt, prüfen Sie das Terminal-Log (Desktop) oder die Container-Logs (Server). Die App läuft weiter, damit Sie weiterhin `about:` und **Einstellungen** öffnen können.

## Interfaces in Einstellungen

Öffnen Sie **Einstellungen** und suchen Sie den Reticulum-Bereich. Dort können Sie:

- Sehen, welche Interfaces aktiv sind
- Sende- und Empfangsstatistiken anzeigen
- Die Konfiguration bearbeiten und per Hot Reload anwenden, ohne die ganze App neu zu starten

Nutzen Sie das, wenn Sie ein neues Interface hinzufügen oder Schlüssel ändern und der Browser die Änderungen schnell übernehmen soll.

## Dem Mesh beitreten

Sie brauchen mindestens einen Pfad zu anderen Reticulum-Knoten. Übliche Optionen:

- **Lokales UDP oder TCP** in einem LAN mit anderen Reticulum-Peers
- **RNode** oder ähnliche Funk-Hardware
- **Interface-Definitionen**, die auf bekannte Peers oder Hubs zeigen

Reticulum liegt außerhalb dieses Handbuchs. Lesen Sie das [Reticulum-Handbuch](https://reticulum.network/manual/) für Interface-Syntax und Identitätsverwaltung.

## NomadNet-Ziele

NomadNet-Seiten liegen an Reticulum-Zielen. In der Adressleiste können Sie verwenden:

- Einen vollständigen Pfad wie `abcdef0123456789abcdef0123456789:/page/index.mu`
- Einen bloßen 32-stelligen Hex-Hash (Ren Browser hängt `:/page/index.mu` an)

Seiten nutzen das Micron-Markup-Format. Ren Browser rendert sie mit der eingebauten Micron-Pipeline.

## Wenn Discovery leer bleibt

Gehen Sie diese Liste durch:

1. Prüfen Sie, ob Reticulum in Ren Browser läuft (Einstellungen zeigen Interfaces)
2. Prüfen Sie, ob Ihre Interfaces zur Konfiguration der Peers im Mesh passen
3. Warten Sie kurz nach der Verbindung. Ankündigungen kommen nicht sofort
4. Prüfen Sie, ob Sie im selben logischen Netzwerk wie die erwarteten Knoten sind

## Wenn Seiten Zeitüberschreitungen haben oder fehlschlagen

1. Prüfen Sie, ob der Ziel-Hash korrekt ist
2. Prüfen Sie, ob Sie eine Route zu diesem Ziel haben (nicht nur Sichtbarkeit in Discovery)
3. Versuchen Sie einen anderen bekannten Knoten aus Discovery
4. Schauen Sie in DevTools oder Logs nach LXMF- oder Transportfehlern

## Server und Docker

Wenn Sie `renbrowser-server` in Docker betreiben, binden Sie die Host-Konfiguration schreibgeschützt ein:

```sh
-v "$HOME/.reticulum-go:/root/.reticulum-go:ro"
```

Der Container-Benutzer muss Schlüssel und Interface-Definitionen in diesem Verzeichnis lesen können.

## Nächste Schritte

- [Discovery](discovery.md) zum Durchsuchen angekündigter Knoten
- [Navigation und URLs](navigation-and-urls.md) für Adressleisten-Formate
- [Fehlerbehebung](troubleshooting.md) für Fehlermeldungen
