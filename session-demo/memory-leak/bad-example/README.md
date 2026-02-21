# js.Value 参照保持デモ（悪い例）

js.FuncOf で作ったコールバックを Release() せずに addEventListener で登録し続ける例。メモリリークが発生します。

## ビルド

```bash
GOOS=js GOARCH=wasm go build -o main.wasm main.go
```

## 実行

```bash
python3 -m http.server 8080
```

ブラウザで http://localhost:8080 を開き、index.html を表示します。

## 動作確認手順

1. 「Add Listener 100回」を2〜3回押す
2. 「Test Click」を押す → コンソールに登録した数だけ "listener fired" が出力される（重複登録の証拠）
3. メモリリークの確認（下記）

## メモリプロファイラでの確認

1. Chrome でページを開き、F12 → Memory タブ
2. 「Heap snapshot」を選択し、スナップショットを1回取る（基準）
3. 「Add Listener 100回」を数回押す
4. もう一度 Heap snapshot を取る
5. 2つのスナップショットを比較するか、ヒープサイズが増えていることを確認
6. 悪い例では、Release() していないため js.Func が解放されずヒープが増え続ける
