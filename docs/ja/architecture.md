# アーキテクチャ

Ren Browser の構造の概要。

## 概要

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 5 frontend (Wails webview or browser in server) │
│  Tabs, chrome, Micron viewer, panels, settings          │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails bindings / HTTP API
┌───────────────────────▼─────────────────────────────────┐
│  internal/app : BrowserService                          │
│  Navigation, tabs, history, settings, plugin bridge     │
└───────┬─────────────┬─────────────┬─────────────────────┘
        │             │             │
        ▼             ▼             ▼
   internal/rns   internal/store  internal/plugins
   Reticulum      SQLite          Extensions
        │
        ▼
   internal/nomadnet : LXMF page fetch for announced NomadNet nodes
        │
        ▼
   internal/micron : Markup parse and HTML render
```

## エントリーポイント

| ファイル | ビルドタグ | 役割 |
|---------|-----------|------|
| `main_desktop.go` | `!server && !android` | Wails ウィンドウ、埋め込み `frontend/dist` |
| `main_server.go` | `server` | HTTP サーバー、同じ埋め込みアセット |
| Android main | `android` | モバイルシェル |

`internal/bootstrap` は設定、ストア、プラグイン、Wails アプリを結合します。

## フロントエンド

- **フレームワーク：** Vite を使った Svelte 5
- **バインディング：** `frontend/bindings/renbrowser/` 以下に生成
- **メイン UI：** `frontend/src/App.svelte` がクロームとパネルをオーケストレート
- **コンポーネント：** `frontend/src/lib/components/`（タブバー、ディスカバリー、設定など）
- **ブラウザロジック：** `frontend/src/lib/browser/`（URL、キーバインド、エラー）

Micron レンダリングは SRI 検証を持つ `MicronWasmManager` が管理する WASM パーサーを使用できます。

## バックエンドサービス

### BrowserService（`internal/app`）

UI の中央 API：

- URL をナビゲートしてタブ状態を管理する
- ディスカバリー、履歴、お気に入りを公開する
- 設定を読み込み・保存する
- プラグインホストへのブリッジ

### Reticulum スタック（`internal/rns`）

`quad4/reticulum-go` をラップします：

- トランスポートの開始と停止
- インターフェース統計の報告
- 設定からのホットリロード設定

### ページフェッチ（`internal/nomadnet`）

LXMF と Reticulum 経由でリモートの `.mu` および関連コンテンツをフェッチします。アナウンスが一致した場合、ディスカバリーはノードを NomadNet としてラベル付けします。Ren Browser は NomadNet クライアントライブラリを使用しません。

### ストア（`internal/store` + `internal/db`）

レガシー `state.json` からの移行を含む SQLite 永続化。

### プラグイン（`internal/plugins`）

- マニフェスト検証
- 権限の強制
- 組み込みスキーム（`about:`、`license:`、`editor:`）
- JS および WASM プラグインランタイム

## コンテンツとレンダリング

| パッケージ | 役割 |
|-----------|------|
| `internal/content` | 静的ページ（about、license） |
| `internal/micron` | Micron から HTML へ |
| `internal/micronwasm` | WASM パーサー統合 |
| `internal/cache` | ページキャッシュヘルパー |

## サーバーミドルウェア

`internal/servermw` はサーバーモードでベースパスヘッダーとプロキシ対応の URL 構築を処理します。

## 設定

`internal/config` はフラグ、`.env`、`REN_BROWSER_*` 変数をブートストラップが使用する `Runtime` 構造体にパースします。

## ブランドとパス

`internal/brand`（`build/brand.yml` から生成）は安定した名前を定義します：

- データディレクトリ `.renbrowser`
- DB ファイル `renbrowser.db`
- 表示名とバージョンラベル

## ビルドとパッケージング

- `Taskfile.yml`：開発者コマンド
- `build/`：OS 別パッケージング（Linux AppImage、Windows NSIS、macOS、Android、Docker）
- `build/config.yml`：Wails プロジェクト設定

## CI

GitHub Actions は Go テスト、フロントエンドチェック、セキュリティスキャン、デスクトップおよびサーバーのスモークビルド、リリースアーティファクトを実行します。`.github/workflows/` を参照してください。

## 拡張ポイント

1. **プラグイン**：マニフェスト駆動の UI とスキーム
2. **テーマ**：JSON トークンファイル
3. **コミュニティインターフェース**：設定内の Reticulum 設定スニペット

## 次のステップ

- ローカルビルドは [開発](development.md)
- プラグイン API の表面は [拡張機能](extensions.md)
- リポジトリの `README.md` レイアウトテーブルのソースツリー
