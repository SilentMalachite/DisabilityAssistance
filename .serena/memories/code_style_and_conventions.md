# コードスタイルと規約

## Go言語規約準拠
- Go標準のコーディング規約に従う
- `go fmt`での自動フォーマット
- `go vet`での静的解析クリア

## 命名規約

### パッケージ名
- 小文字、短縮形を使用
- 例: `domain`, `usecase`, `adapter`

### 型名
- CamelCase（先頭大文字）でエクスポート
- 例: `Staff`, `Recipient`, `BenefitCertificate`

### フィールド名
- CamelCase、JSONタグ付き
- 例: `CreatedAt time.Time \`json:"created_at"\``

### メソッド名
- CamelCase、動詞から始める
- 例: `CreateRecipient`, `GetStaffByID`

## ファイル命名
- snake_case使用
- テストファイル: `*_test.go`
- 例: `recipient_repository.go`, `staff_usecase_test.go`

## コメント規約

### 日本語・英語使い分け
- **関数・型のコメント**: 英語（godoc対応）
- **実装詳細のコメント**: 日本語可
- **TODO・FIXME**: 日本語可

```go
// CreateRecipient creates a new recipient with encrypted sensitive data
func (u *RecipientUseCase) CreateRecipient(ctx context.Context, req CreateRecipientRequest) (*CreateRecipientResponse, error) {
    // バリデーション実行
    if err := u.validateCreateRequest(req); err != nil {
        return nil, err
    }
    // ...
}
```

## エラーハンドリング
- エラーメッセージは英語
- ログメッセージは日本語可
- ユーザー向けメッセージは日本語

## セキュリティ規約
- 機微情報は必ず暗号化
- パスワードは平文で保存しない
- 全操作に監査ログを記録
- セッション管理の適切な実装

## テスト規約
- テスト関数名: `Test<FunctionName>_<Scenario>`
- サブテスト使用: `t.Run("scenario", func(t *testing.T) {...})`
- モックの活用: インターフェースベースの設計