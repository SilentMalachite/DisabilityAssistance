-- セッション管理テーブル追加
-- セキュアなセッション永続化のための暗号化テーブル

-- セッションテーブル（暗号化フィールドあり）
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,                    -- セッションID（暗号学的ランダム）
    user_id TEXT NOT NULL REFERENCES staff(id) ON DELETE CASCADE,  -- ユーザーID
    user_role_cipher BLOB NOT NULL,        -- 暗号化されたユーザーロール
    client_ip_cipher BLOB,                 -- 暗号化されたクライアントIP
    user_agent_cipher BLOB,                -- 暗号化されたUser-Agent
    csrf_token TEXT NOT NULL,              -- CSRF保護トークン
    created_at TEXT NOT NULL,              -- セッション作成時刻
    expires_at TEXT NOT NULL,              -- 有効期限
    last_accessed_at TEXT NOT NULL,        -- 最終アクセス時刻
    is_active INTEGER NOT NULL DEFAULT 1,  -- アクティブフラグ
    invalidation_reason TEXT,              -- 無効化理由
    invalidated_at TEXT                     -- 無効化時刻
);

-- セッション用インデックス
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_active ON sessions(is_active, expires_at);
CREATE INDEX idx_sessions_last_accessed ON sessions(last_accessed_at);

-- セッション設定テーブル
CREATE TABLE session_config (
    id INTEGER PRIMARY KEY CHECK (id = 1), -- 単一行制約
    max_sessions_per_user INTEGER NOT NULL DEFAULT 3,   -- ユーザーあたり最大セッション数
    session_timeout_hours INTEGER NOT NULL DEFAULT 24,  -- セッションタイムアウト（時間）
    cleanup_interval_hours INTEGER NOT NULL DEFAULT 1,  -- クリーンアップ間隔（時間）
    force_single_session INTEGER NOT NULL DEFAULT 0,    -- 単一セッション強制
    require_ip_validation INTEGER NOT NULL DEFAULT 1,   -- IP検証要求
    require_user_agent_validation INTEGER NOT NULL DEFAULT 1,  -- User-Agent検証要求
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- デフォルト設定を挿入
INSERT INTO session_config (
    max_sessions_per_user,
    session_timeout_hours,
    cleanup_interval_hours,
    force_single_session,
    require_ip_validation,
    require_user_agent_validation
) VALUES (3, 24, 1, 0, 1, 1);

-- セッション履歴テーブル（監査目的）
CREATE TABLE session_history (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,              -- 元のセッションID
    user_id TEXT NOT NULL REFERENCES staff(id),
    action TEXT NOT NULL,                  -- CREATE, ACCESS, REFRESH, INVALIDATE, EXPIRE
    client_ip_cipher BLOB,                 -- 暗号化されたクライアントIP
    user_agent_cipher BLOB,                -- 暗号化されたUser-Agent
    details TEXT,                          -- 追加詳細（JSON形式）
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- セッション履歴用インデックス
CREATE INDEX idx_session_history_session_id ON session_history(session_id);
CREATE INDEX idx_session_history_user_id ON session_history(user_id);
CREATE INDEX idx_session_history_action ON session_history(action);
CREATE INDEX idx_session_history_created_at ON session_history(created_at);