-- ブルートフォース攻撃対策のためのテーブル

-- ログイン試行記録テーブル
CREATE TABLE login_attempts (
    id TEXT PRIMARY KEY,
    ip_address TEXT NOT NULL,
    username TEXT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    attempted_at TEXT NOT NULL,
    user_agent TEXT,
    FOREIGN KEY (username) REFERENCES staff(name) ON DELETE CASCADE
);

-- アカウントロックアウトテーブル
CREATE TABLE account_lockouts (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    lockout_type TEXT NOT NULL CHECK (lockout_type IN ('account', 'ip', 'mixed', 'manual')),
    locked_at TEXT NOT NULL,
    unlocked_at TEXT,
    reason TEXT NOT NULL,
    failure_count INTEGER NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0, -- seconds
    FOREIGN KEY (username) REFERENCES staff(name) ON DELETE CASCADE
);

-- レート制限設定テーブル
CREATE TABLE rate_limit_config (
    id TEXT PRIMARY KEY,
    max_attempts_per_ip INTEGER NOT NULL DEFAULT 5,
    max_attempts_per_user INTEGER NOT NULL DEFAULT 3,
    window_size_minutes INTEGER NOT NULL DEFAULT 15,
    lockout_duration_minutes INTEGER NOT NULL DEFAULT 30,
    backoff_multiplier REAL NOT NULL DEFAULT 2.0,
    max_lockout_hours INTEGER NOT NULL DEFAULT 24,
    whitelist_ips TEXT, -- JSON array of whitelisted IPs
    enable_progressive_lockout BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 攻撃パターン検知テーブル
CREATE TABLE attack_patterns (
    id TEXT PRIMARY KEY,
    pattern_type TEXT NOT NULL, -- 'brute_force', 'dictionary', 'distributed', etc.
    source_ip TEXT NOT NULL,
    target_usernames TEXT NOT NULL, -- JSON array of target usernames
    attempts_count INTEGER NOT NULL DEFAULT 0,
    first_detected_at TEXT NOT NULL,
    last_detected_at TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status TEXT NOT NULL CHECK (status IN ('active', 'blocked', 'resolved')),
    details TEXT
);

-- インデックス作成
CREATE INDEX idx_login_attempts_ip_time ON login_attempts(ip_address, attempted_at);
CREATE INDEX idx_login_attempts_username_time ON login_attempts(username, attempted_at);
CREATE INDEX idx_login_attempts_success ON login_attempts(success, attempted_at);

CREATE INDEX idx_account_lockouts_username ON account_lockouts(username);
CREATE INDEX idx_account_lockouts_ip ON account_lockouts(ip_address);
CREATE INDEX idx_account_lockouts_unlocked ON account_lockouts(unlocked_at);
CREATE INDEX idx_account_lockouts_type_locked ON account_lockouts(lockout_type, locked_at);

CREATE INDEX idx_attack_patterns_ip ON attack_patterns(source_ip);
CREATE INDEX idx_attack_patterns_status ON attack_patterns(status);
CREATE INDEX idx_attack_patterns_detected_at ON attack_patterns(first_detected_at);

-- デフォルトのレート制限設定を挿入
INSERT INTO rate_limit_config (
    id, 
    max_attempts_per_ip, 
    max_attempts_per_user, 
    window_size_minutes, 
    lockout_duration_minutes,
    backoff_multiplier,
    max_lockout_hours,
    whitelist_ips,
    enable_progressive_lockout,
    created_at, 
    updated_at
) VALUES (
    'default-config', 
    5, 
    3, 
    15, 
    30,
    2.0,
    24,
    '[]',
    1,
    datetime('now'), 
    datetime('now')
);