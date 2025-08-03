# 日本語フォント配置

このディレクトリには、PDF生成用の日本語フォントファイルを配置してください。

## 必要なフォントファイル

- `NotoSansCJK-Regular.ttf` - 通常の日本語フォント
- `NotoSansCJK-Bold.ttf` - 太字の日本語フォント

## フォントの入手方法

1. Google Fonts から Noto Sans CJK をダウンロード
   - https://fonts.google.com/noto/specimen/Noto+Sans+JP
   
2. または Adobe からダウンロード
   - https://github.com/adobe-fonts/source-han-sans

## ライセンス

Noto Sans CJK は SIL Open Font License 1.1 でライセンスされています。
商用利用も可能です。

## 実装ノート

- フォントファイルがない場合は、デフォルトフォント（Arial）にフォールバック
- 日本語テキストは自動的に適切なフォントで表示される
- PDF生成時に動的にフォントを読み込み