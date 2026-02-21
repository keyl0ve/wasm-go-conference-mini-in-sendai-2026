# グローバル関数公開デモ（理想的な例）

`js.Global().Set()` で Go 関数を1つだけグローバルに公開し、HTML の onclick から呼び出す例。addEventListener を Go で使わないため、リスナーが増え続けることがありません。

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

1. 「Click Me」ボタンをクリック → 画面上に "Clicked!" と表示
2. 「Increment」ボタンを何度もクリック → カウンターが増える
3. 何度実行してもメモリリークしない（js.FuncOf を使っていないため）

## この方法のメリット

- ✅ **js.FuncOf が不要** → Release() 忘れによるメモリリークの心配がない
- ✅ **addEventListener を Go で使わない** → リスナーが増え続けることがない
- ✅ **コードがシンプル** → main() で登録処理が不要
- ✅ **パフォーマンスが良い** → JS ↔ WASM の境界越えが最適化されやすい

## 比較

| アプローチ                              | メリット               | デメリット                   |
| --------------------------------------- | ---------------------- | ---------------------------- |
| addEventListener + js.FuncOf (悪い例)   | -                      | Release() 忘れでメモリリーク |
| removeEventListener + Release (良い例)  | リークしない           | 管理が複雑                   |
| **js.Global().Set() + onclick（推奨）** | **管理不要、シンプル** | -                            |
