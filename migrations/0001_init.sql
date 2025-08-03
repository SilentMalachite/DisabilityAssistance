-- 職員テーブル
CREATE TABLE staff (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'staff', 'readonly')),
    password_hash TEXT NOT NULL,
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

-- Performance optimization indexes
CREATE INDEX idx_recipients_name_search ON recipients(name_cipher);
CREATE INDEX idx_recipients_created_at ON recipients(created_at);
CREATE INDEX idx_recipients_discharge_status ON recipients(discharge_date);
CREATE INDEX idx_certificates_date_range ON benefit_certificates(start_date, end_date);
CREATE INDEX idx_certificates_expiry ON benefit_certificates(end_date);
CREATE INDEX idx_assignments_active ON staff_assignments(recipient_id, unassigned_at);
CREATE INDEX idx_consents_type_status ON consents(consent_type, revoked_at);

-- 初期管理者データ
-- デフォルトパスワード: admin123 (bcryptハッシュ)
INSERT INTO staff (id, name, role, password_hash, created_at, updated_at) 
VALUES ('admin-001', '管理者', 'admin', '$2a$10$yPZKJV9e0KO8XfZBix0wn.zYQeVcVBr5vV9oKGhKQ7VhW6Jvr.TPC', datetime('now'), datetime('now'));