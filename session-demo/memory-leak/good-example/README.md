# js.Value 参照保持デモ（良い例）

① 前回のリスナーを removeEventListener で外してから Release() し、新しいリスナーだけを登録する例。
② onclick で Go 側の関数を1つだけグローバルに公開し、addEventListener を Go で使わない例。

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

1. **① Replace Listener**: 「Replace Listener」を何度か押す → 「Test Click」を押すとコンソールに1回だけログが出る（常に1リスナーのみ）
2. **② onclick**: 「Click（onclick で Go を呼ぶ）」を押す → 画面上に "clicked!" と表示される
3. メモリプロファイラで Heap snapshot を繰り返し取っても、Replace を繰り返してもヒープが増え続けないことを確認できる
