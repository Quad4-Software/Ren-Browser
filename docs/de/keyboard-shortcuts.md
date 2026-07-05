# Tastenkürzel

Standard-Tastenbelegungen auf dem Desktop. Sie können sie unter **Einstellungen → Tastenkürzel** ändern.

`mod` bedeutet **Ctrl** unter Windows und Linux, **Cmd** unter macOS.

## Standardbelegungen

| Aktion | Akkord | Beschreibung |
|--------|--------|--------------|
| Adressleiste fokussieren | `mod+l` | Tastaturfokus auf das URL-Feld |
| Seite neu laden | `mod+r` | Aktiven Tab neu laden |
| Entwicklertools | `mod+shift+i` | DevTools öffnen oder schließen |
| Auf Seite suchen | `mod+f` | Suchleiste auf der aktuellen Seite öffnen |
| Discovery-Panel | `mod+shift+d` | Discovery-Seitenleiste öffnen |
| Einstellungen-Panel | `mod+,` | Einstellungen öffnen |
| Neuer Tab | `mod+t` | Leeren Tab öffnen |
| Neues Fenster | `mod+shift+n` | Weiteres Fenster öffnen (wenn unterstützt) |
| Tab schließen | `mod+w` | Aktiven Tab schließen |
| Vollbild | `f11` | Vollbild umschalten |

## Neue Belegung aufzeichnen

1. Öffnen Sie **Einstellungen → Tastenkürzel**
2. Klicken Sie die Aktion, die Sie ändern möchten
3. Drücken Sie die neue Tastenkombination
4. Konflikte mit einer anderen Aktion erscheinen in der UI. Lösen Sie sie vor dem Speichern.

Während der Aufzeichnung sind andere Kürzel pausiert, damit der Recorder nur Ihren neuen Akkord sieht.

## Akkord-Syntax

Belegungen werden als Kleinbuchstaben-Akkorde mit `+` gespeichert:

- `mod`: Ctrl oder Cmd
- `shift`: Shift
- `alt`: Alt
- Letztes Segment ist der Tastenname (`l`, `r`, `,`, `f11` usw.)

Beispiel: `mod+shift+d` ist Ctrl+Shift+D unter Linux.

## Erweiterungs-Tastenkürzel

Erweiterungen können Befehle mit optionalem `keybind`-Feld in `renbrowser.plugin.json` bereitstellen. Aktivierte Erweiterungen fügen ihre Befehle in die Host-Befehlspalette und Kürzel-Verarbeitung ein.

## Android

Hardware-Tastaturen unter Android folgen denselben Akkord-Regeln, wenn angeschlossen. Die Touch-Oberfläche erfordert keine Kürzel.

## Nächste Schritte

- [Einstellungen](settings.md) für weitere Voreinstellungen
- [Browser verwenden](using-the-browser.md) für Panel-Überblick
