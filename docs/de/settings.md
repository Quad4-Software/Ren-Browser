# Einstellungen

Einstellungen steuern, wie Ren Browser aussieht, sich mit Reticulum verbindet und sich auf Ihrem Rechner verhält.

Öffnen Sie Einstellungen über die Seitenleiste oder mit `Ctrl+,` / `Cmd+,` (anpassbar in den Tastenkürzel-Einstellungen).

## Erscheinungsbild

### Themes

Ren Browser bringt dunkle und helle Themes mit. Sie können:

- Das Theme in Einstellungen wechseln
- Eine Theme-JSON-Datei importieren
- Ihr aktuelles Theme sichern oder teilen

Themes beeinflussen Chrome-Farben, Typografie-Tokens und Micron-Seiten-Styling, wo zutreffend.

### Fenster-Chrome

Auf dem Desktop können Sie zwischen einer nativen Titelleiste und einem rahmenlosen Fenster mit eigenen Minimieren-, Maximieren- und Schließen-Steuerelementen wählen. Der rahmenlose Modus nutzt ziehbare Bereiche oben im Fenster.

## Reticulum

Der Reticulum-Bereich zeigt:

- Aktive Interfaces und ihren Status
- Sende- und Empfangs-Bytezähler
- Konfigurationseditor mit Hot Reload

Nach dem Bearbeiten der Konfiguration wenden Sie Änderungen in Einstellungen an, statt die App neu zu starten, wenn möglich.

## Tastenkürzel

Jede Aktion kann ein Akkord wie `mod+t` für einen neuen Tab haben. `mod` bedeutet Ctrl unter Windows und Linux, Cmd unter macOS.

Siehe [Tastenkürzel](keyboard-shortcuts.md) für Standardbelegungen und wie Sie neue Bindungen aufzeichnen.

## Erweiterungen

Verwalten Sie Plugins unter **Einstellungen → Erweiterungen**:

- Aus einem Zip-Archiv installieren
- Aus einem Ordner installieren
- Installierte Erweiterungen aktivieren oder deaktivieren
- Berechtigungen jeder Erweiterung anzeigen

Siehe [Erweiterungen](extensions.md) für Manifest-Format und Installationspfade.

## Profil und Daten

Einstellungen zeigen Pfade für:

- SQLite-Datenbank-Speicherort
- Reticulum-Konfigurationspfad
- Plugin-Verzeichnis unter `~/.renbrowser/plugins/`

Für benannte Profile und Import oder Export siehe [Daten und Profile](data-and-profiles.md).

## Browser-Voreinstellungen

Weitere Schalter können umfassen:

- Native Titelleiste vs. rahmenloses Fenster
- Standard-Panel-Verhalten
- Optionen, die mit gespeicherten Browser-Voreinstellungen in SQLite synchronisiert werden

Die genauen Schalter können je nach Version variieren. Im Zweifel prüfen Sie die Beschriftung in der UI und dieses Dokument für Ihr Release-Tag.

## Mobiles Layout

Unter Android nutzen Einstellungen dieselben Daten, gruppieren Einträge aber ggf. für Touch-Navigation. Kernoptionen für Reticulum und Themes bleiben verfügbar.

## Nächste Schritte

- [Tastenkürzel](keyboard-shortcuts.md)
- [Erweiterungen](extensions.md)
- [Daten und Profile](data-and-profiles.md)
