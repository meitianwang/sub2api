-- Drop platform column and related indexes from accounts table
DROP INDEX IF EXISTS idx_accounts_platform;
DROP INDEX IF EXISTS account_platform_priority;
ALTER TABLE accounts DROP COLUMN IF EXISTS platform;
