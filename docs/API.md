# API仕様書

## 📖 概要

障害者サービス管理システムの内部APIアーキテクチャについて説明します。このシステムは Clean Architecture パターンを採用し、ドメイン駆動設計（DDD）の原則に従って設計されています。

## 🏗️ アーキテクチャ構成

```
┌─────────────────────────────────────┐
│              UI Layer               │
│         (Fyne Widgets)              │
├─────────────────────────────────────┤
│            Use Case Layer           │
│      (Business Logic)               │
├─────────────────────────────────────┤
│            Domain Layer             │
│        (Entities & Rules)           │
├─────────────────────────────────────┤
│            Adapter Layer            │
│    (DB, Crypto, PDF, Backup)       │
└─────────────────────────────────────┘
```

## 🎯 ドメインモデル

### エンティティ定義

#### 利用者 (Recipient)
```go
type Recipient struct {
    ID               ID         `json:"id"`
    Name             string     `json:"name"`                    // 氏名
    Kana             string     `json:"kana"`                    // フリガナ
    Sex              Sex        `json:"sex"`                     // 性別
    BirthDate        time.Time  `json:"birth_date"`              // 生年月日
    DisabilityName   string     `json:"disability_name"`         // 障害名
    HasDisabilityID  bool       `json:"has_disability_id"`       // 障害者手帳保持
    Grade            string     `json:"grade"`                   // 等級
    Address          string     `json:"address"`                 // 住所
    Phone            string     `json:"phone"`                   // 電話番号
    Email            string     `json:"email"`                   // メールアドレス
    PublicAssistance bool       `json:"public_assistance"`       // 生活保護
    AdmissionDate    *time.Time `json:"admission_date,omitempty"` // 利用開始日
    DischargeDate    *time.Time `json:"discharge_date,omitempty"` // 利用終了日
    CreatedAt        time.Time  `json:"created_at"`              // 作成日時
    UpdatedAt        time.Time  `json:"updated_at"`              // 更新日時
}

type Sex string
const (
    SexFemale Sex = "female"
    SexMale   Sex = "male"
    SexOther  Sex = "other"
    SexNA     Sex = "na"
)
```

#### 職員 (Staff)
```go
type Staff struct {
    ID        ID        `json:"id"`
    Name      string    `json:"name"`        // 職員名
    Role      StaffRole `json:"role"`        // 役割
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type StaffRole string
const (
    RoleAdmin    StaffRole = "admin"     // 管理者
    RoleStaff    StaffRole = "staff"     // 職員
    RoleReadOnly StaffRole = "readonly"  // 閲覧専用
)
```

#### 受給者証 (BenefitCertificate)
```go
type BenefitCertificate struct {
    ID                     ID        `json:"id"`
    RecipientID            ID        `json:"recipient_id"`             // 利用者ID
    StartDate              time.Time `json:"start_date"`               // 開始日
    EndDate                time.Time `json:"end_date"`                 // 終了日
    Issuer                 string    `json:"issuer"`                   // 発行者
    ServiceType            string    `json:"service_type"`             // サービス種別
    MaxBenefitDaysPerMonth int       `json:"max_benefit_days_per_month"` // 月間最大給付日数
    BenefitDetails         string    `json:"benefit_details"`          // 給付詳細
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}
```

#### 監査ログ (AuditLog)
```go
type AuditLog struct {
    ID      ID        `json:"id"`
    ActorID ID        `json:"actor_id"`    // 実行者ID
    Action  string    `json:"action"`      // 操作
    Target  string    `json:"target"`      // 対象
    At      time.Time `json:"at"`          // 実行日時
    IP      string    `json:"ip"`          // IPアドレス
    Details string    `json:"details"`     // 詳細
}
```

## 🔧 Use Case インターフェース

### 利用者管理 (RecipientUseCase)

```go
type RecipientUseCase interface {
    // 利用者作成
    CreateRecipient(ctx context.Context, req CreateRecipientRequest) (*Recipient, error)
    
    // 利用者取得
    GetRecipientByID(ctx context.Context, id ID) (*Recipient, error)
    
    // 利用者更新
    UpdateRecipient(ctx context.Context, req UpdateRecipientRequest) (*Recipient, error)
    
    // 利用者削除
    DeleteRecipient(ctx context.Context, id ID) error
    
    // 利用者一覧取得
    ListRecipients(ctx context.Context, req ListRecipientsRequest) (*ListRecipientsResponse, error)
    
    // 利用者検索
    SearchRecipients(ctx context.Context, req SearchRecipientsRequest) (*SearchRecipientsResponse, error)
}
```

#### リクエスト/レスポンス構造

```go
type CreateRecipientRequest struct {
    Name             string     `json:"name" validate:"required,max=100"`
    Kana             string     `json:"kana" validate:"max=100"`
    Sex              Sex        `json:"sex" validate:"required"`
    BirthDate        time.Time  `json:"birth_date" validate:"required"`
    DisabilityName   string     `json:"disability_name" validate:"max=200"`
    HasDisabilityID  bool       `json:"has_disability_id"`
    Grade            string     `json:"grade" validate:"max=50"`
    Address          string     `json:"address" validate:"max=500"`
    Phone            string     `json:"phone" validate:"max=20"`
    Email            string     `json:"email" validate:"email,max=100"`
    PublicAssistance bool       `json:"public_assistance"`
    AdmissionDate    *time.Time `json:"admission_date,omitempty"`
    ActorID          ID         `json:"actor_id" validate:"required"` // 監査用
}

type UpdateRecipientRequest struct {
    ID               ID         `json:"id" validate:"required"`
    Name             string     `json:"name" validate:"required,max=100"`
    Kana             string     `json:"kana" validate:"max=100"`
    Sex              Sex        `json:"sex" validate:"required"`
    BirthDate        time.Time  `json:"birth_date" validate:"required"`
    DisabilityName   string     `json:"disability_name" validate:"max=200"`
    HasDisabilityID  bool       `json:"has_disability_id"`
    Grade            string     `json:"grade" validate:"max=50"`
    Address          string     `json:"address" validate:"max=500"`
    Phone            string     `json:"phone" validate:"max=20"`
    Email            string     `json:"email" validate:"email,max=100"`
    PublicAssistance bool       `json:"public_assistance"`
    AdmissionDate    *time.Time `json:"admission_date,omitempty"`
    DischargeDate    *time.Time `json:"discharge_date,omitempty"`
    ActorID          ID         `json:"actor_id" validate:"required"`
}

type ListRecipientsRequest struct {
    Limit    int               `json:"limit" validate:"min=1,max=1000"`
    Offset   int               `json:"offset" validate:"min=0"`
    FilterBy FilterRecipients  `json:"filter_by"`
    SortBy   SortRecipients    `json:"sort_by"`
}

type FilterRecipients struct {
    AssignedToStaff *ID       `json:"assigned_to_staff,omitempty"`
    DischargedAfter *time.Time `json:"discharged_after,omitempty"`
    ActiveOnly      bool      `json:"active_only"`
}

type SortRecipients struct {
    Field SortField `json:"field"`
    Order SortOrder `json:"order"`
}

type SortField string
const (
    SortFieldName        SortField = "name"
    SortFieldCreatedAt   SortField = "created_at"
    SortFieldUpdatedAt   SortField = "updated_at"
    SortFieldAdmissionDate SortField = "admission_date"
)

type SortOrder string
const (
    SortOrderAsc  SortOrder = "asc"
    SortOrderDesc SortOrder = "desc"
)

type ListRecipientsResponse struct {
    Recipients []*Recipient `json:"recipients"`
    Total      int          `json:"total"`
    HasMore    bool         `json:"has_more"`
}
```

### 認証 (AuthUseCase)

```go
type AuthUseCase interface {
    // ログイン
    Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
    
    // ログアウト
    Logout(ctx context.Context, sessionID string) error
    
    // セッション検証
    ValidateSession(ctx context.Context, sessionID string) (*Staff, error)
    
    // パスワード変更
    ChangePassword(ctx context.Context, req ChangePasswordRequest) error
}

type LoginRequest struct {
    Username string `json:"username" validate:"required,max=50"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    SessionID string `json:"session_id"`
    User      *Staff `json:"user"`
    CSRFToken string `json:"csrf_token"`
    ExpiresAt time.Time `json:"expires_at"`
}

type ChangePasswordRequest struct {
    UserID      ID     `json:"user_id" validate:"required"`
    OldPassword string `json:"old_password" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=8"`
    ActorID     ID     `json:"actor_id" validate:"required"`
}
```

### 受給者証管理 (CertificateUseCase)

```go
type CertificateUseCase interface {
    // 受給者証作成
    CreateCertificate(ctx context.Context, req CreateCertificateRequest) (*BenefitCertificate, error)
    
    // 受給者証取得
    GetCertificateByID(ctx context.Context, id ID) (*BenefitCertificate, error)
    
    // 利用者の受給者証一覧取得
    GetCertificatesByRecipient(ctx context.Context, recipientID ID) ([]*BenefitCertificate, error)
    
    // 期限切れ間近の受給者証取得
    GetExpiringCertificates(ctx context.Context, withinDays int) ([]*BenefitCertificate, error)
    
    // 受給者証更新
    UpdateCertificate(ctx context.Context, req UpdateCertificateRequest) (*BenefitCertificate, error)
    
    // 受給者証削除
    DeleteCertificate(ctx context.Context, id ID) error
}
```

### 監査ログ (AuditUseCase)

```go
type AuditUseCase interface {
    // ログ記録
    LogAction(ctx context.Context, req LogActionRequest) error
    
    // ログ取得
    GetAuditLogs(ctx context.Context, req GetAuditLogsRequest) (*GetAuditLogsResponse, error)
    
    // 特定対象のログ取得
    GetAuditLogsByTarget(ctx context.Context, target string, limit int) ([]*AuditLog, error)
    
    // 特定ユーザーのログ取得
    GetAuditLogsByActor(ctx context.Context, actorID ID, limit int) ([]*AuditLog, error)
}

type LogActionRequest struct {
    ActorID ID     `json:"actor_id" validate:"required"`
    Action  string `json:"action" validate:"required"`
    Target  string `json:"target" validate:"required"`
    IP      string `json:"ip"`
    Details string `json:"details"`
}

type GetAuditLogsRequest struct {
    Limit     int        `json:"limit" validate:"min=1,max=1000"`
    Offset    int        `json:"offset" validate:"min=0"`
    ActorID   *ID        `json:"actor_id,omitempty"`
    Action    *string    `json:"action,omitempty"`
    Target    *string    `json:"target,omitempty"`
    StartDate *time.Time `json:"start_date,omitempty"`
    EndDate   *time.Time `json:"end_date,omitempty"`
}

type GetAuditLogsResponse struct {
    Logs    []*AuditLog `json:"logs"`
    Total   int         `json:"total"`
    HasMore bool        `json:"has_more"`
}
```

### バックアップ (BackupUseCase)

```go
type BackupUseCase interface {
    // バックアップ作成
    CreateBackup(ctx context.Context, req CreateBackupRequest) (*BackupInfo, error)
    
    // バックアップ一覧取得
    ListBackups(ctx context.Context) ([]*BackupInfo, error)
    
    // バックアップ復元
    RestoreBackup(ctx context.Context, backupID string) error
    
    // バックアップ削除
    DeleteBackup(ctx context.Context, backupID string) error
    
    // 自動バックアップ設定
    ConfigureAutoBackup(ctx context.Context, config AutoBackupConfig) error
}

type CreateBackupRequest struct {
    Description string `json:"description"`
    ActorID     ID     `json:"actor_id" validate:"required"`
}

type BackupInfo struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    Size        int64     `json:"size"`
    Checksum    string    `json:"checksum"`
}

type AutoBackupConfig struct {
    Enabled       bool   `json:"enabled"`
    IntervalHours int    `json:"interval_hours" validate:"min=1,max=168"`
    MaxBackups    int    `json:"max_backups" validate:"min=1,max=100"`
    ActorID       ID     `json:"actor_id" validate:"required"`
}
```

## 🔐 セキュリティコンテキスト

### コンテキスト管理

```go
// Context Keys
type contextKey string

const (
    ContextKeyUserID    contextKey = "user_id"
    ContextKeySessionID contextKey = "session_id"
    ContextKeyCSRFToken contextKey = "csrf_token"
    ContextKeyIP        contextKey = "ip_address"
)

// セキュリティコンテキストの取得
func GetUserIDFromContext(ctx context.Context) (ID, error) {
    userID, ok := ctx.Value(ContextKeyUserID).(ID)
    if !ok {
        return "", fmt.Errorf("user ID not found in context")
    }
    return userID, nil
}

func GetSessionIDFromContext(ctx context.Context) (string, error) {
    sessionID, ok := ctx.Value(ContextKeySessionID).(string)
    if !ok {
        return "", fmt.Errorf("session ID not found in context")
    }
    return sessionID, nil
}
```

### 認可チェック

```go
// 権限チェック関数
func RequireRole(ctx context.Context, requiredRole StaffRole) error {
    userID, err := GetUserIDFromContext(ctx)
    if err != nil {
        return fmt.Errorf("unauthorized: %w", err)
    }
    
    // ユーザーの役割を取得して検証
    // 実装は省略
    return nil
}

// 利用例
func (uc *RecipientUseCaseImpl) DeleteRecipient(ctx context.Context, id ID) error {
    // 管理者権限が必要
    if err := RequireRole(ctx, RoleAdmin); err != nil {
        return err
    }
    
    // 監査ログの記録
    if err := uc.auditUseCase.LogAction(ctx, LogActionRequest{
        ActorID: mustGetUserIDFromContext(ctx),
        Action:  "DELETE_RECIPIENT",
        Target:  fmt.Sprintf("recipient:%s", id),
    }); err != nil {
        return fmt.Errorf("failed to log action: %w", err)
    }
    
    return uc.recipientRepo.Delete(ctx, id)
}
```

## ⚠️ エラーハンドリング

### カスタムエラー定義

```go
// ドメインエラー
var (
    ErrRecipientNotFound      = errors.New("recipient not found")
    ErrCertificateNotFound    = errors.New("certificate not found")
    ErrStaffNotFound          = errors.New("staff not found")
    ErrInvalidCredentials     = errors.New("invalid credentials")
    ErrInsufficientPrivileges = errors.New("insufficient privileges")
    ErrSessionExpired         = errors.New("session expired")
    ErrInvalidCSRFToken       = errors.New("invalid CSRF token")
)

// バリデーションエラー
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
    if len(ve) == 0 {
        return ""
    }
    
    var messages []string
    for _, err := range ve {
        messages = append(messages, err.Error())
    }
    return strings.Join(messages, ", ")
}
```

### エラーレスポンス形式

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

// エラーコード定義
const (
    ErrCodeNotFound           = "NOT_FOUND"
    ErrCodeValidation         = "VALIDATION_ERROR"
    ErrCodeUnauthorized       = "UNAUTHORIZED"
    ErrCodeForbidden          = "FORBIDDEN"
    ErrCodeInternal           = "INTERNAL_ERROR"
    ErrCodeEncryption         = "ENCRYPTION_ERROR"
    ErrCodeDatabase           = "DATABASE_ERROR"
)
```

## 📊 パフォーマンス考慮事項

### データベースアクセス最適化

```go
// バッチ処理での効率的なデータ取得
func (r *RecipientRepository) GetRecipientsByIDs(ctx context.Context, ids []ID) ([]*Recipient, error) {
    // INクエリでバッチ取得
    query := "SELECT id, name_cipher, email_cipher FROM recipients WHERE id IN (" +
        strings.Repeat("?,", len(ids)-1) + "?)"
    
    args := make([]interface{}, len(ids))
    for i, id := range ids {
        args[i] = id
    }
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query recipients: %w", err)
    }
    defer rows.Close()
    
    // 結果処理...
}

// ページネーション対応
func (r *RecipientRepository) ListWithPagination(ctx context.Context, limit, offset int) ([]*Recipient, error) {
    query := `
        SELECT id, name_cipher, email_cipher, created_at 
        FROM recipients 
        ORDER BY created_at DESC 
        LIMIT ? OFFSET ?`
    
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    // 処理続行...
}
```

### キャッシュ戦略

```go
// メモリキャッシュの実装例
type CachedRecipientRepository struct {
    repo  RecipientRepository
    cache map[ID]*Recipient
    mutex sync.RWMutex
    ttl   time.Duration
}

func (r *CachedRecipientRepository) GetByID(ctx context.Context, id ID) (*Recipient, error) {
    // キャッシュから取得試行
    r.mutex.RLock()
    if cached, exists := r.cache[id]; exists {
        r.mutex.RUnlock()
        return cached, nil
    }
    r.mutex.RUnlock()
    
    // データベースから取得
    recipient, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // キャッシュに保存
    r.mutex.Lock()
    r.cache[id] = recipient
    r.mutex.Unlock()
    
    return recipient, nil
}
```

## 📝 使用例

### 利用者作成のフロー

```go
func ExampleCreateRecipient() {
    ctx := context.Background()
    
    // 認証コンテキストの設定
    ctx = context.WithValue(ctx, ContextKeyUserID, "staff-001")
    
    useCase := NewRecipientUseCase(recipientRepo, auditUseCase)
    
    req := CreateRecipientRequest{
        Name:           "田中太郎",
        Kana:           "タナカタロウ",
        Sex:            SexMale,
        BirthDate:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
        Address:        "東京都渋谷区...",
        Phone:          "03-1234-5678",
        Email:          "tanaka@example.com",
        ActorID:        "staff-001",
    }
    
    recipient, err := useCase.CreateRecipient(ctx, req)
    if err != nil {
        log.Fatalf("Failed to create recipient: %v", err)
    }
    
    fmt.Printf("Created recipient: %s (ID: %s)\n", recipient.Name, recipient.ID)
}
```

---

**この API 仕様書は内部アーキテクチャの理解を深め、一貫性のある実装を支援するためのガイドです。**