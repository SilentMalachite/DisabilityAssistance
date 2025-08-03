# アーキテクチャと構造

## クリーンアーキテクチャ採用
```
internal/
├── domain/          # ドメインモデルと業務ルール
├── usecase/         # ビジネスロジック層
└── adapter/         # インフラストラクチャ層
    ├── db/          # データベース層
    ├── crypto/      # 暗号化層
    ├── pdf/         # PDF生成層
    ├── audit/       # 監査ログ層
    └── session/     # セッション管理層
```

## 主要なドメインモデル
- **Staff**: 職員管理（admin/staff/readonly ロール）
- **Recipient**: 利用者情報（暗号化フィールド付き）
- **BenefitCertificate**: 受給者証管理
- **StaffAssignment**: 担当者割り当て
- **AuditLog**: 監査ログ

## ディレクトリ構造
```
shien-system/
├── cmd/desktop/              # GUIアプリケーション
├── internal/
│   ├── domain/              # ドメインモデル
│   ├── usecase/             # ビジネスロジック
│   ├── adapter/             # インフラストラクチャ
│   └── ui/                  # ユーザーインターフェース
│       ├── theme/           # 日本語テーマ
│       ├── widgets/         # カスタムウィジェット
│       └── views/           # 画面コンポーネント
├── migrations/              # データベーススキーマ
├── testdata/               # テスト用データ
└── assets/                 # 静的ファイル
```

## セキュリティ設計
- **暗号化**: フィールドレベルでの自動暗号化/復号化
- **アクセス制御**: ロールベース認可
- **監査**: 全操作の不変ログ記録
- **セッション管理**: メモリベースのセキュアセッション