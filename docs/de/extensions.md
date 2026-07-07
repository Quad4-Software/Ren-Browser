# Erweiterungen

Ren Browser unterstützt Plugins, die URL-Schemes, Seitenpanels, Befehle, Themes, Einstellungsseiten und DevTools-Tabs hinzufügen.

## Erweiterungen installieren

### Über Einstellungen

1. Öffnen Sie **Einstellungen → Erweiterungen**
2. Wählen Sie **Erweiterung installieren**, dann **.zip**, **Ordner** oder **gebündeltes .wasm-Modul**
3. Prüfen Sie die Installationsvorschau:
   - Angeforderte Berechtigungen (einzelne Berechtigungen können vor der Installation deaktiviert werden)
   - Externe URLs, die die Erweiterung kontaktieren kann (aus Manifest und Paketdateien gescannt)
   - Status der Herausgeber-Signatur (unsigniert, signiert, vertrauenswürdiger Herausgeber, ungültig)
   - Sicherheitsbewertung
   - Gebündelte UI-Sprachen (wenn `locales/*.json` mitgeliefert wird)
4. Bestätigen und die Erweiterung aktivieren

Erweiterungen mit `network.fetch` zeigen einen Bestätigungsdialog mit erkannten Endpunkten. Die URL-Liste bleibt sichtbar, auch wenn Sie `network.fetch` vor der Installation deaktivieren.

### Manuelle Installation

Entpacken Sie ein Plugin nach:

```
~/.renbrowser/plugins/<id>/
```

Der Ordner muss `renbrowser.plugin.json` enthalten. Die `id` im Manifest sollte dem Ordnernamen entsprechen.

## Beispiel-Erweiterungen

Das Repository enthält `extensions/hello-extension/`:

- Registriert das `hello:`-URL-Scheme
- Fügt ein **Hello**-Seitenpanel hinzu
- Definiert den Befehl **Say hello** mit `mod+shift+h`

`extensions/micron-translator/` übersetzt Micron-Seiten (`.mu`) über Google Translate oder LibreTranslate. Befehle: **Translate Micron page** (`mod+shift+t`) und **Restore original** (`mod+shift+r`).

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

Optionale Felder: `description`, `author`, `license`, `engines`, `backend`, `network`, `contributes`.

### Engine-Einschränkung

```json
"engines": { "renbrowser": ">=0.1.0" }
```

Der Host lädt das Plugin nicht, wenn Ihre App-Version zu alt ist.

### Netzwerk-Endpunkte

Erweiterungen mit `network.fetch` sollten kontaktierte Hosts oder URLs deklarieren:

```json
"network": {
  "endpoints": [
    "https://api.example.com/",
    "Benutzerkonfigurierte Dienst-URL"
  ]
}
```

Bei der Installation scannt RenBrowser zusätzlich `.js`, `.go`, `.wasm` und andere Paketdateien nach `http`/`https`-URLs.

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

Der Host erzwingt Berechtigungen zur Laufzeit. Bei der Installation deaktivierte Berechtigungen werden gespeichert und nicht an JS `ctx.network.fetch` oder WASM `http_fetch` vergeben.

## Herausgeber-Signaturen

Erweiterungen können eine Ed25519-Signatur in `renbrowser.plugin.rsg` enthalten (kompatibel mit Reticulum `rnid`). Ungültige Signaturen blockieren die Installation.

Badges in Vorschau und Liste:

| Badge | Bedeutung |
|-------|-----------|
| Nicht signiert | Keine Signaturdatei |
| Signiert | Gültige Reticulum-Identität |
| Vertrauenswürdig | Signiert von einem vertrauenswürdigen Herausgeber |
| Manipuliert | Dateien außerhalb von RenBrowser geändert (deaktiviert bis zur erneuten Aktivierung) |

Bei der Installation können Sie **Dieser Herausgeber-Identität vertrauen** wählen. Die Benutzerliste liegt in `~/.renbrowser/trusted_publishers.json` und ist per Digest in der Profil-Datenbank geschützt.

Signieren mit `build/scripts/sign-extension.sh` (Python `rnid` erforderlich).

## Plugin-UI-Übersetzungen

Erweiterungen können UI-Texte unter `locales/<code>.json` mitliefern. Manifest-Titel können `%schlüssel.pfad%` verwenden; der Host lädt Kataloge von `/_plugins/<id>/locales/<code>.json`.

Die Installationsvorschau listet vorhandene Locale-Codes.

## Frontend-Einstiegsskript

Typische `main.js`-Exporte: `activate(ctx)`, `deactivate()`, `mount(el)`, `handleScheme(url)`.

Mit gewährtem `network.fetch`: `ctx.network.fetch()`. Prüfen Sie `ctx.capabilities.networkFetch` vor netzwerkabhängiger Arbeit.

Mit WASM-`backend`: `ctx.wasm.call("export", input)`. Strings über `ctx.i18n.t("key")`.

## Gebündelte WASM-Module

Eine `.wasm`-Datei kann Manifest (`renbrowser.plugin`), Dateien (`renbrowser.files`) und optional eine Signatur (`renbrowser.signature`) enthalten.

Installieren über **Einstellungen → Erweiterungen → .wasm-Modul wählen**.

`extensions/micron-translator/` nutzt TinyGo (`build-wasm.sh`, `go run ./extensions/micron-translator/bundle`).

## WASM-Backend

`backend` verweist auf ein WASM-Modul. Der Host stellt bei gewährtem `network.fetch` `renhost.http_fetch` bereit. Es gelten Anfrage-Limits, Timeouts und Größenobergrenzen.

## DevTools

Unter **Entwicklertools → Netzwerk** erscheinen ausgehende Plugin-HTTP-Anfragen als **Erweiterungsabruf** mit Status und Dauer.

## Integrität

Nach der Installation speichert RenBrowser einen Hash der Paketdateien. Änderungen außerhalb der App deaktivieren die Erweiterung (**Manipuliert**). Erneutes Aktivieren akzeptiert den aktuellen Stand.

## Sicherheitshinweise

- Nur vertrauenswürdige Quellen
- Berechtigungen und Endpunkte vor der Installation lesen
- Signierte Erweiterungen bevorzugen
- Plugins wie lokale Programme mit Profilzugriff behandeln

## Nächste Schritte

- Quellreferenz: `internal/plugins/manifest.go`
- [Sicherheit](security.md)
- [Entwicklung](development.md)
