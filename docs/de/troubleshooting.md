# Fehlerbehebung

Häufige Probleme und was Sie zuerst versuchen sollten.

## Reticulum startet nicht

**Symptome:** Log-Zeile `reticulum start: ...`, Discovery leer, alle Mesh-Seiten schlagen fehl.

**Prüfungen:**

1. Konfigurationspfad in `about:` passt zu Ihren Dateien
2. `REN_BROWSER_CONFIG` oder `--config` zeigt auf eine gültige Datei
3. Interface-Definitionen sind syntaktisch korrekt
4. Schlüssel und Speicherpfade sind für den Benutzer lesbar, der Ren Browser startet

Beheben Sie die Reticulum-Konfiguration außerhalb der App, dann laden Sie in Einstellungen neu oder starten neu.

## Discovery ist leer

Siehe [Discovery](discovery.md) und [Reticulum einrichten](reticulum-setup.md).

Kurzliste:

- Nach der Verbindung auf Ankündigungen warten
- Prüfen, ob Peers auf Ihren Interfaces existieren
- Firewall-Regeln für konfigurierte UDP- oder TCP-Ports prüfen

## Seitenlade-Timeout

1. Hash in der Adressleiste prüfen
2. Einen anderen Knoten aus Discovery öffnen
3. Prüfen, ob Reticulum Traffic auf Interfaces in Einstellungen zeigt
4. Nach Funk- oder Pfadänderungen im Mesh erneut versuchen

## Datenbank korrupt oder öffnet nicht

**Symptome:** Fehler zu Profildaten, Angebot zum Zurücksetzen der Datenbank.

**Optionen:**

1. `renbrowser.db` aus Backup wiederherstellen ([Daten und Profile](data-and-profiles.md))
2. Über die UI zurücksetzen (löscht lokale Tabs, Verlauf, Favoriten, Einstellungen)
3. Die fehlerhafte Datei umbenennen und Ren Browser eine frische Datenbank anlegen lassen

Die Reticulum-Identität bleibt von einem Browser-DB-Reset unberührt.

## WASM- oder Micron-Parser-Fehler

Wenn die SRI-Prüfung für Micron-WASM fehlschlägt:

1. Die Prüfung nicht deaktivieren
2. Aus offiziellen Releases neu installieren
3. Bei Bau aus dem Quellcode `task build` erneut ausführen, ohne `frontend/dist/vendor/` von Hand zu ändern

## Servermodus: leere Seite oder falsche Assets

1. Prüfen, ob `--base-path` zum Reverse-Proxy-Mount passt
2. `--trust-proxy` aktivieren, wenn TLS upstream endet
3. Port-Mapping in Docker prüfen (`-p 8080:8080`)

## Servermodus: geteilter Verlauf, obwohl unerwünscht

Mit `--public-mode` starten, damit jeder Browser seine eigene `localStorage`-Kopie behält.

## Erweiterung lädt nicht

1. Manifest muss gültiges JSON in `renbrowser.plugin.json` sein
2. `id` muss dem Ordnernamen unter `plugins/` entsprechen
3. `engines.renbrowser` muss von Ihrer App-Version erfüllt sein
4. Unbekannte Berechtigungszeichenketten führen zum Ladefehler

Prüfen Sie Einstellungen auf die Fehlerzeichenkette.

## Android-Build schlägt fehl

1. `ANDROID_HOME` setzen
2. `task android:install:deps` ausführen
3. API 34 und NDK r26+ wie in [Installation](installation.md) dokumentiert nutzen

## Entwicklung: `task check` schlägt fehl

| Bereich | Befehl |
|---------|--------|
| Go-Format | `task fmt:go` |
| Go-Tests | `task test:go` |
| Frontend | `task frontend:check` |
| Sicherheitsscan | `task gosec` |

Führen Sie `task check` aus, bevor Sie Patches senden.

## Immer noch festgefahren

1. Version aus `about:` notieren
2. Logs aus Terminal oder Docker erfassen
3. In Ihrer Mesh-Community fragen oder einen detaillierten Fehlerbericht über Projektkanäle senden

Siehe [Mitwirken](contributing.md) für Patch-Einreichung über LXMF.
