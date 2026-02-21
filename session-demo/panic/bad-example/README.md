# Panic デモ（悪い例）

recover なしで panic を発生させる例。panic が起きると WASM 全体が止まり、ページをリロードするまで復帰しません。

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

1. 「Trigger Panic」をクリック → ブラウザのコンソールにスタックトレースが表示される
2. 「Ping」をクリック → 反応しない（WASM が停止している）
3. ページをリロードするまで WASM は復帰しない
