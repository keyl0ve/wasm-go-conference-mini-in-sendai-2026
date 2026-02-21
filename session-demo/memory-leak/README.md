# js.Value 参照保持デモ

section5 落とし穴のスライド3「js.Value の参照保持問題」を実際に確認するデモです。

- **bad-example/** … js.FuncOf で作ったコールバックを Release() せずに登録し続ける → メモリリーク
- **good-example/** … removeEventListener + Release() で適切に解放する
- **global-function-example/** … グローバル関数公開 + onclick の理想的な実装。addEventListener を使わないためリスナーが増えない

各ディレクトリの README にビルド・実行・メモリ確認方法を記載しています。

3つのアプローチの比較
bad-example good-example wasmexport-example
js.FuncOf 使う（Releaseなし） 使う 使わない
Release 管理 ❌ なし → リーク ✅ あり ✅ 不要
コードの複雑さ シンプルだがバグ やや複雑 最もシンプル
メモリリーク ❌ あり ✅ なし ✅ なし
Go バージョン 任意 任意 1.24+
