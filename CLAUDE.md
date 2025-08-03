# 障害者サービス管理システム - Claude Code実装指示書

## システムプロンプト
```
You are Claude Code acting as a senior Go engineer specializing in desktop applications with strong security practices. You excel at:
- Building secure, privacy-first applications for sensitive data
- Implementing proper cryptography and access controls
- Creating accessible Japanese applications with Fyne
- Following clean architecture principles
- Writing comprehensive tests for business-critical systems

Always prioritize data security, accessibility, and maintainability.
```

## プロジェクト概要

**目的**: 日本の障害者福祉サービス事業所向けの利用者情報管理システム

**重要な特徴**:
- 完全オフライン動作（SQLite）
- 機微情報の暗号化保存
- 日本語対応とアクセシビリティ
- Windows/macOS両対応
- 監査ログ完備

## 初期実装コマンド

### 1. プロジェクト初期化
```bash
# プロジェクト作成
mkdir shien-system && cd shien-system
go mod init shien-system

# 基本ディレクトリ構造作成
mkdir -p {cmd/desktop,internal/{domain,usecase,adapter/{db,crypto,pdf,audit},ui},migrations,testdata}

# 必要な依存関係追加
go get fyne.io/fyne/v2@latest
go get github.com/mattn/go-sqlite3@latest
go get golang.org/x/crypto@latest
go get github.com/google/uuid@latest
```

### 2. データモデル実装
まず`internal/domain/models.go`を実装してください：

```go
package domain

import "time"

type ID = string

type Staff struct {
    ID        ID        `json:"id"`
    Name      string    `json:"name"`
    Role      StaffRole `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type StaffRole string
const (
    RoleAdmin    StaffRole = "admin"
    RoleStaff    StaffRole = "staff"
    RoleReadOnly StaffRole = "readonly"
)

type Recipient struct {
    ID               ID         `json:"id"`
    Name             string     `json:"name"`
    Kana             string     `json:"kana"`
    Sex              Sex        `json:"sex"`
    BirthDate        time.Time  `json:"birth_date"`
    DisabilityName   string     `json:"disability_name"`
    HasDisabilityID  bool       `json:"has_disability_id"`
    Grade            string     `json:"grade"`
    Address          string     `json:"address"`
    Phone            string     `json:"phone"`
    Email            string     `json:"email"`
    PublicAssistance bool       `json:"public_assistance"`
    AdmissionDate    *time.Time `json:"admission_date,omitempty"`
    DischargeDate    *time.Time `json:"discharge_date,omitempty"`
    CreatedAt        time.Time  `json:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at"`
}

type Sex string
const (
    SexFemale Sex = "female"
    SexMale   Sex = "male"
    SexOther  Sex = "other"
    SexNA     Sex = "na"
)

type BenefitCertificate struct {
    ID                     ID        `json:"id"`
    RecipientID            ID        `json:"recipient_id"`
    StartDate              time.Time `json:"start_date"`
    EndDate                time.Time `json:"end_date"`
    Issuer                 string    `json:"issuer"`
    ServiceType            string    `json:"service_type"`
    MaxBenefitDaysPerMonth int       `json:"max_benefit_days_per_month"`
    BenefitDetails         string    `json:"benefit_details"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}

type StaffAssignment struct {
    ID           ID         `json:"id"`
    RecipientID  ID         `json:"recipient_id"`
    StaffID      ID         `json:"staff_id"`
    Role         string     `json:"role"`
    AssignedAt   time.Time  `json:"assigned_at"`
    UnassignedAt *time.Time `json:"unassigned_at,omitempty"`
}

type Consent struct {
    ID          ID         `json:"id"`
    RecipientID ID         `json:"recipient_id"`
    StaffID     ID         `json:"staff_id"`
    ConsentType string     `json:"consent_type"`
    Content     string     `json:"content"`
    Method      string     `json:"method"`
    ObtainedAt  time.Time  `json:"obtained_at"`
    RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

type AuditLog struct {
    ID      ID        `json:"id"`
    ActorID ID        `json:"actor_id"`
    Action  string    `json:"action"`
    Target  string    `json:"target"`
    At      time.Time `json:"at"`
    IP      string    `json:"ip"`
    Details string    `json:"details"`
}
```

### 3. データベース初期化
`migrations/0001_init.sql`を作成：

```sql
-- 職員テーブル
CREATE TABLE staff (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'staff', 'readonly')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 利用者テーブル（暗号化フィールド）
CREATE TABLE recipients (
    id TEXT PRIMARY KEY,
    name_cipher BLOB NOT NULL,
    kana_cipher BLOB,
    sex_cipher BLOB NOT NULL,
    birth_date_cipher BLOB NOT NULL,
    disability_name_cipher BLOB,
    has_disability_id_cipher BLOB NOT NULL,
    grade_cipher BLOB,
    address_cipher BLOB,
    phone_cipher BLOB,
    email_cipher BLOB,
    public_assistance_cipher BLOB NOT NULL,
    admission_date TEXT,
    discharge_date TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 受給者証テーブル
CREATE TABLE benefit_certificates (
    id TEXT PRIMARY KEY,
    recipient_id TEXT NOT NULL REFERENCES recipients(id) ON DELETE CASCADE,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    issuer_cipher BLOB,
    service_type_cipher BLOB,
    max_benefit_days_per_month_cipher BLOB,
    benefit_details_cipher BLOB,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 担当者割り当てテーブル
CREATE TABLE staff_assignments (
    id TEXT PRIMARY KEY,
    recipient_id TEXT NOT NULL REFERENCES recipients(id) ON DELETE CASCADE,
    staff_id TEXT NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    role TEXT,
    assigned_at TEXT NOT NULL,
    unassigned_at TEXT,
    UNIQUE(recipient_id, staff_id, unassigned_at)
);

-- 同意管理テーブル
CREATE TABLE consents (
    id TEXT PRIMARY KEY,
    recipient_id TEXT NOT NULL REFERENCES recipients(id) ON DELETE CASCADE,
    staff_id TEXT NOT NULL REFERENCES staff(id),
    consent_type TEXT NOT NULL,
    content_cipher BLOB NOT NULL,
    method_cipher BLOB NOT NULL,
    obtained_at TEXT NOT NULL,
    revoked_at TEXT
);

-- 監査ログテーブル
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    actor_id TEXT NOT NULL REFERENCES staff(id),
    action TEXT NOT NULL,
    target TEXT NOT NULL,
    at TEXT NOT NULL,
    ip TEXT,
    details TEXT
);

-- インデックス
CREATE INDEX idx_assignments_staff ON staff_assignments(staff_id);
CREATE INDEX idx_assignments_recipient ON staff_assignments(recipient_id);
CREATE INDEX idx_certificates_recipient ON benefit_certificates(recipient_id);
CREATE INDEX idx_consents_recipient ON consents(recipient_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_at ON audit_logs(at);

-- 初期管理者データ
INSERT INTO staff (id, name, role, created_at, updated_at) 
VALUES ('admin-001', '管理者', 'admin', datetime('now'), datetime('now'));
```

### 4. 暗号化レイヤー実装
`internal/adapter/crypto/cipher.go`を実装してください：

```go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
)

type FieldCipher struct {
    gcm cipher.AEAD
}

func NewFieldCipher(key []byte) (*FieldCipher, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("creating cipher: %w", err)
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("creating GCM: %w", err)
    }
    
    return &FieldCipher{gcm: gcm}, nil
}

func (c *FieldCipher) Encrypt(plaintext string) ([]byte, error) {
    if plaintext == "" {
        return nil, nil
    }
    
    nonce := make([]byte, c.gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, fmt.Errorf("generating nonce: %w", err)
    }
    
    ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return ciphertext, nil
}

func (c *FieldCipher) Decrypt(ciphertext []byte) (string, error) {
    if len(ciphertext) == 0 {
        return "", nil
    }
    
    nonceSize := c.gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", fmt.Errorf("decryption failed: %w", err)
    }
    
    return string(plaintext), nil
}
```

### 5. 基本的なGUI構造
`cmd/desktop/main.go`のスケルトンを作成：

```go
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/container"
)

func main() {
    myApp := app.New()
    myApp.Settings().SetTheme(&JapaneseTheme{})
    
    myWindow := myApp.NewWindow("障害者サービス管理システム")
    myWindow.Resize(fyne.NewSize(1200, 800))
    
    // メイン画面の構築
    content := container.NewBorder(
        createHeader(),
        createFooter(),
        createSidebar(),
        nil,
        createMainContent(),
    )
    
    myWindow.SetContent(content)
    myWindow.ShowAndRun()
}

func createHeader() *container.Container {
    return container.NewHBox(
        widget.NewLabel("障害者サービス管理システム"),
        widget.NewButton("ログアウト", func() {
            // ログアウト処理
        }),
    )
}

func createSidebar() *container.Container {
    return container.NewVBox(
        widget.NewButton("利用者一覧", func() {}),
        widget.NewButton("担当者管理", func() {}),
        widget.NewButton("受給者証管理", func() {}),
        widget.NewButton("監査ログ", func() {}),
        widget.NewSeparator(),
        widget.NewButton("設定", func() {}),
    )
}

func createMainContent() *container.Container {
    return container.NewVBox(
        widget.NewLabel("メイン画面"),
    )
}

func createFooter() *container.Container {
    return container.NewHBox(
        widget.NewLabel("Ready"),
    )
}
```

## 実装優先順位

### Phase 1: 基盤機能
1. **データベース接続とマイグレーション**
   - SQLite接続
   - マイグレーション実行機能
   - 基本的なCRUD操作

2. **暗号化システム**
   - フィールド暗号化/復号化
   - 鍵管理（OS Keychain/DPAPI連携）
   - セキュアメモリ管理

3. **認証・認可**
   - ログイン画面
   - ロールベースアクセス制御
   - セッション管理

### Phase 2: コア機能
1. **利用者管理**
   - 利用者の登録・編集・削除
   - 検索・フィルタリング
   - 担当者割り当て

2. **受給者証管理**
   - 受給者証の登録・更新
   - 期限アラート機能
   - 有効性チェック

3. **監査ログ**
   - 全操作の記録
   - ログ閲覧機能
   - エクスポート機能

### Phase 3: 高度な機能
1. **帳票出力（PDF）**
   - 個別支援計画書
   - 利用者一覧
   - 監査報告書

2. **バックアップ・リストア**
   - 暗号化バックアップ
   - データ復元機能
   - スケジュール実行

3. **アクセシビリティ**
   - キーボードナビゲーション
   - フォントサイズ調整
   - コントラスト設定

## 具体的な実装コマンド例

### 利用者一覧画面の実装
```
利用者一覧画面を実装してください：
- Fyneのtable.Tableウィジェットを使用
- 検索ボックスでリアルタイムフィルタ
- ダブルクリックで詳細画面へ遷移
- 新規登録ボタン
- 担当者による絞り込み表示
- 受給者証の有効期限アラート表示
```

### 暗号化データベースアクセス
```
RecipientRepositoryを実装してください：
- internal/adapter/db/recipient.go
- 暗号化フィールドの自動暗号化/復号化
- エラーハンドリング
- トランザクション対応
- 検索用のインデックス（平文ハッシュ）
```

### 監査ログミドルウェア
```
すべてのデータ操作に監査ログを記録するミドルウェアを実装：
- context.Contextからユーザー情報取得
- 操作前後のフック
- PIIを含まない安全なログ形式
- 非同期書き込み
```

## セキュリティ要件

- **暗号化**: AES-256-GCM で機微情報を暗号化
- **鍵管理**: OS固有の安全な鍵保存領域を使用
- **アクセス制御**: ロールベース認可を徹底
- **監査**: 全操作をログに記録（PII除外）
- **メモリ保護**: 機微データの適切なクリア

## 日本語・アクセシビリティ要件

- **フォント**: Noto Sans CJK を埋め込み
- **入力検証**: 日本の住所・電話番号形式対応
- **文字正規化**: Unicode NFKC正規化
- **キーボード操作**: 全機能をキーボードで操作可能
- **色覚配慮**: 情報を色以外でも識別可能

## テスト戦略

```
単体テスト、統合テスト、UIテストの実装：
- testdata/に匿名化サンプルデータ
- 暗号化の往復テスト
- 権限チェックのテスト
- 日本語入力のテスト
- PDFレンダリングのテスト
```

## パフォーマンス目標

- 利用者1000人規模での快適な動作
- 検索結果表示: 500ms以内
- PDF生成: 5秒以内
- 起動時間: 3秒以内

このシステムは障害者福祉の現場で実際に使用されることを想定し、使いやすさとセキュリティを両立することが重要です。