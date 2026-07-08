# インストール

このページでは、ビルド済みダウンロード、Docker、ソースからのビルドについて説明します。

## ビルド済みダウンロード（推奨）

[GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) からお使いのシステム向けの最新リリースを入手してください。

| プラットフォーム | ファイル | 備考 |
|----------------|---------|------|
| Linux x86_64 | `renbrowser-linux-amd64.AppImage` | `chmod +x` してから実行。プレーンバイナリも同梱。 |
| Linux ARM64 | `renbrowser-linux-arm64.AppImage` | x86_64 と同じ手順。 |
| Windows | `renbrowser-windows-amd64.exe` | そのまま実行。インストーラー不要。 |
| macOS | `renbrowser-macos-universal.zip` | 解凍して `renbrowser.app` を開く。 |
| サーバー（Linux x86_64） | `renbrowser-server-linux-amd64` | Docker またはセルフホスティング用ヘッドレスバイナリ。 |
| サーバー（Linux ARM64） | `renbrowser-server-linux-arm64` | Raspberry Pi 3/4/5 および他の 64 ビット ARM ボード。 |
| サーバー（Linux ARMv6） | `renbrowser-server-linux-armv6` | Raspberry Pi Zero W および他の 32 ビット ARMv6 デバイス。 |
| サーバー（FreeBSD） | `renbrowser-server-freebsd-amd64`、`renbrowser-server-freebsd-arm64` | FreeBSD でのヘッドレス動作。 |
| サーバー（OpenBSD / NetBSD） | `renbrowser-server-openbsd-amd64`、`renbrowser-server-netbsd-amd64` | BSD でのヘッドレス動作。 |
| Android | `renbrowser.apk` | リリースパイプラインに含まれている場合。 |

各リリースにはダウンロードを検証するための `SHA256SUMS.txt` が含まれています。[セキュリティ](security.md)を参照してください。

### ダウンロードの検証（Linux または macOS）

```sh
sha256sum -c SHA256SUMS.txt
```

チェックサムファイルに多くのアセットが含まれている場合は、ダウンロードしたファイルのみを確認してください。

### システム要件

| パッケージ | ホストに必要なもの |
|-----------|-----------------|
| **Linux AppImage** | GTK 4、WebKitGTK 6、その他のライブラリをバンドル。WebKit の個別インストール不要。一部のディストリビューションでは FUSE または `APPIMAGE_EXTRACT_AND_RUN=1` が必要。 |
| **Linux Flatpak** | Flatpak と `org.gnome.Platform` ランタイム（GTK 4 および WebKitGTK 6）。 |
| **Linux プレーンバイナリ** | 実行時に GTK 4 と WebKitGTK 6.0 が必要（例：Debian/Ubuntu 24.04 以降、Fedora、Arch）。 |
| **Windows `.exe`** | [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/)。通常 Windows 10/11 に含まれる。NSIS インストーラーでインストール可能。ポータブル `.exe` には含まれない。 |
| **macOS `.app`** | システム WebKit を使用した最新の macOS（追加ランタイム不要）。 |
| **Android APK** | Android 5.0 以降（API 21 以降）。 |
| **サーバーバイナリ / Docker** | デスクトップ GUI スタック不要。UI には任意のブラウザを使用。リリースサーバービルド：Linux amd64/arm64/armv6、FreeBSD amd64/arm64、OpenBSD/NetBSD amd64。 |

## Docker または Podman（サーバーモード）

公式イメージ：`ghcr.io/quad4-software/renbrowser`

コンテナがメッシュに参加できるよう、Reticulum の設定とプロファイルデータをマウントします。イメージは非 root ユーザーで実行されるため、ホストの UID/GID を渡してください：

```sh
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

同じフラグは `podman run` でも動作します。Podman では `--user "$(id -u):$(id -g)"` の代わりに `--userns=keep-id` を使用できます。SELinux がバインドマウントをブロックする場合は、ボリュームフラグに `:Z` を追加してください。

同じマシンの任意のブラウザで `http://localhost:8080` を開きます。

このリポジトリからビルドして実行：

```sh
task build:docker
task run:docker
```

サーバーイメージには**ログイン画面がありません**。信頼できるネットワークにのみ公開してください。[サーバーモード](server-mode.md)と[セキュリティ](security.md)を参照してください。

## ソースからのビルド

コントリビューターや、ビルド済みパッケージがないプラットフォーム向けです。

### 要件

- [Go](https://go.dev/) 1.26 以降
- [Node.js](https://nodejs.org/) 22 以降と [pnpm](https://pnpm.io/) 11 以降
- [Task](https://taskfile.dev/)（推奨）
- `~/.reticulum-go/` の Reticulum 設定（または `REN_BROWSER_CONFIG` を設定）

### 基本的なビルド

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Go モジュールは GitHub から Quad4 の依存関係を自動的に取得します。

### プラットフォーム固有のビルド

```sh
task build:windows
task build:darwin
task build:android      # 実機（arm64）
task build:android:emu  # エミュレーター（ホスト ABI）
```

### インストーラーとパッケージ

```sh
task package                  # 現在の OS
task package:linux:appimage   # Linux AppImage
task package:darwin:universal # macOS ユニバーサル
task package:windows          # Windows NSIS インストーラー
```

### Android SDK

Android ビルドには [Android SDK](https://developer.android.com/studio)（API 34、NDK r26 以降）が必要です。ビルド時にツールが見つからないと報告された場合は、`ANDROID_HOME` を設定して `task android:install:deps` を実行してください。

## ソースからのサーバーバイナリ

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

環境変数とデプロイに関する注意は [サーバーモード](server-mode.md)を参照してください。

## インストール後

1. Reticulum の設定が整っていることを確認する（[Reticulum の設定](reticulum-setup.md)）
2. アプリを起動して `about:` を開く
3. [ブラウザの使い方](using-the-browser.md)を読む
