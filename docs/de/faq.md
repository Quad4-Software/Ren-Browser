# Häufige Fragen

Kurze Antworten auf häufige Fragen.

## Was ist Ren Browser?

Ein Browser für NomadNet-Seiten im Reticulum-Mesh. Er ist kein allgemeiner Webbrowser für das öffentliche Internet.

## Brauche ich das Internet?

Sie brauchen Reticulum-Verbindung zu anderen Knoten. Das kann völlig ohne das öffentliche Internet sein (Funk, LAN usw.).

## Wo bekomme ich reticulum-go?

Installieren und konfigurieren Sie [reticulum-go](https://reticulum-go.quad4.io) auf dem Rechner, auf dem Ren Browser läuft. Die App nutzt Ihre Reticulum-Konfiguration (zum Beispiel unter `~/.reticulum-go/`), erstellt aber weder Identität noch Interfaces für Sie.

## Was ist NomadNet?

[Nomad Network](https://github.com/markqvist/NomadNet) ist ein Off-Grid-Mesh-Kommunikationsprogramm auf Basis von LXMF und Reticulum. Anschließbare Knoten können Seiten und Dateien hosten, oft in der Micron-Auszeichnungssprache. Ren Browser bindet den NomadNet-Client nicht ein. Es findet Knoten, die sich als NomadNet ankündigen, und öffnet deren gehostete Seiten über Reticulum.

## Was ist Micron?

Eine bandbreiteneffiziente Auszeichnungssprache auf Nomad Network-Knoten. Ren Browser rendert sie zu HTML im Inhaltsviewer.

## Kann ich normale HTTPS-Seiten öffnen?

Nein. Ren Browser zielt auf Reticulum-Mesh-Inhalte, einschließlich Seiten auf NomadNet-Knoten, nicht auf beliebige öffentliche Web-URLs.

## Desktop vs. Server?

Desktop startet ein natives Fenster und speichert Daten in lokaler SQLite. Der Servermodus liefert die UI per HTTP für die Nutzung in einem anderen Browser. Siehe [Servermodus](server-mode.md).

## Ist der Servermodus im Internet sicher?

Nicht standardmäßig. Es gibt kein Login. Nutzen Sie VPN, Firewall oder einen authentifizierten Reverse Proxy. Siehe [Sicherheit](security.md).

## Wo liegen meine Daten?

`~/.renbrowser/renbrowser.db` auf dem Desktop standardmäßig. Siehe [Daten und Profile](data-and-profiles.md).

## Wie prüfe ich einen Release-Download?

Nutzen Sie `SHA256SUMS.txt` von der Release-Seite. Siehe [Sicherheit](security.md).

## Wie installiere ich eine Erweiterung?

Einstellungen → Erweiterungen, oder entpacken nach `~/.renbrowser/plugins/<id>/`. Siehe [Erweiterungen](extensions.md).

## Wie gebe ich eine Knotenadresse ein?

Fügen Sie den 32-stelligen Hex-Hash oder die vollständige URL `hash:/page/file.mu` ein. Siehe [Navigation und URLs](navigation-and-urls.md).

## Discovery zeigt nichts

Prüfen Sie Reticulum-Interfaces in Einstellungen und [Reticulum einrichten](reticulum-setup.md).

## Wie melde ich einen Fehler?

Für Sicherheitsprobleme LXMF gemäß [Sicherheit](security.md). Für Code-Fixes siehe [Mitwirken](contributing.md).

## Unter welcher Lizenz steht das Projekt?

MIT. Geben Sie `license:` in die Adressleiste ein oder lesen Sie [LICENSE](../../LICENSE).

## Mit welchem Stack ist es gebaut?

Go, Wails v3, Svelte 5, SQLite und Quad4-Reticulum-Bibliotheken. Siehe [Architektur](architecture.md).

## Kann ich unter Android laufen?

Ja, wenn für Ihr Release eine APK veröffentlicht ist. Bauen Sie bei Bedarf aus dem Quellcode mit dem Android SDK. Siehe [Installation](installation.md).

## Wie ändere ich Tastenkürzel?

Einstellungen → Tastenkürzel. Siehe [Tastenkürzel](keyboard-shortcuts.md).
