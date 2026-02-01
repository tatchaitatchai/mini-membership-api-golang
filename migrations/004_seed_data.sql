-- =========================================================
-- seed_data.sql - Factory initial default data
-- Run after schema migration
-- =========================================================

BEGIN;

-- =========================================================
-- 1) Stores (2 stores)
-- =========================================================
INSERT INTO stores (id, store_name) VALUES
  (1, 'ร้านกาแฟ คาทอม'),
  (2, 'ร้านขนมหวาน สวีทมี');

SELECT setval('stores_id_seq', 2);

-- =========================================================
-- 2) Branches (2 branches per store = 4 total)
-- =========================================================
INSERT INTO branches (id, store_id, branch_name) VALUES
  -- Store 1: ร้านกาแฟ คาทอม
  (1, 1, 'สาขาสยาม'),
  (2, 1, 'สาขาลาดพร้าว'),
  -- Store 2: ร้านขนมหวาน สวีทมี
  (3, 2, 'สาขาเซ็นทรัล'),
  (4, 2, 'สาขาเมกาบางนา');

SELECT setval('branches_id_seq', 4);

-- =========================================================
-- 3) Staff Accounts
--    - Manager: branch_id = NULL, is_store_master = true
--    - Staff: branch_id = assigned branch
--    Password: 'password123' -> bcrypt hash
--    PIN: '1234' -> bcrypt hash
-- =========================================================
INSERT INTO staff_accounts (id, store_id, branch_id, email, password_hash, pin_hash, is_store_master) VALUES
  -- Store 1: Manager (no branch)
  (1, 1, NULL, 'manager@katom.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', true),
  -- Store 1: Staff for branch 1 (สาขาสยาม)
  (2, 1, 1, NULL, 'em1@katom.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', false),
  -- Store 1: Staff for branch 2 (สาขาลาดพร้าว)
  (3, 1, 2, NULL, 'em2@katom.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', false),
  
  -- Store 2: Manager (no branch)
  (4, 2, NULL, 'manager@sweetme.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', true),
  -- Store 2: Staff for branch 3 (สาขาเซ็นทรัล)
  (5, 2, 3, NULL, 'em3@katom.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', false),
  -- Store 2: Staff for branch 4 (สาขาเมกาบางนา)
  (6, 2, 4, NULL, 'em4@katom.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1J1lYS6Xq6Qp5c6Xq6Qp5c6Xq6Qp5c6', false);

SELECT setval('staff_accounts_id_seq', 6);

-- =========================================================
-- 4) Categories (2 per store = 4 total)
-- =========================================================
INSERT INTO categories (id, store_id, category_name) VALUES
  -- Store 1: ร้านกาแฟ
  (1, 1, 'เครื่องดื่มร้อน'),
  (2, 1, 'เครื่องดื่มเย็น'),
  -- Store 2: ร้านขนมหวาน
  (3, 2, 'เค้ก'),
  (4, 2, 'ไอศกรีม');

SELECT setval('categories_id_seq', 4);

-- =========================================================
-- 5) Products (2 per category = 8 total)
-- =========================================================
INSERT INTO products (id, store_id, category_id, product_name, sku, barcode, base_price) VALUES
  -- Store 1, Category 1: เครื่องดื่มร้อน
  (1, 1, 1, 'อเมริกาโน่ร้อน', 'HOT-AMR-001', '8850001000001', 55.00),
  (2, 1, 1, 'ลาเต้ร้อน', 'HOT-LAT-001', '8850001000002', 65.00),
  -- Store 1, Category 2: เครื่องดื่มเย็น
  (3, 1, 2, 'อเมริกาโน่เย็น', 'ICE-AMR-001', '8850001000003', 65.00),
  (4, 1, 2, 'ลาเต้เย็น', 'ICE-LAT-001', '8850001000004', 75.00),
  
  -- Store 2, Category 3: เค้ก
  (5, 2, 3, 'ช็อกโกแลตเค้ก', 'CAKE-CHO-001', '8850002000001', 120.00),
  (6, 2, 3, 'สตรอว์เบอร์รี่ชีสเค้ก', 'CAKE-STR-001', '8850002000002', 135.00),
  -- Store 2, Category 4: ไอศกรีม
  (7, 2, 4, 'ไอศกรีมวานิลลา', 'ICE-VAN-001', '8850002000003', 45.00),
  (8, 2, 4, 'ไอศกรีมช็อกโกแลต', 'ICE-CHO-001', '8850002000004', 45.00);

SELECT setval('products_id_seq', 8);

-- =========================================================
-- 6) Branch Products (link products to all branches of same store)
-- =========================================================
INSERT INTO branch_products (store_id, branch_id, product_id, on_stock, reorder_level) VALUES
  -- Store 1, Branch 1 (สาขาสยาม) - all 4 products
  (1, 1, 1, 100, 10),
  (1, 1, 2, 100, 10),
  (1, 1, 3, 100, 10),
  (1, 1, 4, 100, 10),
  -- Store 1, Branch 2 (สาขาลาดพร้าว) - all 4 products
  (1, 2, 1, 80, 10),
  (1, 2, 2, 80, 10),
  (1, 2, 3, 80, 10),
  (1, 2, 4, 80, 10),
  
  -- Store 2, Branch 3 (สาขาเซ็นทรัล) - all 4 products
  (2, 3, 5, 50, 5),
  (2, 3, 6, 50, 5),
  (2, 3, 7, 200, 20),
  (2, 3, 8, 200, 20),
  -- Store 2, Branch 4 (สาขาเมกาบางนา) - all 4 products
  (2, 4, 5, 40, 5),
  (2, 4, 6, 40, 5),
  (2, 4, 7, 150, 20),
  (2, 4, 8, 150, 20);

COMMIT;

-- =========================================================
-- Summary:
-- =========================================================
-- Stores: 2
--   1. ร้านกาแฟ คาทอม
--   2. ร้านขนมหวาน สวีทมี
--
-- Branches: 4 (2 per store)
--   Store 1: สาขาสยาม, สาขาลาดพร้าว
--   Store 2: สาขาเซ็นทรัล, สาขาเมกาบางนา
--
-- Staff Accounts: 6
--   Store 1: 1 manager + 2 staff (1 per branch)
--   Store 2: 1 manager + 2 staff (1 per branch)
--   Login: manager@katom.com / manager@sweetme.com
--   PIN: 1234 (all staff)
--
-- Categories: 4 (2 per store)
--   Store 1: เครื่องดื่มร้อน, เครื่องดื่มเย็น
--   Store 2: เค้ก, ไอศกรีม
--
-- Products: 8 (2 per category)
--   Store 1: อเมริกาโน่ร้อน, ลาเต้ร้อน, อเมริกาโน่เย็น, ลาเต้เย็น
--   Store 2: ช็อกโกแลตเค้ก, สตรอว์เบอร์รี่ชีสเค้ก, ไอศกรีมวานิลลา, ไอศกรีมช็อกโกแลต
