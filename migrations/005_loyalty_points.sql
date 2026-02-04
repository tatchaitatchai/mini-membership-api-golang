-- =========================================================
-- 005_loyalty_points.sql - Loyalty Points System
-- =========================================================

BEGIN;

-- =========================================================
-- 1) Add points_to_redeem field to products table
--    This defines how many points are required to redeem this product as a reward
-- =========================================================

ALTER TABLE products 
ADD COLUMN IF NOT EXISTS points_to_redeem INTEGER DEFAULT NULL;

COMMENT ON COLUMN products.points_to_redeem IS 'Number of points required to redeem this product as a reward. NULL means not redeemable.';

CREATE INDEX IF NOT EXISTS idx_products_redeemable 
ON products(store_id, points_to_redeem) 
WHERE points_to_redeem IS NOT NULL AND points_to_redeem > 0;


-- =========================================================
-- 2) Customer Product Points Table
--    Tracks points earned and redeemed per customer per product
--    e.g., Buy 5 lattes = 5 latte points, use 5 latte points = 1 free latte
-- =========================================================

CREATE TABLE IF NOT EXISTS customer_product_points (
  id            BIGSERIAL PRIMARY KEY,
  store_id      BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  customer_id   BIGINT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
  product_id    BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

  -- current redeemable points for this product (deducted when redeeming)
  points        INTEGER NOT NULL DEFAULT 0,
  -- total points earned for this product (never deducted, for history tracking)
  total_points  INTEGER NOT NULL DEFAULT 0,

  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- one record per customer per product
  UNIQUE (store_id, customer_id, product_id),

  CONSTRAINT chk_customer_product_points_non_negative CHECK (points >= 0),
  CONSTRAINT chk_customer_product_total_points_non_negative CHECK (total_points >= 0)
);

CREATE INDEX IF NOT EXISTS idx_customer_product_points_store ON customer_product_points(store_id);
CREATE INDEX IF NOT EXISTS idx_customer_product_points_customer ON customer_product_points(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_product_points_product ON customer_product_points(product_id);
CREATE INDEX IF NOT EXISTS idx_customer_product_points_lookup ON customer_product_points(store_id, customer_id, product_id);

CREATE TRIGGER trg_customer_product_points_updated_at
BEFORE UPDATE ON customer_product_points
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 3) Point Transactions Table
--    Logs every point earn/redeem event for audit trail
-- =========================================================

CREATE TABLE IF NOT EXISTS point_transactions (
  id              BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id       BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  customer_id     BIGINT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,

  -- EARN = points added from purchase, REDEEM = points deducted for reward
  transaction_type TEXT NOT NULL,

  -- positive for EARN, negative for REDEEM
  points_change   INTEGER NOT NULL,

  -- reference to source (order or redemption)
  reference_table TEXT,
  reference_id    BIGINT,

  -- for EARN: which product was purchased
  -- for REDEEM: which product was redeemed
  product_id      BIGINT REFERENCES products(id) ON DELETE SET NULL,

  note            TEXT,

  -- staff who processed this transaction
  staff_id        BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_point_transaction_type 
    CHECK (transaction_type IN ('EARN', 'REDEEM', 'ADJUST', 'EXPIRE'))
);

CREATE INDEX IF NOT EXISTS idx_point_transactions_store_customer 
ON point_transactions(store_id, customer_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_point_transactions_reference 
ON point_transactions(reference_table, reference_id);

CREATE INDEX IF NOT EXISTS idx_point_transactions_store_time 
ON point_transactions(store_id, created_at DESC);


-- =========================================================
-- 4) Redemptions Table
--    Records each redemption event with details
-- =========================================================

CREATE TABLE IF NOT EXISTS point_redemptions (
  id              BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id       BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  customer_id     BIGINT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,

  product_id      BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
  points_used     INTEGER NOT NULL,
  quantity        INTEGER NOT NULL DEFAULT 1,

  status          TEXT NOT NULL DEFAULT 'COMPLETED',

  staff_id        BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_redemption_points_positive CHECK (points_used > 0),
  CONSTRAINT chk_redemption_quantity_positive CHECK (quantity > 0),
  CONSTRAINT chk_redemption_status CHECK (status IN ('COMPLETED', 'CANCELLED'))
);

CREATE INDEX IF NOT EXISTS idx_point_redemptions_store_customer 
ON point_redemptions(store_id, customer_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_point_redemptions_store_product 
ON point_redemptions(store_id, product_id);

CREATE INDEX IF NOT EXISTS idx_point_redemptions_store_time 
ON point_redemptions(store_id, created_at DESC);

CREATE TRIGGER trg_point_redemptions_updated_at
BEFORE UPDATE ON point_redemptions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
