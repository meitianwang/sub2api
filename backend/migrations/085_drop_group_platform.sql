-- Drop the platform column and its indexes from the groups table.
-- In passthrough relay mode, groups no longer have a platform concept.
DROP INDEX IF EXISTS idx_groups_platform;
DROP INDEX IF EXISTS group_platform;
ALTER TABLE groups DROP COLUMN IF EXISTS platform;
