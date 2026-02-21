# Panic 挙動デモ

Go WASM で panic が発生したときの挙動を、悪い例・良い例で比較するデモです（section5 落とし穴のスライド用）。

- **bad-example/** … recover なし。panic で WASM 全体が止まる
- **good-example/** … defer + recover で捕捉し、継続動作

各ディレクトリの README にビルド・実行方法を記載しています。
