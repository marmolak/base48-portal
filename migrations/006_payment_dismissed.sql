-- Migration 006: Add dismissed_at column to payments
-- Allows marking payments as "seen/ignored" without assigning them

ALTER TABLE payments ADD COLUMN dismissed_at DATETIME NULL;
ALTER TABLE payments ADD COLUMN dismissed_by INTEGER NULL REFERENCES users(id);
ALTER TABLE payments ADD COLUMN dismissed_reason TEXT NULL;

-- Index for filtering non-dismissed payments efficiently
CREATE INDEX IF NOT EXISTS idx_payments_dismissed_at ON payments(dismissed_at);
