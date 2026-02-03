-- =========================================================
-- schema.sql (PostgreSQL) - POS ME / Membership (Multi-store)
-- Best practice version (payments = single source of truth)
-- Copy & Run
-- =========================================================

BEGIN;

-- ---------- Extensions ----------
CREATE EXTENSION IF NOT EXISTS citext;

-- ---------- updated_at trigger helper ----------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =========================================================
-- 1) Core: stores / branches / staff
-- =========================================================

CREATE TABLE stores (
  id          BIGSERIAL PRIMARY KEY,
  store_name  TEXT NOT NULL,
  is_active   BOOLEAN NOT NULL DEFAULT true,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stores_is_active ON stores(is_active);

CREATE TRIGGER trg_stores_updated_at
BEFORE UPDATE ON stores
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE branches (
  id               BIGSERIAL PRIMARY KEY,
  store_id          BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,

  branch_name       TEXT NOT NULL,

  is_shift_opened   BOOLEAN NOT NULL DEFAULT false,
  shift_opened_at   TIMESTAMPTZ,
  shift_closed_at   TIMESTAMPTZ,

  is_active         BOOLEAN NOT NULL DEFAULT true,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- ร้านเดียวกัน ห้ามมีชื่อสาขาซ้ำ (store_id ซ้ำได้ตามปกติ)
  UNIQUE (store_id, branch_name)
);

CREATE INDEX idx_branches_store_id ON branches(store_id);
CREATE INDEX idx_branches_store_active ON branches(store_id, is_active);

CREATE TRIGGER trg_branches_updated_at
BEFORE UPDATE ON branches
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE staff_accounts (
  id              BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id       BIGINT REFERENCES branches(id) ON DELETE SET NULL,

  email           CITEXT,
  password_hash   TEXT,      -- email login (store master)
  pin_hash        TEXT,      -- staff pin 4-digit (store hash, not raw)

  is_active       BOOLEAN NOT NULL DEFAULT true,
  is_store_master BOOLEAN NOT NULL DEFAULT false,
  is_working      BOOLEAN NOT NULL DEFAULT false,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (store_id, email)
);

CREATE INDEX idx_staff_store_id ON staff_accounts(store_id);
CREATE INDEX idx_staff_branch_id ON staff_accounts(branch_id);
CREATE INDEX idx_staff_store_active ON staff_accounts(store_id, is_active);

CREATE TRIGGER trg_staff_updated_at
BEFORE UPDATE ON staff_accounts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 2) Customers (members)
-- =========================================================

CREATE TABLE customers (
  id            BIGSERIAL PRIMARY KEY,
  store_id      BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,

  customer_code TEXT,
  full_name     TEXT,
  phone         TEXT,
  phone_last4   CHAR(4),
  email         CITEXT,

  is_active     BOOLEAN NOT NULL DEFAULT true,

  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (store_id, customer_code)
);

CREATE INDEX idx_customers_store_id ON customers(store_id);
CREATE INDEX idx_customers_store_phone_last4 ON customers(store_id, phone_last4);
CREATE INDEX idx_customers_store_phone ON customers(store_id, phone);

CREATE TRIGGER trg_customers_updated_at
BEFORE UPDATE ON customers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 3) Catalog: categories / products / branch_products
-- =========================================================

CREATE TABLE categories (
  id            BIGSERIAL PRIMARY KEY,
  store_id      BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  category_name TEXT NOT NULL,

  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (store_id, category_name)
);

CREATE INDEX idx_categories_store_id ON categories(store_id);

CREATE TRIGGER trg_categories_updated_at
BEFORE UPDATE ON categories
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE products (
  id           BIGSERIAL PRIMARY KEY,
  store_id     BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  category_id  BIGINT REFERENCES categories(id) ON DELETE SET NULL,

  product_name TEXT NOT NULL,
  image_path   TEXT,
  is_active    BOOLEAN NOT NULL DEFAULT true,

  sku          TEXT,
  barcode      TEXT,
  base_price   NUMERIC(12,2) NOT NULL DEFAULT 0,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_store_id ON products(store_id);
CREATE INDEX idx_products_store_active ON products(store_id, is_active);
CREATE INDEX idx_products_store_category ON products(store_id, category_id);
CREATE INDEX idx_products_store_sku ON products(store_id, sku);
CREATE INDEX idx_products_store_barcode ON products(store_id, barcode);

CREATE TRIGGER trg_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE branch_products (
  id            BIGSERIAL PRIMARY KEY,
  store_id       BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id      BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  product_id     BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

  is_active      BOOLEAN NOT NULL DEFAULT true,
  on_stock       INTEGER NOT NULL DEFAULT 0,

  -- low stock warning support
  reorder_level  INTEGER NOT NULL DEFAULT 0,

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (branch_id, product_id)
);

CREATE INDEX idx_branch_products_store_branch ON branch_products(store_id, branch_id);
CREATE INDEX idx_branch_products_branch_product ON branch_products(branch_id, product_id);
CREATE INDEX idx_branch_products_product_id ON branch_products(product_id);
CREATE INDEX idx_branch_products_low_stock ON branch_products(store_id, branch_id, reorder_level, on_stock);

CREATE TRIGGER trg_branch_products_updated_at
BEFORE UPDATE ON branch_products
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 4) Shifts
-- =========================================================

CREATE TABLE shifts (
  id                 BIGSERIAL PRIMARY KEY,
  store_id            BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id           BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,

  start_money_inbox   NUMERIC(12,2) NOT NULL DEFAULT 0,
  end_money_inbox     NUMERIC(12,2),

  started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at            TIMESTAMPTZ,

  is_active_shift     BOOLEAN NOT NULL DEFAULT true,

  opened_by           BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,
  closed_by           BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shifts_store_branch_active ON shifts(store_id, branch_id, is_active_shift);
CREATE INDEX idx_shifts_store_branch_started ON shifts(store_id, branch_id, started_at DESC);

CREATE TRIGGER trg_shifts_updated_at
BEFORE UPDATE ON shifts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 5) Promotions
-- =========================================================

CREATE TABLE promotion_types (
  id         BIGSERIAL PRIMARY KEY,
  store_id   BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,

  name       TEXT NOT NULL,
  detail     TEXT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (store_id, name)
);

CREATE INDEX idx_promotion_types_store ON promotion_types(store_id);

CREATE TRIGGER trg_promotion_types_updated_at
BEFORE UPDATE ON promotion_types
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE promotion_type_branches (
  id                BIGSERIAL PRIMARY KEY,
  store_id           BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id          BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  promotion_type_id  BIGINT NOT NULL REFERENCES promotion_types(id) ON DELETE RESTRICT,

  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (branch_id, promotion_type_id)
);

CREATE INDEX idx_promotion_type_branches_store_branch ON promotion_type_branches(store_id, branch_id);
CREATE INDEX idx_promotion_type_branches_type ON promotion_type_branches(promotion_type_id);

CREATE TRIGGER trg_promotion_type_branches_updated_at
BEFORE UPDATE ON promotion_type_branches
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE promotions (
  id                 BIGSERIAL PRIMARY KEY,
  store_id            BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  promotion_type_id   BIGINT NOT NULL REFERENCES promotion_types(id) ON DELETE RESTRICT,

  promotion_name      TEXT NOT NULL,
  is_active           BOOLEAN NOT NULL DEFAULT true,
  starts_at           TIMESTAMPTZ,
  ends_at             TIMESTAMPTZ,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_promotions_store_active ON promotions(store_id, is_active);
CREATE INDEX idx_promotions_store_time ON promotions(store_id, starts_at, ends_at);
CREATE INDEX idx_promotions_store_type ON promotions(store_id, promotion_type_id);

CREATE TRIGGER trg_promotions_updated_at
BEFORE UPDATE ON promotions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE promotion_configs (
  id                       BIGSERIAL PRIMARY KEY,
  promotion_id              BIGINT NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,

  percent_discount          NUMERIC(8,4),
  baht_discount             NUMERIC(12,2),
  total_price_set_discount  NUMERIC(12,2),
  old_price_set             NUMERIC(12,2),

  count_condition_product   INTEGER,
  product_id                BIGINT REFERENCES products(id) ON DELETE SET NULL,

  created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_promotion_configs_promotion_id ON promotion_configs(promotion_id);
CREATE INDEX idx_promotion_configs_product_id ON promotion_configs(product_id);

CREATE TRIGGER trg_promotion_configs_updated_at
BEFORE UPDATE ON promotion_configs
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE promotion_products (
  id                BIGSERIAL PRIMARY KEY,
  promotion_id       BIGINT NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
  product_id         BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (promotion_id, product_id)
);

CREATE INDEX idx_promotion_products_promotion ON promotion_products(promotion_id);
CREATE INDEX idx_promotion_products_product ON promotion_products(product_id);

CREATE TRIGGER trg_promotion_products_updated_at
BEFORE UPDATE ON promotion_products
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 6) Orders / Order items (best practice)
--    payments is the ONLY source of truth for paid amounts
-- =========================================================

CREATE TABLE orders (
  id              BIGSERIAL PRIMARY KEY,
  store_id         BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id        BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  shift_id         BIGINT REFERENCES shifts(id) ON DELETE SET NULL,

  customer_id      BIGINT REFERENCES customers(id) ON DELETE SET NULL,
  staff_id         BIGINT NOT NULL REFERENCES staff_accounts(id) ON DELETE RESTRICT,

  -- totals
  subtotal         NUMERIC(12,2) NOT NULL DEFAULT 0,
  discount_total   NUMERIC(12,2) NOT NULL DEFAULT 0,
  total_price      NUMERIC(12,2) NOT NULL DEFAULT 0,  -- grand total after discount

  -- cash change (usually from cash part only)
  change_amount    NUMERIC(12,2) NOT NULL DEFAULT 0,

  status           TEXT NOT NULL DEFAULT 'PAID',
  cancelled_by     BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,
  cancel_reason    TEXT,
  cancelled_at     TIMESTAMPTZ,

  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_orders_status
    CHECK (status IN ('OPEN','PAID','CANCELLED','VOID'))
);

CREATE INDEX idx_orders_store_branch_created ON orders(store_id, branch_id, created_at DESC);
CREATE INDEX idx_orders_shift_id ON orders(shift_id);
CREATE INDEX idx_orders_store_status ON orders(store_id, status);
CREATE INDEX idx_orders_customer ON orders(store_id, customer_id);

CREATE TRIGGER trg_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE order_items (
  id               BIGSERIAL PRIMARY KEY,
  order_id          BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id        BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

  quantity          INTEGER NOT NULL,
  price             NUMERIC(12,2) NOT NULL DEFAULT 0,

  from_stock_count  INTEGER,
  to_stock_count    INTEGER,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

CREATE TRIGGER trg_order_items_updated_at
BEFORE UPDATE ON order_items
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- allow multiple promotions per order
CREATE TABLE order_promotions (
  id               BIGSERIAL PRIMARY KEY,
  order_id          BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  promotion_id      BIGINT NOT NULL REFERENCES promotions(id) ON DELETE RESTRICT,

  discount_amount   NUMERIC(12,2) NOT NULL DEFAULT 0,
  metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_promotions_order ON order_promotions(order_id);
CREATE INDEX idx_order_promotions_promotion ON order_promotions(promotion_id);


-- =========================================================
-- 7) Payments + slip attachments (split payment supported)
-- =========================================================

CREATE TABLE payments (
  id          BIGSERIAL PRIMARY KEY,
  order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,

  method      TEXT NOT NULL, -- CASH, TRANSFER, QR, etc.
  amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
  paid_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_payments_method
    CHECK (method IN ('CASH','TRANSFER','QR','CARD','OTHER')),

  CONSTRAINT chk_payments_amount_non_negative
    CHECK (amount >= 0)
);

CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_payments_method ON payments(method);
CREATE INDEX idx_payments_paid_at ON payments(paid_at DESC);


CREATE TABLE payment_attachments (
  id          BIGSERIAL PRIMARY KEY,
  payment_id   BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
  file_path   TEXT NOT NULL,
  file_type   TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_attachments_payment ON payment_attachments(payment_id);


-- =========================================================
-- 8) Inventory ledger + stock transfers
-- =========================================================

CREATE TABLE inventory_movements (
  id               BIGSERIAL PRIMARY KEY,
  store_id          BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id         BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  product_id        BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

  movement_type     TEXT NOT NULL,
  quantity_change   INTEGER NOT NULL,

  from_stock_count  INTEGER,
  to_stock_count    INTEGER,

  reason            TEXT,
  note              TEXT,

  changed_by        BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,

  reference_table   TEXT,
  reference_id      BIGINT,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_inventory_movement_type
    CHECK (movement_type IN (
      'SALE','CANCEL_SALE','RECEIVE','ISSUE','ADJUST','TRANSFER_IN','TRANSFER_OUT','DAMAGE'
    ))
);

CREATE INDEX idx_inventory_movements_store_branch_time ON inventory_movements(store_id, branch_id, created_at DESC);
CREATE INDEX idx_inventory_movements_product_time ON inventory_movements(product_id, created_at DESC);
CREATE INDEX idx_inventory_movements_reference ON inventory_movements(reference_table, reference_id);


CREATE TABLE stock_transfers (
  id                BIGSERIAL PRIMARY KEY,
  store_id           BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,

  from_branch_id     BIGINT REFERENCES branches(id) ON DELETE SET NULL,
  to_branch_id       BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,

  status             TEXT NOT NULL DEFAULT 'CREATED',

  sent_by            BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,
  received_by        BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,

  sent_at            TIMESTAMPTZ,
  received_at        TIMESTAMPTZ,

  note               TEXT,

  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_stock_transfers_status
    CHECK (status IN ('CREATED','SENT','RECEIVED','CANCELLED'))
);

CREATE INDEX idx_stock_transfers_store ON stock_transfers(store_id);
CREATE INDEX idx_stock_transfers_store_to_branch ON stock_transfers(store_id, to_branch_id, created_at DESC);
CREATE INDEX idx_stock_transfers_store_from_branch ON stock_transfers(store_id, from_branch_id, created_at DESC);
CREATE INDEX idx_stock_transfers_status ON stock_transfers(store_id, status);

CREATE TRIGGER trg_stock_transfers_updated_at
BEFORE UPDATE ON stock_transfers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE stock_transfer_items (
  id                 BIGSERIAL PRIMARY KEY,
  stock_transfer_id   BIGINT NOT NULL REFERENCES stock_transfers(id) ON DELETE CASCADE,
  product_id          BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

  send_count          INTEGER NOT NULL DEFAULT 0,
  receive_count       INTEGER NOT NULL DEFAULT 0,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT chk_stock_transfer_items_counts
    CHECK (send_count >= 0 AND receive_count >= 0)
);

CREATE INDEX idx_stock_transfer_items_transfer ON stock_transfer_items(stock_transfer_id);
CREATE INDEX idx_stock_transfer_items_product ON stock_transfer_items(product_id);

CREATE TRIGGER trg_stock_transfer_items_updated_at
BEFORE UPDATE ON stock_transfer_items
FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- =========================================================
-- 9) App sessions (PIN re-auth support)
-- =========================================================

CREATE TABLE app_sessions (
  id            BIGSERIAL PRIMARY KEY,
  store_id       BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id      BIGINT REFERENCES branches(id) ON DELETE SET NULL,
  staff_id       BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,

  session_token  TEXT NOT NULL UNIQUE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at     TIMESTAMPTZ
);

CREATE INDEX idx_app_sessions_store_staff ON app_sessions(store_id, staff_id, last_seen_at DESC);
CREATE INDEX idx_app_sessions_store_branch ON app_sessions(store_id, branch_id, last_seen_at DESC);

-- 1) Enum types (ถ้ายังไม่มี)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cash_direction') THEN
    CREATE TYPE cash_direction AS ENUM ('IN', 'OUT');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'shift_cash_movement_type') THEN
    CREATE TYPE shift_cash_movement_type AS ENUM (
      'ADD_FLOAT',     -- เติมเงินทอน
      'PAID_OUT',      -- จ่ายเงินสดออก (ค่าใช้จ่าย)
      'CASH_DROP',     -- ถอนเงินออกไปเก็บ/เข้าตู้เซฟ
      'ADJUST',        -- ปรับยอด (กรณีพิเศษ)
      'REFUND_CASH'    -- คืนเงินสด (ถ้ามี)
    );
  END IF;
END $$;

-- 2) Table: shift_cash_movements
CREATE TABLE IF NOT EXISTS shift_cash_movements (
  id BIGSERIAL PRIMARY KEY,

  store_id  BIGINT NOT NULL,
  branch_id BIGINT NOT NULL,
  shift_id  BIGINT NOT NULL,

  movement_type shift_cash_movement_type NOT NULL,
  direction     cash_direction NOT NULL,

  amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),

  note TEXT NULL,

  created_by_staff_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE shift_cash_movements
  ADD CONSTRAINT fk_shift_cash_movements_store
  FOREIGN KEY (store_id) REFERENCES stores(id)
  ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE shift_cash_movements
  ADD CONSTRAINT fk_shift_cash_movements_branch
  FOREIGN KEY (branch_id) REFERENCES branches(id)
  ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE shift_cash_movements
  ADD CONSTRAINT fk_shift_cash_movements_shift
  FOREIGN KEY (shift_id) REFERENCES shifts(id)
  ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE shift_cash_movements
  ADD CONSTRAINT fk_shift_cash_movements_staff
  FOREIGN KEY (created_by_staff_id) REFERENCES staff_accounts(id)
  ON UPDATE CASCADE ON DELETE RESTRICT;

BEGIN;

-- เก็บเหตุการณ์ "นับสต็อก" ตอนปิดกะ (1 shift อาจนับ 1 ครั้ง)
CREATE TABLE IF NOT EXISTS shift_stock_counts (
  id               BIGSERIAL PRIMARY KEY,

  store_id          BIGINT NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
  branch_id         BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
  shift_id          BIGINT NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,

  counted_by        BIGINT REFERENCES staff_accounts(id) ON DELETE SET NULL,
  counted_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

  note              TEXT,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- ปกติ 1 กะ นับได้ 1 ครั้ง (ถ้าคุณอยากให้นับได้หลายรอบ ให้เอา UNIQUE นี้ออก)
  UNIQUE (shift_id)
);

CREATE INDEX IF NOT EXISTS idx_shift_stock_counts_shift_time
  ON shift_stock_counts (shift_id, counted_at DESC);

CREATE INDEX IF NOT EXISTS idx_shift_stock_counts_store_branch_time
  ON shift_stock_counts (store_id, branch_id, counted_at DESC);

-- ใช้ trigger updated_at ตัวเดียวกับที่คุณมีอยู่แล้ว
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'set_updated_at') THEN
    DROP TRIGGER IF EXISTS trg_shift_stock_counts_updated_at ON shift_stock_counts;
    CREATE TRIGGER trg_shift_stock_counts_updated_at
    BEFORE UPDATE ON shift_stock_counts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END $$;

-- เก็บรายละเอียดสินค้าแต่ละตัวที่นับ
CREATE TABLE IF NOT EXISTS shift_stock_count_items (
  id                    BIGSERIAL PRIMARY KEY,
  shift_stock_count_id  BIGINT NOT NULL REFERENCES shift_stock_counts(id) ON DELETE CASCADE,
  product_id            BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
  expected_stock        INT NOT NULL DEFAULT 0,
  actual_stock          INT NOT NULL DEFAULT 0,
  difference            INT NOT NULL DEFAULT 0,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_shift_stock_count_items_count_id
  ON shift_stock_count_items (shift_stock_count_id);

CREATE INDEX IF NOT EXISTS idx_shift_stock_count_items_product
  ON shift_stock_count_items (product_id);

COMMIT;