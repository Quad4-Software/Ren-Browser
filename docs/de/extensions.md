# Erweiterungen

Ren Browser unterstützt Plugins, die URL-Schemes, Seitenpanels, Befehle, Themes, Einstellungsseiten und DevTools-Tabs hinzufügen.

## Erweiterungen installieren

### Über Einstellungen

1. Öffnen Sie **Einstellungen → Erweiterungen**
2. Wählen Sie **Aus Zip installieren** oder **Aus Ordner installieren**
3. Prüfen Sie, ob das Manifest lädt und die Berechtigungen plausibel sind
4. Aktivieren Sie die Erweiterung

### Manuelle Installation

Entpacken Sie ein Plugin nach:

```
~/.renbrowser/plugins/<id>/
```

Der Ordner muss `renbrowser.plugin.json` enthalten. Die `id` im Manifest sollte dem Ordnernamen entsprechen.

## Beispiel-Erweiterung

Das Repository enthält `extensions/hello-extension/`:

- Registriert das `hello:`-URL-Scheme
- Fügt ein **Hello**-Seitenpanel hinzu
- Definiert den Befehl **Say hello** mit `mod+shift+h`

Nutzen Sie es als Vorlage für eigene Plugins.

## Manifest-Datei

Dateiname: `renbrowser.plugin.json`

Pflichtfelder:

| Feld | Zweck |
|------|-------|
| `manifestVersion` | Derzeit `1` |
| `id` | Eindeutige ID (`a-z`, `A-Z`, `0-9`, `.`, `-`, 3 bis 128 Zeichen) |
| `name` | Anzeigename |
| `version` | Semver-Zeichenkette |
| `main` | Frontend-Einstiegsskript (optional, wenn nur Backend) |
| `permissions` | Berechtigungsliste (siehe unten) |

Optionale Felder sind `description`, `author`, `license`, `engines`, `backend` und `contributes`.

### Engine-Einschränkung

```json
"engines": { "renbrowser": ">=0.1.0" }
```

Der Host lädt das Plugin nicht, wenn Ihre App-Version zu alt ist.

### Beiträge

| Typ | Zweck |
|-----|-------|
| `urlSchemes` | Eigene Schemes verarbeiten |
| `panels` | Seitenleisten- oder andere Panel-Slots |
| `commands` | Befehlspalette und Tastenkürzel |
| `themes` | Zusätzliche Theme-JSON-Dateien |
| `settings` | Einstellungs-Unterseiten |
| `devtools` | DevTools-Tabs |
| `renderers` | Eigene Renderer für MIME-Typen oder Erweiterungen |

## Berechtigungen

Plugins müssen deklarieren, was sie brauchen. Bekannte Berechtigungen:

| Berechtigung | Erlaubt |
|--------------|---------|
| `storage.plugin` | Privater Schlüssel-Wert-Speicher für das Plugin |
| `navigation.read` | Aktuelle URL und Tab-Infos lesen |
| `navigation.write` | Navigation auslösen |
| `network.fetch` | Abruf über erlaubte Netzwerk-APIs |
| `events.emit` | Host-Ereignisse senden |
| `events.subscribe` | Auf Host-Ereignisse hören |
| `devtools.network` | Zusätzliche Netzwerkdetails in DevTools |
| `render.unsanitized` | Einige HTML-Sanitisierung überspringen (gefährlich) |

Der Host erzwingt Berechtigungen zur Laufzeit. Ein Plugin kann keine Fähigkeit nutzen, die es nicht deklariert hat.

## Frontend-Einstiegsskript

Ein typisches `main.js` exportiert:

- `activate(ctx)`: Ereignisse abonnieren, UI registrieren
- `deactivate()`: Aufräumen
- `mount(el)`: Seitenpanel-HTML rendern
- `handleScheme(url)`: für URL-Scheme-Handler

Die Hello-Erweiterung zeigt minimale Versionen davon.

## WASM-Backend

Plugins können `backend` auf einen WASM-Modulpfad setzen für schwerere Logik. WASM-Plugins laufen in einer eingeschränkten Laufzeit mit expliziten Grants.

## Sicherheitshinweise

- Installieren Sie Plugins nur aus Quellen, denen Sie vertrauen
- Lesen Sie die Berechtigungsliste vor dem Aktivieren
- Behandeln Sie Plugins wie jedes lokale Programm mit Zugriff auf Ihre Profildaten

## Nächste Schritte

- Quellreferenz: `internal/plugins/manifest.go` im Repository
- [Sicherheit](security.md) für Plugin-Bedrohungsmodell
- [Entwicklung](development.md) zum Arbeiten am Plugin-Host
