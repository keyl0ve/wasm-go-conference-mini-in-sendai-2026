# Panic デモ（良い例）

defer + recover で panic を捕捉する例。エラーを JS 側に通知し、WASM は継続して動作します。

## ビルド

```bash
GOOS=js GOARCH=wasm go build -o main.wasm main.go
```

## 実行

簡易 HTTP サーバーで配信してください（file:// では fetch が失敗するため）。

```bash
# Python 3
python3 -m http.server 8080
```

ブラウザで http://localhost:8080 を開き、index.html を表示します。

## 動作確認手順

1. 「Trigger Panic」をクリック → 画面上に `recovered: ...` が表示される
2. 「Ping」をクリック → 「WASM is alive!」と表示される（WASM は生きている）
3. 「Trigger Panic」を何度でも押せる（recover により継続動作）
