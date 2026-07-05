# Navigation und URLs

Ren Browser akzeptiert mehrere URL-Formen in der Adressleiste. Diese Seite listet sie auf und erklärt die Normalisierung.

## NomadNet-Ziele

Eine vollständige NomadNet-URL sieht so aus:

```
<32-hex-chars>:/page/path.mu
```

Beispiel:

```
a1b2c3d4e5f6789012345678abcdef01:/page/index.mu
```

Der Hash ist eine Reticulum-Zielidentität in Hexadezimal. Der Pfad nach `:/` ist die Datei auf diesem NomadNet-Knoten.

## Kurzform Hash

Wenn Sie nur eine 32-stellige Hex-Zeichenkette eingeben, erweitert Ren Browser sie zu:

```
<hash>:/page/index.mu
```

Das entspricht dem üblichen NomadNet-Startseitenpfad.

## Pfade ohne Hash

Wenn der aktuelle Kontext bereits ein Ziel hat, kann ein Pfad, der mit `/page/` beginnt, relativ zu diesem Knoten aufgelöst werden. Für eine Navigation ohne Kontext bevorzugen Sie die vollständige Form `hash:/page/...`.

## Eingebaute Schemes

Diese Schemes werden in der App verarbeitet. Sie nutzen nicht das Mesh.

| Scheme | Aliase | Beschreibung |
|--------|--------|--------------|
| `about:` | `about` | Version, Build, Reticulum-Konfigurationspfad, Datenverzeichnis |
| `license:` | `license` | Projekt-Lizenz (MIT) |
| `editor:` | `editor` | Micron-Quelltext-Editor |

Die Zuordnung ist unabhängig von Groß- und Kleinschreibung. Nachgestellte Leerzeichen werden entfernt.

## Erweiterungs-URL-Schemes

Installierte Erweiterungen können eigene Schemes in `renbrowser.plugin.json` registrieren. Die Hello-Erweiterung registriert zum Beispiel `hello:`. Siehe [Erweiterungen](extensions.md).

## Einstellungen und interne Routen

Die UI kann interne Routen für Panels nutzen. Die Adressleiste konzentriert sich auf Mesh- und eingebaute Schemes. Öffnen Sie **Einstellungen** über die Schaltfläche in der Seitenleiste oder `Ctrl+,` / `Cmd+,`.

## Tab-Titel

Tab-Titel stammen aus:

1. Seiten-Metadaten, wenn der Knoten einen Titel liefert
2. Anzeigenamen aus Discovery, wenn der Hash zu einem bekannten Knoten passt
3. Einem gekürzten Hash oder Pfad als Fallback

## Verlaufseinträge

Jede Navigation, die Inhalt lädt, kann eine Verlaufszeile mit URL, Titel, Ziel-Hash und Zeitstempel erzeugen. Eingebaute Seiten wie `about:` sind enthalten.

## Link-Auflösung auf Micron-Seiten

Wenn Sie auf einen Link auf einer gerenderten Seite klicken:

- `about:`, `license:` und `editor:` öffnen lokal
- Absolute Mesh-URLs navigieren direkt
- Relative Pfade werden mit dem Ziel der aktuellen Seite kombiniert

## Normalisierungsregeln (Zusammenfassung)

| Sie geben ein | Normalisierte URL |
|---------------|-------------------|
| `about` | `about:` |
| `license` | `license:` |
| `editor` | `editor:` |
| `abcdef...` (32 Hex) | `abcdef...:/page/index.mu` |
| Vollständig `hash:/page/foo.mu` | unverändert |

## Nächste Schritte

- [Browser verwenden](using-the-browser.md) für Tabs und Panels
- [Discovery](discovery.md) zum Auswählen von Knoten ohne Hash-Eingabe
