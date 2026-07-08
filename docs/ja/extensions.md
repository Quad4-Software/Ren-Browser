# 拡張機能

Ren Browser は、URL スキーム、サイドバーパネル、コマンド、テーマ、設定ページ、devtools タブを追加するプラグインをサポートしています。

## 拡張機能のインストール

### 設定から

1. **設定 → 拡張機能** を開く
2. **拡張機能のインストール** を選択し、**.zip**、**フォルダー**、または**バンドルされた .wasm モジュール**を選ぶ
3. インストールプレビューを確認する：
   - 要求された権限（インストール前に個別の権限を無効化できる）
   - 拡張機能が接触する可能性のある外部 URL（マニフェストとパッケージファイルからスキャン）
   - パブリッシャーの署名ステータス（未署名、署名済み、信頼できるパブリッシャーによる署名済み、または無効）
   - セキュリティ評価の警告
   - バンドルされた UI 言語（拡張機能が `locales/*.json` を同梱している場合）
4. 確認して拡張機能を有効化する

`network.fetch` を要求する拡張機能は、検出されたエンドポイントをリストした確認ダイアログを表示します。インストール時に `network.fetch` を無効にしてもエンドポイントは表示されたままなので、その権限が付与された場合にパッケージが何に接触するかを確認できます。

### 手動インストール

プラグインを以下に解凍します：

```
~/.renbrowser/plugins/<id>/
```

フォルダーには `renbrowser.plugin.json` が含まれている必要があります。マニフェストの `id` はフォルダー名と一致する必要があります。

## サンプル拡張機能

リポジトリには `extensions/hello-extension/` が含まれています：

- `hello:` URL スキームを登録する
- **Hello** サイドバーパネルを追加する
- `mod+shift+h` のキーバインドを持つ **Say hello** コマンドを定義する

独自のプラグインを書く際のテンプレートとして使用してください。

`extensions/micron-translator/` は Google Translate（公開エンドポイント）または LibreTranslate インスタンス（サイドバーパネルで URL とオプションの API キーを設定）を使用して Micron（`.mu`）ページを翻訳します。コマンド：**Translate Micron page**（`mod+shift+t`）と **Restore original**（`mod+shift+r`）。

## マニフェストファイル

ファイル名：`renbrowser.plugin.json`

必須フィールド：

| フィールド | 用途 |
|-----------|------|
| `manifestVersion` | 現在は `1` |
| `id` | 一意の ID（`a-z`、`A-Z`、`0-9`、`.`、`-`、3 から 128 文字） |
| `name` | 表示名 |
| `version` | セマバー文字列 |
| `main` | フロントエンドエントリースクリプト（バックエンドのみの場合はオプション） |
| `permissions` | 機能リスト（以下を参照） |

オプションフィールドには `description`、`author`、`license`、`engines`、`backend`、`network`、`contributes` が含まれます。

### エンジン制約

```json
"engines": { "renbrowser": ">=0.1.0" }
```

アプリのバージョンが古すぎる場合、ホストはプラグインの読み込みを拒否します。

### ネットワークエンドポイント

`network.fetch` を使用する拡張機能は、接触するホストまたは URL を宣言する必要があります：

```json
"network": {
  "endpoints": [
    "https://api.example.com/",
    "User-configured service URL"
  ]
}
```

インストール時に RenBrowser は `.js`、`.go`、`.wasm`、その他のパッケージファイルで `http`/`https` URL をスキャンし、見つかったものをマニフェストエントリーと並べて一覧表示します。

### コントリビューション

| 種類 | 用途 |
|------|------|
| `urlSchemes` | カスタムスキームを処理する |
| `panels` | サイドバーまたは他のパネルスロット |
| `commands` | コマンドパレットエントリーとキーバインド |
| `themes` | 追加のテーマ JSON ファイル |
| `settings` | 設定サブページ |
| `devtools` | DevTools タブ |
| `renderers` | MIME タイプまたは拡張子のカスタムレンダラー |

## 権限

プラグインは必要なものを宣言する必要があります。既知の権限：

| 権限 | 許可内容 |
|------|---------|
| `storage.plugin` | プラグイン用のプライベートキーバリューストレージ |
| `navigation.read` | 現在の URL とタブ情報を読み取る |
| `navigation.write` | ナビゲーションをトリガーする |
| `network.fetch` | 許可されたネットワーク API 経由でフェッチする |
| `events.emit` | ホストイベントを発行する |
| `events.subscribe` | ホストイベントをリッスンする |
| `devtools.network` | DevTools の追加ネットワーク詳細 |
| `render.unsanitized` | 一部の HTML サニタイズをスキップする（危険） |

ホストは実行時に権限を強制します。インストール時に無効にした権限は拡張機能ごとに保存され、JS の `ctx.network.fetch` や WASM の `http_fetch` には付与されません。

## パブリッシャー署名

拡張機能は `renbrowser.plugin.rsg` に Ed25519 署名を同梱できます（Reticulum の `rnid` ツールと互換性あり）。無効な署名を持つ署名済みパッケージはインストールできません。

インストールプレビューと拡張機能リストにはバッジが表示されます：

| バッジ | 意味 |
|-------|------|
| Unsigned（未署名） | 署名ファイルなし |
| Signed（署名済み） | Reticulum アイデンティティからの有効な署名 |
| Trusted（信頼済み） | 信頼リストのパブリッシャーによる署名 |
| Tampered（改ざん済み） | RenBrowser の外部で拡張ファイルが変更された（再有効化するまで拡張機能は無効） |

インストール時に **このパブリッシャー ID を信頼する** を選択すると、有効な署名者をユーザーの信頼リスト（`~/.renbrowser/trusted_publishers.json`）に追加できます。RenBrowser には小さなバンドル済み信頼リストも含まれています。ユーザーリストはプロファイルデータベースに保存されたダイジェストで保護されており、データベースを更新せずに外部から編集すると検出されます。

`build/scripts/sign-extension.sh` でディレクトリまたは zip に署名します（Python の `rnid` が必要）。

## プラグイン UI の翻訳

拡張機能は `locales/<code>.json`（例：`locales/en.json`）以下に独自の UI 文字列をバンドルできます。パネルタイトルとコマンドはマニフェストで `%key.path%` プレースホルダーを使用できます。ホストは `/_plugins/<id>/locales/<code>.json` からカタログを読み込みます。

インストールプレビューには、存在する場合にバンドルされたロケールコードが一覧表示されます。

## フロントエンドエントリースクリプト

一般的な `main.js` はエクスポートします：

- `activate(ctx)`：イベントをサブスクライブし、UI を登録する
- `deactivate()`：クリーンアップ
- `mount(el)`：サイドバーパネルの HTML をレンダリングする
- `handleScheme(url)`：URL スキームハンドラー用

`network.fetch` を持つプラグインは、インストール時にその権限が付与された場合、公開の `http`/`https` URL への HTTP GET/POST に `ctx.network.fetch()` を呼び出せます。ネットワークバックアップの作業を開始する前に `ctx.capabilities.networkFetch` を確認してください。

`backend` WASM モジュールを持つプラグインは、`ctx.wasm.call(export, input)` を呼び出して `translate_micron` などのエクスポート関数を実行できます。Micron ソースを変換した後にアクティブなタブを再レンダリングするには、`ctx.content.getActivePage()`、`ctx.content.renderRaw(path, raw)`、`ctx.content.updateActivePage()` を使用します。

拡張機能のロケールファイルの文字列には `ctx.i18n.t("key")` を使用します。

## バンドル済み WASM モジュール

配布可能な拡張機能は 1 つの `.wasm` ファイルとして出荷できます。モジュールにはカスタムセクションが含まれます：

- `renbrowser.plugin` — マニフェスト JSON（`renbrowser.plugin.json`）
- `renbrowser.files` — 相対パスから UTF-8 ファイル内容へのマップ（例：`main.js`、`locales/en.json`）
- `renbrowser.signature` — オプションの RSG 署名バイト

**設定 → 拡張機能 → 拡張機能のインストール → .wasm モジュールを選択** からインストールします。ホストはメタデータをプラグインディレクトリに展開し、WASM バイナリをマニフェストの `backend` として保持します。

`extensions/micron-translator/` は `translator.wasm`（TinyGo）を同梱します。`extensions/micron-translator/build-wasm.sh` で再ビルドするか、ビルド後に `go run ./extensions/micron-translator/bundle` でバンドルします。

## WASM バックエンド

プラグインは重い処理のために `backend` を WASM モジュールパスに設定できます。WASM プラグインは明示的な付与を持つ制約されたランタイムで実行されます。

ホストは、インストール時に `network.fetch` が付与された場合に `http_fetch` を含む `renhost` モジュールを提供します。`translate_micron(in_ptr, in_len) -> out_len` などのエクスポート関数は JSON 入力を読み取り、線形メモリに JSON 出力を書き込みます。

セーフガードには、呼び出しごとのネットワークリクエスト制限、WASM 呼び出しタイムアウト、入力サイズキャップが含まれます。`network.fetch` が付与されていない場合、ネットワーク集約エクスポートは完全にブロックされます。

## DevTools

**開発者ツール → ネットワーク** が開いているとき、拡張機能（JS の `PluginFetch` および WASM の `http_fetch`）が行ったアウトバウンド HTTP リクエストが、ソース **Extension fetch**、ステータスコード、所要時間とともにログに表示されます。

## 整合性と改ざん

インストール後、RenBrowser は各拡張機能のファイルペイロードの暗号化ハッシュを保存します（署名ファイルを除く）。アプリの外部でファイルがディスク上で変更された場合、拡張機能は無効化され **Tampered** とマークされます。再有効化すると現在のファイルが受け入れられ、保存されたハッシュが更新されます。

## セキュリティに関する注意事項

- 信頼できるソースからのみプラグインをインストールする
- インストールを確認する前に権限と検出されたネットワークエンドポイントを読む
- 認識しているパブリッシャーからの署名済み拡張機能を優先する
- プラグインをプロファイルデータへのアクセス権を持つローカルプログラムとして扱う

## 次のステップ

- ソースリファレンス：リポジトリ内の `internal/plugins/manifest.go`
- プラグインの脅威モデルと署名については [セキュリティ](security.md)
- プラグインホストのハックには [開発](development.md)
