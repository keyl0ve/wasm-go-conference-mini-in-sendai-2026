# LGTM 画像ジェネレーター (Go + WASM)

画像の中央にテキスト（LGTM など）を描画して JPEG でダウンロードする Web アプリです。

## ビルド・実行

```bash
cd wasm-image
GOOS=js GOARCH=wasm go build -o main.wasm .
# wasm_exec.js は Go 1.21+ では $(go env GOROOT)/lib/wasm にあります
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
go run server.go
```

ブラウザで http://localhost:8080 を開いてください。
