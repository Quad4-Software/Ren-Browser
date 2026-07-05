# Discovery

Discovery zeigt NomadNet-Knoten, die über Ihre Reticulum-Interfaces angekündigt werden. Nutzen Sie es, wenn Sie den Ziel-Hash noch nicht kennen.

## Discovery öffnen

- Klicken Sie auf **Discovery** in der Seitenleiste
- Nutzen Sie das Tastenkürzel (`Ctrl+Shift+D` / `Cmd+Shift+D` standardmäßig)

Das Panel listet Knoten mit Anzeigenamen, Hashes und Metadaten, wenn die Ankündigung sie enthält.

## Was in der Liste erscheint

Ein Knoten erscheint, wenn:

1. Reticulum eine Ankündigung für dieses Ziel empfängt
2. Die Ankündigung zu NomadNet oder kompatiblen Knotentypen passt, die der Browser versteht
3. Ihre Interfaces den Pfad erreichen können, der die Ankündigung trug

Discovery ist live. Die Liste aktualisiert sich, wenn neue Ankündigungen eintreffen und alte ablaufen.

## Einen Knoten öffnen

Klicken Sie auf eine Zeile, um zur Standardseite des Knotens zu navigieren (meist `index.mu`). Sie können den Hash auch für die Adressleiste kopieren.

## Favoriten aus Discovery

Viele Zeilen bieten eine Möglichkeit, den Knoten zu Favoriten hinzuzufügen. Favoriten werden in Ihrer lokalen Profildatenbank gespeichert.

## Leere Liste

Wenn Discovery leer bleibt, gehen Sie [Reticulum einrichten](reticulum-setup.md) durch:

- Interfaces müssen in Einstellungen aktiv sein
- Sie brauchen Verbindung zu Peers, die Ankündigungen tragen
- Neue Beitritte können kurz brauchen, bis die Liste gefüllt ist

## Ankündigungen vs. Erreichbarkeit

Einen Knoten in Discovery zu sehen garantiert nicht, dass jeder Seitenaufruf gelingt. Sie brauchen weiterhin eine Route zum Ziel für LXMF-Seitenanfragen über Reticulum.

Wenn Discovery einen Knoten zeigt, Seiten aber fehlschlagen, prüfen Sie Transportpfade und ob der entfernte Knoten online ist.

## Community-Interfaces

Einstellungen können Community- oder gemeinsame Interface-Definitionen auflisten, wenn aktiviert. Diese helfen beim Beitritt zu größeren Mesh-Segmenten. Änderungen wenden Sie im Reticulum-Bereich in Einstellungen an.

## Nächste Schritte

- [Navigation und URLs](navigation-and-urls.md) für manuelle Hash-Eingabe
- [Browser verwenden](using-the-browser.md) für Favoriten und Verlauf
