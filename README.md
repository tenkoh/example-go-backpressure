## 概要

[go-secure-sse-server.md](./go-secure-sse-server.md)に記載したSSE実験用のコードです。Go 1.24以上で動作確認できます。

- `cmd/sse-server`: 基準ケースのSSEサーバー。
- `cmd/sse-improved-server`: 書き込みごとにタイムアウトを設定した改善版SSEサーバー。
- `cmd/slow-client`: レスポンスの読み取りが遅いクライアント。

実験の背景や詳細な解説は技術文書を参照してください。

## 使い方

### サーバーの起動

- 待ち受け: `:8080` 固定
- トークン生成間隔: 50ms 固定
- チャネルバッファ: 100 固定

```bash
# 基準ケース
go run ./cmd/sse-server -token-bytes=50000

# 改善版 (1秒の書き込みタイムアウト)
go run ./cmd/sse-improved-server -token-bytes=50000 -write-timeout=1s
```

主なフラグ:
- `-token-bytes`: 1イベントあたりのペイロードサイズ
- `-write-timeout` (改善版のみ): 各書き込みのタイムアウト

### 遅延クライアントの実行

デフォルトで`http://localhost:8080/events`に接続します。

```bash
# 1イベントごとに10秒待って読み取る例
go run ./cmd/slow-client -delay=10s
```

主なフラグ:
- `-delay`: イベント読み取り後の遅延
- `-limit`: 読み取るイベント数の上限 (0で無制限)
- `-request-timeout`: リクエスト全体のタイムアウト

複数クライアントを並列に実行する場合の例 (シェルの`for`を利用):

```bash
for i in $(seq 1 50); do
  go run ./cmd/slow-client -delay=10s &
done
wait
```

### 計測のヒント

- サーバープロセスのメモリー使用量は`top`で観察できます。
- TCP送信バッファの状態確認例:

```bash
watch -n 0.5 "ss -tm state established '( sport = :8080 )'"
```
