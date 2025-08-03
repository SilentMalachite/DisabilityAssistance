# 開発コマンド一覧

## 基本開発コマンド

### ビルドと実行
```bash
# プロジェクトのビルド
go build ./...

# デスクトップアプリの実行
go run ./cmd/desktop

# テストビルド（macOS向け）
CGO_ENABLED=1 go build -o test_build ./cmd/desktop
```

### テスト関連
```bash
# 全テスト実行
go test ./...

# 詳細テスト実行
go test ./... -v

# 特定パッケージのテスト
go test ./internal/domain -v
go test ./internal/adapter/db -v
go test ./internal/usecase -v

# カバレッジ付きテスト
go test ./... -cover

# ベンチマークテスト
go test ./... -bench=.
```

### 依存関係管理
```bash
# 依存関係の更新
go mod tidy

# 依存関係のダウンロード
go mod download

# セキュリティ脆弱性チェック
go mod audit
```

### データベース関連
```bash
# データベースの初期化（アプリ起動時に自動実行）
# マイグレーション: migrations/0001_init.sql が自動適用される

# データベースファイルの確認
ls -la *.db*
```

### 開発支援
```bash
# Goのフォーマット
go fmt ./...

# 静的解析
go vet ./...

# インポートの整理（必要に応じて）
goimports -w .
```

## トラブルシューティング

### macOS固有の問題
```bash
# CGOエラーの場合
export CGO_ENABLED=1

# Xcode Command Line Toolsが必要
xcode-select --install
```

### データベースリセット（開発時のみ）
```bash
# データベースファイルの削除（注意：全データが失われます）
rm -f shien-system.db*
```