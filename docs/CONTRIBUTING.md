# 開発者向けガイド

## 🚀 開発環境のセットアップ

このガイドでは、障害者サービス管理システムの開発環境構築から、コントリビューションまでの手順を説明します。

### 前提条件

#### 必須ソフトウェア
- **Go 1.21+**: [公式サイト](https://golang.org/dl/)からダウンロード
- **Git**: バージョン管理
- **VS Code** または **GoLand**: 推奨IDE

#### 推奨ツール
- **golangci-lint**: 静的解析
- **gofumpt**: コードフォーマッター
- **gotests**: テスト生成
- **air**: ホットリロード

### 開発環境構築

#### 1. リポジトリのセットアップ

```bash
# リポジトリのフォーク
# GitHubでフォークボタンをクリック

# クローン (フォークしたリポジトリから)
git clone https://github.com/YOUR_USERNAME/DisabilityAssistance.git
cd DisabilityAssistance

# アップストリームの設定
git remote add upstream https://github.com/original-org/DisabilityAssistance.git
```

#### 2. 依存関係のインストール

```bash
# Go モジュールのダウンロード
go mod download

# 開発ツールのインストール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
go install github.com/cweill/gotests/gotests@latest
go install github.com/cosmtrek/air@latest
```

#### 3. 開発用データベースの準備

```bash
# テスト用データベースの作成
export SHIEN_DB_PATH="./test-dev.sqlite"
go run ./cmd/desktop
```

#### 4. 開発サーバーの起動（ホットリロード）

```bash
# air設定ファイルの作成
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/desktop"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
EOF

# 開発サーバー起動
air
```

### コーディング規約

#### Goコーディングスタイル

```go
// ✅ 良い例
func (r *RecipientRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Recipient, error) {
    // 入力検証
    if id == "" {
        return nil, fmt.Errorf("recipient ID is required")
    }
    
    // データベースアクセス
    query := "SELECT id, name_cipher, email_cipher FROM recipients WHERE id = ?"
    row := r.db.QueryRowContext(ctx, query, id)
    
    var recipient domain.Recipient
    if err := row.Scan(&recipient.ID, &encryptedName, &encryptedEmail); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domain.ErrRecipientNotFound
        }
        return nil, fmt.Errorf("failed to get recipient: %w", err)
    }
    
    return &recipient, nil
}

// ❌ 悪い例
func (r *RecipientRepository) GetByID(id string) *domain.Recipient {
    // エラーハンドリングなし、型安全性なし
    row := r.db.QueryRow("SELECT * FROM recipients WHERE id = '" + id + "'")
    // ... SQLインジェクション脆弱性
}
```

#### 命名規約

```go
// ✅ 適切な命名
type RecipientUseCase interface {
    CreateRecipient(ctx context.Context, req CreateRecipientRequest) (*domain.Recipient, error)
    GetRecipientByID(ctx context.Context, id domain.ID) (*domain.Recipient, error)
    UpdateRecipient(ctx context.Context, req UpdateRecipientRequest) (*domain.Recipient, error)
    DeleteRecipient(ctx context.Context, id domain.ID) error
}

// 日本語コメントは適切に使用
type CreateRecipientRequest struct {
    Name         string    `json:"name"`          // 氏名
    Kana         string    `json:"kana"`          // フリガナ  
    BirthDate    time.Time `json:"birth_date"`    // 生年月日
    Sex          Sex       `json:"sex"`           // 性別
    ActorID      ID        `json:"actor_id"`      // 操作者ID（監査用）
}
```

### テスト指針

#### テストの分類

1. **単体テスト**: 個別の関数・メソッドのテスト
2. **統合テスト**: 複数コンポーネント間の連携テスト
3. **エンドツーエンドテスト**: アプリケーション全体のテスト

#### テスト作成例

```go
func TestRecipientRepository_Create(t *testing.T) {
    // テストデータベースのセットアップ
    db := setupTestDB(t)
    defer db.Close()
    
    repo, err := NewRecipientRepository(db)
    require.NoError(t, err)
    
    tests := []struct {
        name        string
        recipient   *domain.Recipient
        wantErr     bool
        errContains string
    }{
        {
            name: "正常ケース",
            recipient: &domain.Recipient{
                Name:      "田中太郎",
                Kana:      "タナカタロウ",
                BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
                Sex:       domain.SexMale,
            },
            wantErr: false,
        },
        {
            name: "必須項目不足",
            recipient: &domain.Recipient{
                Name: "", // 必須項目が空
            },
            wantErr:     true,
            errContains: "name is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(context.Background(), tt.recipient)
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errContains)
            } else {
                require.NoError(t, err)
                assert.NotEmpty(t, tt.recipient.ID)
            }
        })
    }
}
```

#### セキュリティテスト

```go
func TestValidation_SQLInjection(t *testing.T) {
    validator := validation.NewFormValidator()
    
    sqlInjectionPayloads := []string{
        "'; DROP TABLE users; --",
        "admin' OR '1'='1",
        "UNION SELECT * FROM users",
        "'; INSERT INTO users VALUES ('hacker'); --",
    }
    
    for _, payload := range sqlInjectionPayloads {
        t.Run(fmt.Sprintf("payload_%s", payload), func(t *testing.T) {
            err := validator.ValidateNotContainSQLKeywords("input", payload)
            assert.Error(t, err, "SQLインジェクションペイロードが検出されませんでした")
        })
    }
}
```

### Pull Request ガイドライン

#### 1. ブランチ戦略

```bash
# 機能開発用ブランチ
git checkout -b feature/user-authentication
git checkout -b feature/pdf-export
git checkout -b feature/audit-logging

# バグ修正用ブランチ  
git checkout -b fix/validation-error
git checkout -b fix/memory-leak

# セキュリティ修正用ブランチ
git checkout -b security/sql-injection-fix
```

#### 2. コミットメッセージ

```bash
# ✅ 良いコミットメッセージ
git commit -m "feat: add comprehensive input validation for user forms

- Implement ValidateLoginForm with SQL injection protection
- Add XSS prevention for all text inputs
- Include Japanese character validation for names
- Add sanitization for all user inputs

Fixes #123"

# ✅ セキュリティ修正
git commit -m "security: fix SQL injection vulnerability in search

- Replace string concatenation with parameterized queries
- Add input validation for search parameters
- Update tests to cover injection attempts

Security impact: Prevents unauthorized data access"

# ❌ 悪いコミットメッセージ
git commit -m "fix stuff"
git commit -m "update code"
```

#### 3. PR作成チェックリスト

**機能追加**
- [ ] 機能仕様の明確化
- [ ] 単体テストの追加
- [ ] 統合テストの追加
- [ ] ドキュメントの更新
- [ ] セキュリティ影響の評価

**バグ修正**
- [ ] 根本原因の特定
- [ ] 再現テストケースの追加
- [ ] 修正内容の説明
- [ ] 回帰テストの実行

**セキュリティ修正**
- [ ] 脆弱性の詳細説明（非公開）
- [ ] 影響範囲の特定
- [ ] セキュリティテストの追加
- [ ] セキュリティレビューの要求

#### 4. PR テンプレート

```markdown
## 概要
<!-- この PR で何を実装/修正したかを簡潔に説明 -->

## 変更内容
- [ ] 新機能追加
- [ ] バグ修正
- [ ] パフォーマンス改善
- [ ] リファクタリング
- [ ] セキュリティ修正
- [ ] ドキュメント更新

## 詳細
<!-- 技術的な詳細、設計決定、トレードオフについて説明 -->

## テスト
<!-- 追加/更新したテストについて説明 -->
- [ ] 単体テスト追加/更新
- [ ] 統合テスト追加/更新
- [ ] 手動テスト実行済み

## セキュリティ
<!-- セキュリティに関わる変更がある場合 -->
- [ ] セキュリティ影響なし
- [ ] セキュリティレビュー必要
- [ ] 脆弱性修正

## 関連Issue
Fixes #XXX
Closes #XXX

## スクリーンショット
<!-- UI変更がある場合 -->

## レビュー観点
<!-- レビュアーに重点的に確認してほしい点 -->
```

### 開発ワークフロー

#### 1. 日常的な開発

```bash
# 最新の変更を取得
git fetch upstream
git checkout main
git merge upstream/main

# 機能ブランチの作成
git checkout -b feature/new-feature

# 開発 & テスト
make test
make lint

# コミット & プッシュ
git add .
git commit -m "feat: implement new feature"
git push origin feature/new-feature

# PR作成
# GitHub上でPull Request作成
```

#### 2. コードレビュー

**レビュアーのチェックポイント**
- [ ] 機能要件の満足
- [ ] コードの可読性・保守性
- [ ] テストの網羅性
- [ ] セキュリティの考慮
- [ ] パフォーマンスへの影響
- [ ] ドキュメントの更新

**レビュー依頼者の準備**
- [ ] 自己レビューの実施
- [ ] テストの実行確認
- [ ] 静的解析の通過確認
- [ ] セキュリティチェックの実施

### デバッグ & トラブルシューティング

#### 開発時のログ出力

```go
// 開発時のデバッグログ
import "log/slog"

func (r *RecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
    slog.Debug("Creating recipient",
        "id", recipient.ID,
        "actor_id", getActorIDFromContext(ctx),
    )
    
    // 機微情報はログに出力しない
    // ❌ slog.Debug("recipient data", "recipient", recipient)
    // ✅ slog.Debug("creating recipient", "id", recipient.ID)
}
```

#### よくある問題と解決策

**問題: テストが失敗する**
```bash
# テストデータの確認
go test -v ./internal/adapter/db/...

# 特定のテストのみ実行
go test -run TestRecipientRepository_Create ./internal/adapter/db/

# レースコンディションのチェック
go test -race ./...
```

**問題: 暗号化エラー**
```bash
# 暗号化キーの確認
export SHIEN_DB_PATH="/tmp/test.sqlite"
go run ./cmd/desktop

# テスト用の固定キーを使用
export SHIEN_ENCRYPTION_KEY="test-key-32-bytes-long-for-dev!"
```

**問題: メモリリーク**
```bash
# プロファイリングの実行
go test -memprofile=mem.prof -bench=. ./internal/...
go tool pprof mem.prof
```

### リリースプロセス

#### 1. リリース前チェック

```bash
# 全テストの実行
make test-all

# セキュリティチェック
make security-check

# 静的解析
make lint

# ビルド確認
make build-all-platforms
```

#### 2. バージョニング

セマンティックバージョニングを使用:
- **MAJOR**: 破壊的変更
- **MINOR**: 後方互換性のある機能追加
- **PATCH**: 後方互換性のあるバグ修正

```bash
# タグの作成
git tag -a v1.2.3 -m "Release version 1.2.3"
git push upstream v1.2.3
```

## 📞 サポート

### 開発に関する質問
- **Discord**: [開発者チャンネル](#)
- **GitHub Discussions**: [技術議論](#)

### バグ報告
- **GitHub Issues**: [バグ報告テンプレート](#)

### セキュリティ問題
- **Email**: security@your-org.com (非公開)

---

**プロジェクトへの貢献をお待ちしています！質問があれば遠慮なくお声がけください。**