# Mitwirken

Ren Browser nimmt Patches entgegen, die über Reticulum gesendet werden. GitHub-Pull-Requests können je nach Projektrichtlinie ebenfalls willkommen sein. Diese Seite folgt [CONTRIBUTING.md](../../CONTRIBUTING.md).

## Patch-Workflow

1. Repository klonen oder forken
2. Branch anlegen und fokussierte Änderungen vornehmen
3. `task check` ausführen, wenn Sie Go- oder Frontend-Code anfassen
4. Mit klarer Nachricht committen
5. Patches mit `git format-patch` exportieren
6. Die `.patch`-Datei über LXMF senden

## Patches über LXMF senden

Ziel:

```
f489752fbef161c64d65e385a4e9fc74
```

Hängen Sie den Patch mit Sideband, Meshchat, MeshChatX oder einem beliebigen LXMF-Client mit Dateianhängen an. Fügen Sie eine kurze Beschreibung in den Nachrichtentext ein.

Geduld. Review läuft auf Mesh-Zeit.

## Export-Befehle

```sh
# Letzter Commit
git format-patch -1

# Letzte N Commits
git format-patch -N

# Alle Commits seit main
git format-patch main..HEAD
```

Jeder Commit wird eine `.patch`-Datei.

## Patch-Richtlinien

- Eine logische Änderung pro Patch-Serie, wenn möglich
- Vor dem Senden testen
- Bestehenden Code-Stil einhalten
- `// SPDX-License-Identifier: MIT` auf neuen Go-Dateien beibehalten
- KI-Nutzung im Nachrichtentext offenlegen (siehe unten)

## Lizenzierung

Mit dem Einreichen eines Patches stimmen Sie zu, dass er unter der [MIT License](../../LICENSE) lizenziert ist. Sie bestätigen, dass Sie das Recht haben, die Arbeit einzureichen.

## Richtlinie zu generativer KI

Sie dürfen KI-Tools nutzen, wenn:

- Ihr Setup dem Modell genug Kontext gibt
- Ihr Anbieter nicht mit dem Code trainiert, den Sie einfügen

Lesen Sie [Reticulum Zen](https://reticulum.network/manual/zen.html) und die [Reticulum License](https://reticulum.network/manual/license.html).

**Offenlegen**, welche Tools Sie in der Patch-Nachricht genutzt haben. Wenn Sie KI nicht sinnvoll genutzt haben, sagen Sie das kurz.

Lokale oder Offline-Modelle sind stark bevorzugt.

Sie müssen trotzdem alles lesen, verstehen und testen, was Sie einreichen. Ungeprüfte Massenausgabe wird nicht akzeptiert.

## Sicherheitsprobleme

Senden Sie Schwachstellendetails nicht als beiläufige Patches über LXMF ohne Abstimmung. Nutzen Sie den Prozess in [Sicherheit](security.md).

## Entwicklungsumgebung

Siehe [Entwicklung](development.md) für `task dev`, `task check` und Repository-Layout.

## Nächste Schritte

- [Entwicklung](development.md)
- [Häufige Fragen](faq.md)
