# 障害者サービス管理システム

## 🎯 プロジェクト概要

日本の障害者福祉サービス事業所向けの**セキュアな利用者情報管理システム**です。個人情報保護とアクセシビリティを最重視した、完全オフライン動作のデスクトップアプリケーションです。

### 主な特徴

- 🔒 **最高レベルのセキュリティ**: AES-256-GCM暗号化による機微情報保護
- 🇯🇵 **日本語完全対応**: 日本の福祉制度に特化した設計
- ♿ **アクセシビリティ重視**: 障害者が利用しやすいUI/UX設計
- 💾 **完全オフライン**: インターネット接続不要、データは手元で安全管理
- 🖥️ **クロスプラットフォーム**: Windows・macOS両対応
- 📋 **監査ログ完備**: 全操作を記録、コンプライアンス対応

## 🏗️ システム構成

```
障害者サービス管理システム
├── 利用者管理     - 個人情報の暗号化保存・管理
├── 受給者証管理   - 行政発行証明書の期限管理
├── 職員管理       - ロールベースアクセス制御
├── 監査ログ       - 全操作の追跡記録
├── PDF帳票出力    - 各種報告書の生成
└── バックアップ   - 暗号化データのバックアップ
```

## 🚀 クイックスタート

### 前提条件

- **Go 1.21+** (開発・ビルド用)
- **SQLite3** (データベース)
- **Fyne UI toolkit** (GUI)

### インストール

```bash
# リポジトリのクローン
git clone https://github.com/your-org/DisabilityAssistance.git
cd DisabilityAssistance

# 依存関係のダウンロード
go mod download

# アプリケーションの起動
go run ./cmd/desktop
```

### 初期セットアップ

1. **管理者アカウント作成**: 初回起動時に管理者情報を設定
2. **暗号化キー生成**: システムが自動的にセキュアな暗号化キーを生成
3. **データベース初期化**: SQLiteデータベースが自動作成されます

## 🔐 セキュリティ仕様

### データ保護

- **フィールドレベル暗号化**: 氏名、住所等の機微情報をAES-256-GCMで暗号化
- **鍵管理**: OS固有のセキュアストレージ（macOS Keychain/Windows DPAPI）
- **アクセス制御**: ロールベース認可（管理者・職員・閲覧専用）
- **監査ログ**: 全データアクセスの完全な追跡記録

### 脆弱性対策

- ✅ **SQLインジェクション防御**: パラメータ化クエリ + キーワード検証
- ✅ **XSS防御**: 出力エスケープ + 入力サニタイゼーション  
- ✅ **CSRF防御**: トークンベース認証
- ✅ **入力検証**: 日本語対応の包括的バリデーション
- ✅ **メモリ保護**: 機微データの適切なクリア

## 📊 利用者管理機能

### 基本情報管理
- 氏名（漢字・カナ対応）
- 生年月日・性別
- 住所・連絡先
- 障害情報・等級

### 受給者証管理
- 有効期限の自動アラート
- サービス種別ごとの管理
- 給付日数の追跡

### 担当者割り当て
- 利用者ごとの担当職員設定
- 引き継ぎ履歴の管理

## 🛠️ 開発・デプロイ

### ビルド

```bash
# 開発ビルド
go build ./cmd/desktop

# 本番ビルド（最適化有効）
go build -ldflags="-w -s" ./cmd/desktop

# Windows向けクロスビルド（macOSから）
GOOS=windows GOARCH=amd64 go build ./cmd/desktop
```

### テスト実行

```bash
# 全テストの実行
go test ./...

# 特定パッケージのテスト
go test ./internal/validation/...

# カバレッジ付きテスト
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### デプロイ準備

1. **設定ファイル確認**: `config/` ディレクトリの設定
2. **データベース初期化**: マイグレーションスクリプトの実行
3. **暗号化キー生成**: 本番環境での安全なキー管理
4. **バックアップ設定**: 定期バックアップの自動化

## 📁 プロジェクト構造

```
DisabilityAssistance/
├── cmd/desktop/           # メインアプリケーション
├── internal/
│   ├── domain/           # ビジネスロジック・エンティティ
│   ├── usecase/          # アプリケーションロジック
│   ├── adapter/          # 外部インターフェース
│   │   ├── db/          # データベースアクセス
│   │   ├── crypto/      # 暗号化機能
│   │   ├── pdf/         # PDF生成
│   │   └── backup/      # バックアップ機能
│   ├── ui/              # ユーザーインターフェース
│   │   └── widgets/     # GUI コンポーネント
│   └── validation/      # 入力検証システム
├── migrations/           # データベースマイグレーション
├── docs/                # プロジェクトドキュメント
├── testdata/            # テスト用データ
└── config/              # 設定ファイル
```

## 🎯 使用技術

### バックエンド
- **言語**: Go 1.21+
- **データベース**: SQLite3 (ファイルベース)
- **暗号化**: AES-256-GCM (Go標準crypto)
- **PDF生成**: gofpdf

### フロントエンド
- **GUI フレームワーク**: Fyne v2
- **テーマ**: カスタム日本語対応テーマ
- **フォント**: Noto Sans CJK (埋め込み)

### セキュリティ
- **ハッシュ化**: bcrypt (パスワード)
- **鍵管理**: OS Keychain/DPAPI
- **監査**: 構造化ログ (JSON形式)

## 📋 ライセンス

このプロジェクトは **MIT License** の下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🤝 コントリビューション

プロジェクトへの貢献を歓迎します！

1. **Fork** このリポジトリ
2. **ブランチ作成** (`git checkout -b feature/amazing-feature`)
3. **コミット** (`git commit -m 'Add amazing feature'`)
4. **Push** (`git push origin feature/amazing-feature`)
5. **Pull Request** の作成

### 開発ガイドライン

- **コードスタイル**: `gofmt` と `golint` に準拠
- **テスト**: 新機能には必ずテストを追加
- **セキュリティ**: セキュリティに関わる変更は慎重にレビュー
- **ドキュメント**: 機能追加時はドキュメントも更新

詳細は [CONTRIBUTING.md](docs/CONTRIBUTING.md) を参照してください。

## 📞 サポート

### 技術サポート
- **Issue**: [GitHub Issues](https://github.com/your-org/DisabilityAssistance/issues)
- **Wiki**: [プロジェクトWiki](https://github.com/your-org/DisabilityAssistance/wiki)

### セキュリティ報告
セキュリティ脆弱性を発見した場合は、公開のIssueではなく直接ご連絡ください：
📧 security@your-org.com

## 📈 ロードマップ

### Phase 1 (完了)
- ✅ 基本的な利用者管理機能
- ✅ セキュリティ基盤の構築
- ✅ PDF出力機能
- ✅ 包括的な入力検証

### Phase 2 (進行中)
- 🔄 高度な検索・フィルタ機能
- 🔄 データ移行ツール
- 🔄 モバイル対応検討

### Phase 3 (計画中)
- 📋 レポート機能の拡充
- 📋 API連携機能
- 📋 クラウドバックアップ連携

---

**障害者福祉サービスの現場で実際に使用されることを想定し、使いやすさとセキュリティを両立した設計を心がけています。**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Security](https://img.shields.io/badge/Security-AES--256--GCM-green.svg)](https://en.wikipedia.org/wiki/Galois/Counter_Mode)