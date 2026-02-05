-- =========================================================
-- Migration 006: Add UNIQUE constraint for PIN per store
-- =========================================================
-- Ensures that each staff member in a store has a unique PIN
-- This prevents PIN collisions when verifying staff identity
-- =========================================================

-- Add unique constraint: PIN must be unique within each store
-- Note: pin_hash can be NULL (some staff may not have PIN set)
-- UNIQUE constraint allows multiple NULLs in PostgreSQL
ALTER TABLE staff_accounts 
ADD CONSTRAINT unique_store_pin UNIQUE (store_id, pin_hash);

-- Add index for faster PIN lookup (used in verify PIN flow)
CREATE INDEX IF NOT EXISTS idx_staff_store_pin 
ON staff_accounts(store_id, pin_hash) 
WHERE pin_hash IS NOT NULL AND is_active = true;
