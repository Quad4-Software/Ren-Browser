# サーバーモード

サーバーモードは、デスクトップシェルなしで Ren Browser をウェブアプリとして実行します。HTTP URL で別のブラウザからアクセスします。

## サーバーモードを使う場面

- 既に Reticulum が動作しているホームラボや VPS
- Docker デプロイメント
- デスクトップアプリをインストールしたくない共有マシン
- LAN 上のタブレットやスマートフォンからのアクセス

## クイックスタート

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Firefox、Chromium、または Safari で `http://localhost:8080`（またはホスト IP）を開きます。

## Docker

公開済みイメージ：

```
ghcr.io/quad4-software/renbrowser:latest
```

実行例：

```sh
mkdir -p "$HOME/.reticulum-go" "$HOME/.renbrowser"
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

Reticulum ディレクトリをリードオンリーでマウントしないでください。メッシュは設定ファイルの隣にストレージを更新する必要があります。マウントの詳細と Podman の注意事項については [Reticulum の設定](reticulum-setup.md#server-and-docker) を参照してください。

ローカルでビルド：

```sh
task build:docker
task run:docker
```

## コマンドラインフラグ

| フラグ | 用途 |
|-------|------|
| `--host` | バインドアドレス（サーバービルドのデフォルトは `0.0.0.0`） |
| `--port` | HTTP ポート（デフォルトは `8080`） |
| `--config` | Reticulum 設定のパス |
| `--trust-proxy` | リバースプロキシからの `X-Forwarded-*` を信頼する |
| `--base-path` | サブパスで提供される場合の URL プレフィックス |
| `--public-mode` | お気に入り、履歴、タブをサーバーの SQLite の代わりにブラウザの `localStorage` に保存する |
| `--profile` | 名前付きプロファイルデータベース |
| `--import-profile` / `--export-profile` | 起動時のプロファイル JSON |

## 環境変数

サーバーは作業ディレクトリの `.env` ファイルを読み込みます。環境に既に設定されている変数は上書きされません。

| 変数 | 用途 |
|-----|------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | バインドアドレス |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | ポート |
| `REN_BROWSER_CONFIG` / `RETICULUM_CONFIG` | Reticulum 設定 |
| `REN_BROWSER_TRUST_PROXY` | `true` / `1` / `yes` でトラストプロキシを有効化 |
| `REN_BROWSER_BASE_PATH` | サブパスプレフィックス |
| `REN_BROWSER_PUBLIC_MODE` | 公開モードの切り替え |
| `REN_BROWSER_PROFILE` | プロファイル名 |
| `REN_BROWSER_IMPORT_PROFILE` | 起動時のインポートパス |
| `REN_BROWSER_EXPORT_PROFILE` | 起動時のエクスポートパス |

## 公開モード

`--public-mode` なしの場合、サーバーはタブ、履歴、お気に入りをサーバーディスク上の SQLite データベースに保持します。そのインスタンスを共有するすべてのクライアントが同じデータを見ます。

`--public-mode` ありの場合、これらのアイテムは各ブラウザの `localStorage` に保存されます。多くのユーザーが 1 つのサーバーを使用し、プロファイルを共有すべきでない場合に使用してください。

## リバースプロキシ

一般的な nginx または Caddy のセットアップ：

1. プロキシで TLS を終端する
2. `127.0.0.1:8080` にプロキシする
3. `X-Forwarded-Proto` と `X-Forwarded-Host` を渡す
4. `--trust-proxy` で Ren Browser を起動する
5. アプリがドメインルートにない場合は `--base-path` を設定する

トラストプロキシが有効な場合、ヘッダー `X-RenBrowser-Base-Path` が認識されます。

## 組み込み認証なし

HTTP ポートに到達できる人は誰でもブラウザを使用し Reticulum トラフィックをトリガーできます。以下なしに公開インターネットにポート 8080 を公開しないでください：

- ファイアウォールルール
- VPN
- 認証付きリバースプロキシ
- またはこれらすべて

サーバーを公開する前に [セキュリティ](security.md) を読んでください。

## アセットオーバーライド（上級者向け）

開発用に、埋め込みアセットの代わりにディスクまたは zip からフロントエンドファイルを提供できます：

- `--assets-dir path`
- `--assets-zip path`

環境変数：`REN_BROWSER_ASSETS_DIR`、`REN_BROWSER_ASSETS_ZIP`。

## 次のステップ

- SQLite と公開モードは [データとプロファイル](data-and-profiles.md)
- デプロイのセキュリティ強化は [セキュリティ](security.md)
- リリースバイナリは [インストール](installation.md)
