# よくある質問

よくある質問への簡潔な回答。

## Ren Browser とは何ですか？

Reticulum メッシュ上の NomadNet ページ用のブラウザです。公開インターネット向けの汎用ウェブブラウザではありません。

## インターネットは必要ですか？

他のノードへの Reticulum 接続が必要です。公開インターネットから完全にオフライン（無線、LAN など）でも構いません。

## reticulum-go はどこで入手できますか？

Ren Browser を実行するマシンに [reticulum-go](https://reticulum-go.quad4.io) をインストールして設定してください。アプリは Reticulum の設定（例：`~/.reticulum-go/` 以下）を使用しますが、アイデンティティやインターフェースの作成は行いません。

## NomadNet とは何ですか？

[Nomad Network](https://github.com/markqvist/NomadNet) は LXMF と Reticulum 上に構築されたオフグリッドメッシュ通信プログラムです。接続可能なノードはページやファイルをホストでき、多くの場合 Micron マークアップ言語で書かれています。Ren Browser は NomadNet クライアントを組み込んでいません。NomadNet としてアナウンスするノードを見つけ、Reticulum 経由でホストされたページを開きます。

## Micron とは何ですか？

Nomad Network ノードで使用される帯域効率の良いマークアップ言語です。Ren Browser はコンテンツビューアーで HTML にレンダリングします。

## 通常の HTTPS サイトを閲覧できますか？

できません。Ren Browser は NomadNet ノード上のページを含む Reticulum メッシュコンテンツを対象としており、任意の公開 URL ではありません。

## デスクトップとサーバーの違いは？

デスクトップはネイティブウィンドウを実行し、データをローカル SQLite に保存します。サーバーモードは別のブラウザで使用するために HTTP 経由で UI を提供します。[サーバーモード](server-mode.md) を参照してください。

## サーバーモードはインターネット上で安全ですか？

デフォルトではありません。ログイン機能がありません。VPN、ファイアウォール、または認証付きリバースプロキシを使用してください。[セキュリティ](security.md) を参照してください。

## データはどこにありますか？

デスクトップのデフォルトは `~/.renbrowser/renbrowser.db` です。[データとプロファイル](data-and-profiles.md) を参照してください。

## リリースのダウンロードを検証するには？

リリースページの `SHA256SUMS.txt` を使用してください。[セキュリティ](security.md) を参照してください。

## 拡張機能をインストールするには？

設定 → 拡張機能から、または `~/.renbrowser/plugins/<id>/` に解凍します。[拡張機能](extensions.md) を参照してください。

## ノードのアドレスを入力するには？

32 文字の 16 進数ハッシュ、または完全な `hash:/page/file.mu` URL を貼り付けてください。[ナビゲーションと URL](navigation-and-urls.md) を参照してください。

## ディスカバリーに何も表示されない

設定の Reticulum インターフェースと [Reticulum の設定](reticulum-setup.md) を確認してください。

## バグを報告するには？

セキュリティの問題は [セキュリティ](security.md) に従って LXMF で報告してください。コードの修正については [コントリビュート](contributing.md) を参照してください。

## プロジェクトのライセンスは？

MIT です。アドレスバーに `license:` と入力するか、[LICENSE](../../LICENSE) を読んでください。

## どのような技術スタックで構築されていますか？

Go、Wails v3、Svelte 5、SQLite、Quad4 Reticulum ライブラリです。[アーキテクチャ](architecture.md) を参照してください。

## Android で動作しますか？

リリースに APK が公開されている場合は動作します。必要に応じて Android SDK を使用してソースからビルドしてください。[インストール](installation.md) を参照してください。

## キーボードショートカットを変更するには？

設定 → キーバインドから。[キーボードショートカット](keyboard-shortcuts.md) を参照してください。
