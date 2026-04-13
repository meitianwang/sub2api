-- Drop unused OpenAI Messages dispatch fields from groups table.
-- These fields were never used in backend routing/dispatching logic.
ALTER TABLE groups DROP COLUMN IF EXISTS allow_messages_dispatch;
ALTER TABLE groups DROP COLUMN IF EXISTS default_mapped_model;
