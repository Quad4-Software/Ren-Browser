# Reticulum の設定

Ren Browser は `quad4/reticulum-go` スタックを通じて [Reticulum](https://reticulum.network/) を使用します。このページでは、アプリが何を期待しているか、よくあるメッシュの問題の解決方法を説明します。

## デフォルトの設定場所

| 項目 | デフォルトパス |
|------|--------------|
| Reticulum 設定ディレクトリ | `~/.reticulum-go/` |
| 上書きフラグ | `--config /path/to/config` |
| 上書き環境変数 | `REN_BROWSER_CONFIG` または `RETICULUM_CONFIG` |

ディレクトリ内の具体的なファイルは、Reticulum または reticulum-go のセットアップによって異なります。Ren Browser は起動時にスタックを開始し、**設定**からインターフェースの変更を再読み込みします。

## 起動時の動作

1. Ren Browser が Reticulum の設定を読み込む
2. インターフェースがオンラインになる（UDP、TCP、RNode、その他設定したもの）
3. NomadNet ノードからのアナウンスが **ディスカバリー** に表示される
4. ページリクエストが LXMF と Reticulum 経由で Micron ページをホストするノードに送られる

起動に失敗した場合は、ターミナルログ（デスクトップ）またはコンテナログ（サーバー）を確認してください。`about:` と **設定** を開くためにアプリは引き続き動作します。

## 設定内のインターフェース

**設定** を開いて Reticulum セクションを見つけます。以下のことができます：

- どのインターフェースがアクティブかを確認する
- 送受信統計を表示する
- アプリ全体を再起動せずに設定を編集してホットリロードを適用する

新しいインターフェースを追加したり、鍵を変更してブラウザに変更を素早く反映させたい場合に使用します。

## メッシュへの参加

他の Reticulum ノードへの経路が少なくとも 1 つ必要です。一般的なオプション：

- LAN 上の他の Reticulum ピアへの **ローカル UDP または TCP**
- **RNode** または同様の無線ハードウェア
- 既知のピアやハブを指す **インターフェース定義**

Reticulum の詳細はこのマニュアルの範囲外です。インターフェースの構文とアイデンティティ管理については [Reticulum マニュアル](https://reticulum.network/manual/)をお読みください。

## NomadNet のデスティネーション

NomadNet ページは Reticulum のデスティネーションに存在します。アドレスバーでは以下を使用できます：

- `abcdef0123456789abcdef0123456789:/page/index.mu` のような完全なパス
- 32 文字の裸の 16 進数ハッシュ（Ren Browser が `:/page/index.mu` を追加します）

ページは Micron マークアップ形式を使用します。Ren Browser は組み込みの Micron パイプラインでレンダリングします。

## ディスカバリーが空の場合

以下のリストを順に確認してください：

1. Reticulum が Ren Browser 内で動作していることを確認する（設定にインターフェースが表示される）
2. インターフェースがメッシュ上のピアの設定と一致していることを確認する
3. 接続後しばらく待つ。アナウンスはすぐには届かない
4. 表示されると期待しているノードと同じ論理ネットワーク上にいることを確認する

## ページがタイムアウトまたは失敗する場合

1. デスティネーションハッシュが正しいことを確認する
2. そのデスティネーションへの経路があることを確認する（ディスカバリーの可視性だけでは不十分）
3. ディスカバリーから別の既知の良いノードを試す
4. devtools またはログで LXMF またはトランスポートエラーを確認する

## サーバーと Docker

`renbrowser` Docker イメージを実行する場合は、非 root コンテナが鍵を読み取りメッシュストレージに書き込めるよう、ホストの Reticulum ディレクトリをマウントしてホストユーザーとして実行してください：

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

設定をリードオンリーでマウントしないでください。Reticulum は設定ファイルの隣にストレージを更新する必要があります。

## 次のステップ

- アナウンスされたノードのブラウジングは [ディスカバリー](discovery.md)
- アドレスバーの形式は [ナビゲーションと URL](navigation-and-urls.md)
- エラーメッセージは [トラブルシューティング](troubleshooting.md)
