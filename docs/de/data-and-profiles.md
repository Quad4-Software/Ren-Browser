# Daten und Profile

Ren Browser speichert Lesezeichen, Verlauf, Tabs, Einstellungen und Discovery-Cache auf der Festplatte. Diese Seite erklärt, wo diese Daten liegen und wie Profile funktionieren.

## Standardpfade

| Element | Pfad |
|---------|------|
| Datenverzeichnis | `~/.renbrowser/` |
| Hauptdatenbank | `~/.renbrowser/renbrowser.db` |
| Plugins | `~/.renbrowser/plugins/<id>/` |
| Benannte Profile | `~/.renbrowser/profiles/<name>/renbrowser.db` |
| Legacy-Zustand (migriert) | `~/.renbrowser/state.json` |

Beim ersten Start nach einem Upgrade importiert `state.json` automatisch in SQLite.

## Was SQLite enthält

Typische Tabellen und Blobs umfassen:

- Geöffnete Tabs und Sitzungswiederherstellung
- Browserverlauf mit Zeitstempeln
- Favoriten
- Browser-Voreinstellungen und Tastenkürzel
- Gecachte Discovery-Einträge
- Theme-Auswahl und eigene Theme-Daten

Korruption wird beim Öffnen erkannt. Die UI kann anbieten, die Datenbank zurückzusetzen. Ein Reset entfernt lokale Tabs, Verlauf, Favoriten und Einstellungen.

## Reticulum-Daten sind getrennt

Identitätsschlüssel und Interface-Konfiguration liegen in Ihrem Reticulum-Verzeichnis (standardmäßig `~/.reticulum-go/`). Ren Browser liest diesen Pfad, verschiebt Ihre Reticulum-Identität aber nicht nach `~/.renbrowser/`.

## Benannte Profile

Starten Sie mit `--profile NAME` oder `REN_BROWSER_PROFILE=NAME`, um zu nutzen:

```
~/.renbrowser/profiles/NAME/renbrowser.db
```

Nutzen Sie Profile, wenn Sie getrennte Verläufe auf einem Konto wollen (Arbeit vs. privat oder Tests).

## Import und Export

Nur beim Start:

- `--export-profile /path/to/backup.json` schreibt Profildaten und beendet sich
- `--import-profile /path/to/backup.json` führt Zusammenführung oder Ersetzung aus der Datei durch

Umgebungsspiegel: `REN_BROWSER_EXPORT_PROFILE`, `REN_BROWSER_IMPORT_PROFILE`.

Exportieren Sie vor größeren Upgrades oder beim Wechsel auf einen neuen Rechner.

## Speicher im Servermodus

| Modus | Tabs, Verlauf, Favoriten |
|-------|--------------------------|
| Standard-Server | Serverseitige SQLite im `~/.renbrowser/` des Servers |
| `--public-mode` | `localStorage` des jeweiligen Client-Browsers |

Wählen Sie Public Mode, wenn viele Nutzer eine Serverinstanz teilen.

## Theme-Import und -Export

Themes können als JSON aus Einstellungen exportiert und auf einer anderen Installation importiert werden. Theme-Dateien sind nicht das volle Profil, nur Erscheinungs-Tokens.

## Plugin-Daten

Erweiterungen mit `storage.plugin`-Berechtigung erhalten isolierten Speicher nach Plugin-ID. Deinstallieren entfernt den Ordner nicht immer. Löschen Sie `~/.renbrowser/plugins/<id>/` manuell für eine saubere Entfernung.

## Android

Mobile Builds nutzen dasselbe logische Layout in der App-Sandbox. Pfade unterscheiden sich nach OS-Regeln, das Datenbankschema entspricht dem Desktop.

## Backup-Checkliste

1. Ren Browser beenden
2. `~/.renbrowser/renbrowser.db` kopieren (oder Ihren Profilpfad)
3. `~/.reticulum-go/` kopieren, wenn Sie auch die Mesh-Identität wollen
4. `~/.renbrowser/plugins/` kopieren, wenn Sie Erweiterungen nutzen

Wiederherstellen durch Zurücklegen der Dateien vor dem nächsten Start.

## Nächste Schritte

- [Einstellungen](settings.md) für UI-Pfade
- [Servermodus](server-mode.md) für Public Mode
- [Fehlerbehebung](troubleshooting.md), wenn die Datenbank nicht öffnet
