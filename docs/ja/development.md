# 開発

Ren Browser をローカルで開発する方法。

## 要件

- Go 1.26 以降
- Node.js 22 以降と pnpm 11 以降
- Task（推奨）
- ライブメッシュテスト用の Reticulum 設定（オプション）

## クローンと実行

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task dev
```

`task dev` はデフォルトでポート 9245 上で Vite ホットリロードを使用して Wails を開発モードで実行します。

## よく使うタスク

| タスク | 用途 |
|-------|------|
| `task dev` | デスクトップ開発モード |
| `task build` | 現在の OS の本番ビルド |
| `task run` | ビルド済みバイナリを実行 |
| `task check` | 完全な品質ゲート（Go＋フロントエンド） |
| `task test` | すべてのユニットテスト |
| `task test:interop` | ライブ Reticulum テスト（ネットワーク必要） |
| `task build:server` | ヘッドレスサーバーバイナリ |
| `task run:server` | サーバーをローカルで実行 |
| `task package` | プラットフォームインストーラーまたはバンドル |

## 品質ゲート

変更を送信する前に：

```sh
task check
```

`check` は以下を実行します：

- Go ソースの `gofmt`
- `go test ./...`
- gosec セキュリティスキャン
- ブランド一貫性チェック
- フロントエンドの型チェック、lint、フォーマットチェック、knip、audit、vitest

オプションのより厳しい Go テスト：

```sh
task test:go:race
task test:go:hard
task fuzz:go
```

## フロントエンドのみ

```sh
cd frontend
pnpm install
pnpm check
pnpm test
```

`frontend/bindings/` 以下のバインディングは Wails によって Go サービスから生成されます。

## Go レイアウト

| パス | 役割 |
|-----|------|
| `main_desktop.go` | デスクトップエントリー（Wails ウィンドウ） |
| `main_server.go` | サーバーエントリー（HTTP のみ） |
| `internal/app/` | UI に公開されるブラウザサービス |
| `internal/rns/` | Reticulum スタックラッパー |
| `internal/nomadnet/` | LXMF 経由のページフェッチ。NomadNet ノードはアナウンスから検出され、NomadNet ライブラリは使用しない |
| `internal/micron/` | Micron レンダリング |
| `internal/store/` | SQLite 永続化 API |
| `internal/plugins/` | 拡張機能ホスト |
| `frontend/` | Svelte 5 UI |

詳細なマップは [アーキテクチャ](architecture.md) を参照してください。

## プラグイン開発

```sh
task test:plugins
```

開発中のプラグインを `~/.renbrowser/plugins/` にインストールしてアプリを再読み込みします。

## インターオップテスト

```sh
task test:interop
```

ライブの Reticulum ネットワークと `interop` ビルドタグが必要です。UI のみの作業ではスキップしてください。

## アセットプローブ

どのアセットローダー（埋め込み vs ディスク）がアクティブかデバッグする際は `REN_BROWSER_ASSET_PROBE=1` を設定します。

## SPDX ヘッダー

新しい Go ファイルには以下を含める必要があります：

```go
// SPDX-License-Identifier: MIT
```

隣接するファイルの既存のスタイルに合わせてください。

## 次のステップ

- [アーキテクチャ](architecture.md)
- [コントリビュート](contributing.md)
- [拡張機能](extensions.md)
