-- Fix foreign key constraints in rate limiting tables
-- We need to allow rate limiting for non-existent usernames to prevent enumeration attacks

-- First fix login_attempts table
CREATE TABLE login_attempts_new (
    id TEXT PRIMARY KEY,
    ip_address TEXT NOT NULL,
    username TEXT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    attempted_at TEXT NOT NULL,
    user_agent TEXT
    -- No foreign key constraint to allow tracking attempts for non-existent users
);

-- Copy existing data if any
INSERT INTO login_attempts_new SELECT * FROM login_attempts;

-- Drop old table and rename new one
DROP TABLE login_attempts;
ALTER TABLE login_attempts_new RENAME TO login_attempts;

-- Now fix account_lockouts table
CREATE TABLE account_lockouts_new (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    lockout_type TEXT NOT NULL CHECK (lockout_type IN ('account', 'ip', 'mixed', 'manual')),
    locked_at TEXT NOT NULL,
    unlocked_at TEXT,
    reason TEXT NOT NULL,
    failure_count INTEGER NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0 -- seconds, no foreign key constraint
);

-- Copy existing data if any
INSERT INTO account_lockouts_new SELECT * FROM account_lockouts;

-- Drop old table and rename new one
DROP TABLE account_lockouts;
ALTER TABLE account_lockouts_new RENAME TO account_lockouts;

-- Recreate indexes for both tables
CREATE INDEX idx_attempts_ip ON login_attempts(ip_address);
CREATE INDEX idx_attempts_username ON login_attempts(username);
CREATE INDEX idx_attempts_time ON login_attempts(attempted_at);

CREATE INDEX idx_lockouts_username ON account_lockouts(username);
CREATE INDEX idx_lockouts_ip ON account_lockouts(ip_address);
CREATE INDEX idx_lockouts_active ON account_lockouts(username, unlocked_at);