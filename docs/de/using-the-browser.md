# Browser verwenden

Diese Seite behandelt den täglichen Gebrauch: Tabs, Adressleiste, Verlauf, Favoriten und Seitenanzeige.

## Fensteraufbau

Die Hauptbereiche sind:

- **Tab-Leiste** oben. Tabs ziehen, anheften oder geteilte Ansichten öffnen.
- **Adressleiste** für Ziele und eingebaute Schemes.
- **Inhaltsbereich**, in dem Micron-Seiten gerendert werden.
- **Seitenpanels** für Discovery, Verlauf, Einstellungen, DevTools und Erweiterungs-Panels.

Auf kleineren Bildschirmen ersetzt eine mobile Navigationsleiste Teile der Desktop-Oberfläche.

## Tabs

- **Neuer Tab**: Standardkürzel `Ctrl+T` / `Cmd+T` (siehe [Tastenkürzel](keyboard-shortcuts.md))
- **Tab schließen**: `Ctrl+W` / `Cmd+W`
- **Sitzung wiederherstellen**: geöffnete Tabs werden in der lokalen Datenbank gespeichert und beim nächsten Start wiederhergestellt

Angeheftete Tabs bleiben vorne in der Tab-Leiste. Die geteilte Ansicht zeigt zwei Tabs nebeneinander.

## Adressleiste

Geben Sie ein NomadNet-Ziel ein oder nutzen Sie ein eingebautes Scheme:

| Eingabe | Ergebnis |
|---------|----------|
| 32-stelliger Hex-Hash | Öffnet `hash:/page/index.mu` |
| Vollständige NomadNet-URL | Öffnet wie eingegeben |
| `about:` | About-Seite mit Version und Pfaden |
| `license:` | Lizenztext |
| `editor:` | Micron-Editor |
| `settings` | Öffnet Einstellungen (über UI-Routing) |

Drücken Sie Enter zum Navigieren. Die Leiste lässt sich auch mit `Ctrl+L` / `Cmd+L` fokussieren.

## Links folgen

Micron-Links auf einer Seite werden relativ zum aktuellen Ziel aufgelöst. Interne `about:`, `license:`- und `editor:`-Links funktionieren wie normale Navigation.

Externe Mesh-Links nutzen die Reticulum-Ziel-Syntax. Wenn ein Link fehlschlägt, prüfen Sie, ob der Zielknoten erreichbar ist.

## Verlauf

Öffnen Sie das Verlauf-Panel über die Seitenleiste oder das zugehörige Tastenkürzel. Dort können Sie:

- Nach Titel oder URL suchen
- Einträge nach Datum gruppiert sehen
- Eine frühere Seite im aktuellen oder einem neuen Tab öffnen

Der Verlauf wird lokal in SQLite gespeichert (Desktop), sofern Sie nicht den öffentlichen Servermodus nutzen.

## Favoriten

Speichern Sie häufig besuchte Knoten über das Seitenkontextmenü oder Discovery. Favoriten nutzen denselben Speicher wie Verlauf und Tabs.

## Auf der Seite suchen

Drücken Sie `Ctrl+F` / `Cmd+F`, um Text auf der aktuellen Micron-Seite zu suchen. Treffer werden im Inhaltsviewer hervorgehoben.

## Entwicklertools

Drücken Sie `Ctrl+Shift+I` / `Cmd+Shift+I`, um DevTools zu öffnen. Nützlich für:

- Render-Zeiten prüfen
- Rohen Seitenquelltext anzeigen
- Erweiterungs-Panels debuggen, die DevTools-Einträge bereitstellen

## Downloads

Wenn eine Seite oder Aktion eine Datei anbietet, erscheinen Einträge im Download-Menü. Pfade folgen den Download-Konventionen Ihres Betriebssystems auf dem Desktop.

## Seitenfehler

Wenn eine Seite nicht geladen werden kann, sehen Sie einen Fehlerzustand mit einer kurzen Meldung. Häufige Ursachen:

- Ziel nicht erreichbar
- Ungültiger Micron-Inhalt
- Reticulum nicht verbunden

Nutzen Sie **Neu laden** (`Ctrl+R` / `Cmd+R`), nachdem Sie die Verbindung behoben haben.

## Micron-Editor

Öffnen Sie `editor:`, um Micron im eingebauten Editor zu verfassen. Nutzen Sie das für lokale Entwürfe, bevor Sie auf einem NomadNet-Knoten veröffentlichen.

## Nächste Schritte

- [Navigation und URLs](navigation-and-urls.md) für URL-Regeln im Detail
- [Discovery](discovery.md) zum Durchsuchen des Mesh
- [Einstellungen](settings.md) für Themes und Interfaces
