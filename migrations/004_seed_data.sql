-- =========================================================
-- seed_data.sql - Factory initial default data
-- Run after schema migration
-- =========================================================

BEGIN;

-- =========================================================
-- 1) Stores (2 stores)
-- =========================================================
INSERT INTO stores (id, store_name) VALUES
  (1, 'ร้านกาแฟ เดอะคอฟฟี่'),
  (2, 'ร้านขนมหวาน สวีทมี');

SELECT setval('stores_id_seq', 2);

-- =========================================================
-- 2) Branches (2 branches per store = 4 total)
-- =========================================================
INSERT INTO branches (id, store_id, branch_name) VALUES
  -- Store 1: ร้านกาแฟ เดอะคอฟฟี่
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
--    Password: '123456' -> bcrypt hash
--    PIN: unique per staff (see below)
-- =========================================================
INSERT INTO staff_accounts (id, store_id, branch_id, email, password_hash, pin_hash, is_store_master) VALUES
  -- Store 1: Manager (no branch) - PIN: 1234
  (1, 1, NULL, 'manager@thecoffee.com', '$2y$10$flyI5EBS3p4Szcs.KT7JP.Or7w1HXt07TuDTKUHbL7KmLp0aB29K2', '03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4', true),
  -- Store 1: Staff for branch 1 (สาขาสยาม) - PIN: 2345
  (2, 1, 1, 'staff1@thecoffee.com', '$2y$10$flyI5EBS3p4Szcs.KT7JP.Or7w1HXt07TuDTKUHbL7KmLp0aB29K2', '38083c7ee9121e17401883566a148aa5c2e2d55dc53bc4a94a026517dbff3c6b', false),
  -- Store 1: Staff for branch 2 (สาขาลาดพร้าว) - PIN: 3456
  (3, 1, 2, 'staff2@thecoffee.com', '$2y$10$flyI5EBS3p4Szcs.KT7JP.Or7w1HXt07TuDTKUHbL7KmLp0aB29K2', 'ceaa28bba4caba687dc31b1bbe79eca3c70c33f871f1ce8f528cf9ab5cfd76dd', false),
  
  -- Store 2: Manager (no branch) - PIN: 4567
  (4, 2, NULL, 'manager@sweetme.com', '$2y$10$flyI5EBS3p4Szcs.KT7JP.Or7w1HXt07TuDTKUHbL7KmLp0aB29K2', 'db2e7f1bd5ab9968ae76199b7cc74795ca7404d5a08d78567715ce532f9d2669', true),
  -- Store 2: Staff for branch 3 (สาขาเซ็นทรัล) - PIN: 5678
  (5, 2, 3, 'staff1@sweetme.com', '$2y$10$flyI5EBS3p4Szcs.KT7JP.Or7w1HXt07TuDTKUHbL7KmLp0aB29K2', 'f8638b979b2f4f793ddb6dbd197e0ee25a7a6ea32b0ae22f5e3c5d119d839e75', false),
  -- Store 2: Staff for branch 4 (สาขาเมกาบางนา) - PIN: 6789
  (6, 2, 4, 'staff2@sweetme.com', '$2y$10$flyI5EBS3p4Szcs.KT7JP.Or7w1HXt07TuDTKUHbL7KmLp0aB29K2', '499bc7df9d8873c1c38e6898177c343b2a34d2eb43178a9eb4efcb993366c8cd', false);

SELECT setval('staff_accounts_id_seq', 6);

-- =========================================================
-- 4) Categories (2 per store = 4 total)
-- =========================================================
INSERT INTO categories (id, store_id, category_name) VALUES
  -- Store 1: ร้านกาแฟ เดอะคอฟฟี่
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
  -- Store 1, Branch 2 (สาขาลาดพร้าว) - all 4 products (some low stock for testing)
  (1, 2, 1, 5, 10),   -- LOW: อเมริกาโน่ร้อน สต็อก 5 < เกณฑ์ 10
  (1, 2, 2, 0, 10),   -- CRITICAL: ลาเต้ร้อน หมด!
  (1, 2, 3, 80, 10),
  (1, 2, 4, 8, 10),   -- LOW: ลาเต้เย็น สต็อก 8 < เกณฑ์ 10
  
  -- Store 2, Branch 3 (สาขาเซ็นทรัล) - all 4 products (some low stock for testing)
  (2, 3, 5, 3, 5),    -- LOW: ช็อกโกแลตเค้ก สต็อก 3 < เกณฑ์ 5
  (2, 3, 6, 0, 5),    -- CRITICAL: สตรอว์เบอร์รี่ชีสเค้ก หมด!
  (2, 3, 7, 200, 20),
  (2, 3, 8, 15, 20),  -- LOW: ไอศกรีมช็อกโกแลต สต็อก 15 < เกณฑ์ 20
  -- Store 2, Branch 4 (สาขาเมกาบางนา) - all 4 products
  (2, 4, 5, 40, 5),
  (2, 4, 6, 40, 5),
  (2, 4, 7, 150, 20),
  (2, 4, 8, 150, 20);

-- =========================================================
-- 7) Customers (2 per store = 4 total)
--    PIN: '1234' -> SHA256: 03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4
--    PIN: '1231' -> SHA256: 52a6eb687cd22e80d3342eac6fcc7f2e19209e8f83eb9b82e81c6f3e6f30743b
-- =========================================================
INSERT INTO customers (id, store_id, customer_code, full_name, phone, phone_last4) VALUES
  -- Store 1: ร้านกาแฟ เดอะคอฟฟี่
  (1, 1, 'CUST-001', 'สมชาย ใจดี', '0812345678', '5678'),
  (2, 1, 'CUST-002', 'สมหญิง รักสวย', '0898765432', '5432'),
  -- Store 2: ร้านขนมหวาน สวีทมี
  (3, 2, 'CUST-003', 'วิชัย มั่งมี', '0856781234', '1234'),
  (4, 2, 'CUST-004', 'วิภา สุขใจ', '0843219876', '9876');

SELECT setval('customers_id_seq', 4);

-- =========================================================
-- 8) Promotion Types (5 types per store = 10 total)
--    1. PERCENT_DISCOUNT - ลดเป็นเปอร์เซ็นต์ตรงๆ
--    2. BAHT_DISCOUNT - ลดเป็นจำนวนบาทตรงๆ
--    3. SET_DISCOUNT - ซื้อเป็น set ลดราคา (auto-detect)
--    4. BUY_N_PERCENT_OFF - ซื้อ N ชิ้น ลด N%
--    5. BUY_N_BAHT_OFF - ซื้อ N ชิ้น ลด N บาท
-- =========================================================
INSERT INTO promotion_types (id, store_id, name, detail) VALUES
  -- Store 1: ร้านกาแฟ เดอะคอฟฟี่
  (1, 1, 'ลดเปอร์เซ็นต์', 'ลดราคาเป็นเปอร์เซ็นต์ตรงๆ เช่น ลด 10%'),
  (2, 1, 'ลดบาท', 'ลดราคาเป็นจำนวนบาทตรงๆ เช่น ลด 20 บาท'),
  (3, 1, 'ซื้อเป็นเซ็ต', 'ซื้อสินค้าครบเซ็ตได้ราคาพิเศษ (ระบบ auto-detect)'),
  (4, 1, 'ซื้อครบลดเปอร์เซ็นต์', 'ซื้อครบ N ชิ้น ลด N% เช่น ซื้อ 3 ชิ้น ลด 15%'),
  (5, 1, 'ซื้อครบลดบาท', 'ซื้อครบ N ชิ้น ลด N บาท เช่น ซื้อ 2 ชิ้น ลด 30 บาท'),
  
  -- Store 2: ร้านขนมหวาน สวีทมี
  (6, 2, 'ลดเปอร์เซ็นต์', 'ลดราคาเป็นเปอร์เซ็นต์ตรงๆ เช่น ลด 10%'),
  (7, 2, 'ลดบาท', 'ลดราคาเป็นจำนวนบาทตรงๆ เช่น ลด 20 บาท'),
  (8, 2, 'ซื้อเป็นเซ็ต', 'ซื้อสินค้าครบเซ็ตได้ราคาพิเศษ (ระบบ auto-detect)'),
  (9, 2, 'ซื้อครบลดเปอร์เซ็นต์', 'ซื้อครบ N ชิ้น ลด N% เช่น ซื้อ 3 ชิ้น ลด 15%'),
  (10, 2, 'ซื้อครบลดบาท', 'ซื้อครบ N ชิ้น ลด N บาท เช่น ซื้อ 2 ชิ้น ลด 30 บาท');

SELECT setval('promotion_types_id_seq', 10);

-- =========================================================
-- 9) Promotion Type Branches (link types to all branches)
-- =========================================================
INSERT INTO promotion_type_branches (store_id, branch_id, promotion_type_id) VALUES
  -- Store 1, Branch 1 (สาขาสยาม) - all 5 types
  (1, 1, 1), (1, 1, 2), (1, 1, 3), (1, 1, 4), (1, 1, 5),
  -- Store 1, Branch 2 (สาขาลาดพร้าว) - all 5 types
  (1, 2, 1), (1, 2, 2), (1, 2, 3), (1, 2, 4), (1, 2, 5),
  -- Store 2, Branch 3 (สาขาเซ็นทรัล) - all 5 types
  (2, 3, 6), (2, 3, 7), (2, 3, 8), (2, 3, 9), (2, 3, 10),
  -- Store 2, Branch 4 (สาขาเมกาบางนา) - all 5 types
  (2, 4, 6), (2, 4, 7), (2, 4, 8), (2, 4, 9), (2, 4, 10);

-- =========================================================
-- 10) Promotions (4 promotions per store = 8 total)
-- =========================================================
INSERT INTO promotions (id, store_id, promotion_type_id, promotion_name, is_active, starts_at, ends_at) VALUES
  -- Store 1: ร้านกาแฟ เดอะคอฟฟี่
  -- Type 1: ลดเปอร์เซ็นต์ - ลาเต้ลด 10%
  (1, 1, 1, 'ลาเต้ลด 10%', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- Type 2: ลดบาท - อเมริกาโน่ลด 10 บาท
  (2, 1, 2, 'อเมริกาโน่ลด 10 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- Type 3: ซื้อเป็นเซ็ต - อเมริกาโน่ร้อน + ลาเต้เย็น = 120 บาท (ปกติ 55+75=130)
  (3, 1, 3, 'คู่หูกาแฟ 120 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- Type 4: ซื้อครบลดเปอร์เซ็นต์ - ซื้อ 3 แก้ว ลด 15%
  (4, 1, 4, 'ซื้อ 3 แก้ว ลด 15%', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- BILL-LEVEL: ลดท้ายบิล 5% (ไม่ผูกสินค้า)
  (9, 1, 1, 'ลดท้ายบิล 5%', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- BILL-LEVEL: ลดท้ายบิล 20 บาท (ไม่ผูกสินค้า)
  (10, 1, 2, 'ลดท้ายบิล 20 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  
  -- Store 2: ร้านขนมหวาน สวีทมี
  -- Type 6: ลดเปอร์เซ็นต์ - เค้กลด 15%
  (5, 2, 6, 'เค้กลด 15%', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- Type 7: ลดบาท - ไอศกรีมลด 5 บาท
  (6, 2, 7, 'ไอศกรีมลด 5 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- Type 8: ซื้อเป็นเซ็ต - เค้ก + ไอศกรีม = 150 บาท (ปกติ 120+45=165)
  (7, 2, 8, 'เค้ก+ไอศกรีม 150 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- Type 10: ซื้อครบลดบาท - ซื้อไอศกรีม 2 ถ้วย ลด 20 บาท
  (8, 2, 10, 'ไอศกรีม 2 ถ้วย ลด 20 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- BILL-LEVEL: ลดท้ายบิล 10% (ไม่ผูกสินค้า)
  (11, 2, 6, 'ลดท้ายบิล 10%', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07'),
  -- BILL-LEVEL: ลดท้ายบิล 15 บาท (ไม่ผูกสินค้า)
  (12, 2, 7, 'ลดท้ายบิล 15 บาท', true, '2025-01-01 00:00:00+07', '2026-12-31 23:59:59+07');

SELECT setval('promotions_id_seq', 12);

-- =========================================================
-- 11) Promotion Configs (settings for each promotion)
-- =========================================================
INSERT INTO promotion_configs (id, promotion_id, percent_discount, baht_discount, total_price_set_discount, old_price_set, count_condition_product, product_id) VALUES
  -- Promotion 1: ลาเต้ลด 10% (type: ลดเปอร์เซ็นต์)
  (1, 1, 10.0000, NULL, NULL, NULL, NULL, NULL),
  
  -- Promotion 2: อเมริกาโน่ลด 10 บาท (type: ลดบาท)
  (2, 2, NULL, 10.00, NULL, NULL, NULL, NULL),
  
  -- Promotion 3: คู่หูกาแฟ 120 บาท (type: ซื้อเป็นเซ็ต)
  -- total_price_set_discount = 120, old_price_set = 130 (55+75)
  (3, 3, NULL, NULL, 120.00, 130.00, NULL, NULL),
  
  -- Promotion 4: ซื้อ 3 แก้ว ลด 15% (type: ซื้อครบลดเปอร์เซ็นต์)
  -- count_condition_product = 3, percent_discount = 15%
  (4, 4, 15.0000, NULL, NULL, NULL, 3, NULL),
  
  -- Promotion 5: เค้กลด 15% (type: ลดเปอร์เซ็นต์)
  (5, 5, 15.0000, NULL, NULL, NULL, NULL, NULL),
  
  -- Promotion 6: ไอศกรีมลด 5 บาท (type: ลดบาท)
  (6, 6, NULL, 5.00, NULL, NULL, NULL, NULL),
  
  -- Promotion 7: เค้ก+ไอศกรีม 150 บาท (type: ซื้อเป็นเซ็ต)
  -- total_price_set_discount = 150, old_price_set = 165 (120+45)
  (7, 7, NULL, NULL, 150.00, 165.00, NULL, NULL),
  
  -- Promotion 8: ไอศกรีม 2 ถ้วย ลด 20 บาท (type: ซื้อครบลดบาท)
  -- count_condition_product = 2, baht_discount = 20
  (8, 8, NULL, 20.00, NULL, NULL, 2, NULL),
  
  -- BILL-LEVEL Promotions (ไม่ผูกสินค้า - ลดท้ายบิล)
  -- Promotion 9: ลดท้ายบิล 5% (Store 1)
  (9, 9, 5.0000, NULL, NULL, NULL, NULL, NULL),
  -- Promotion 10: ลดท้ายบิล 20 บาท (Store 1)
  (10, 10, NULL, 20.00, NULL, NULL, NULL, NULL),
  -- Promotion 11: ลดท้ายบิล 10% (Store 2)
  (11, 11, 10.0000, NULL, NULL, NULL, NULL, NULL),
  -- Promotion 12: ลดท้ายบิล 15 บาท (Store 2)
  (12, 12, NULL, 15.00, NULL, NULL, NULL, NULL);

SELECT setval('promotion_configs_id_seq', 12);

-- =========================================================
-- 12) Promotion Products (link products to promotions)
-- =========================================================
INSERT INTO promotion_products (promotion_id, product_id) VALUES
  -- Promotion 1: ลาเต้ลด 10% -> ลาเต้ร้อน (2), ลาเต้เย็น (4)
  (1, 2), (1, 4),
  
  -- Promotion 2: อเมริกาโน่ลด 10 บาท -> อเมริกาโน่ร้อน (1), อเมริกาโน่เย็น (3)
  (2, 1), (2, 3),
  
  -- Promotion 3: คู่หูกาแฟ 120 บาท -> อเมริกาโน่ร้อน (1) + ลาเต้เย็น (4)
  (3, 1), (3, 4),
  
  -- Promotion 4: ซื้อ 3 แก้ว ลด 15% -> ทุกเครื่องดื่ม (1,2,3,4)
  (4, 1), (4, 2), (4, 3), (4, 4),
  
  -- Promotion 5: เค้กลด 15% -> ช็อกโกแลตเค้ก (5), สตรอว์เบอร์รี่ชีสเค้ก (6)
  (5, 5), (5, 6),
  
  -- Promotion 6: ไอศกรีมลด 5 บาท -> ไอศกรีมวานิลลา (7), ไอศกรีมช็อกโกแลต (8)
  (6, 7), (6, 8),
  
  -- Promotion 7: เค้ก+ไอศกรีม 150 บาท -> ช็อกโกแลตเค้ก (5) + ไอศกรีมวานิลลา (7)
  (7, 5), (7, 7),
  
  -- Promotion 8: ไอศกรีม 2 ถ้วย ลด 20 บาท -> ไอศกรีมทั้ง 2 รส (7, 8)
  (8, 7), (8, 8);

-- =========================================================
-- 13) Stock Transfers (pending transfers for testing)
--     Status: SENT = รอสาขารับสินค้า
-- =========================================================
INSERT INTO stock_transfers (id, store_id, from_branch_id, to_branch_id, status, sent_by, sent_at, note, created_at, updated_at) VALUES
  -- Store 1: ส่งสินค้าจากส่วนกลาง (NULL) ไปสาขาสยาม (1) - กำลังส่ง
  (1, 1, NULL, 1, 'SENT', 1, NOW() - INTERVAL '1 day', 'ส่งสินค้าเติมสต็อกประจำสัปดาห์', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  -- Store 1: ส่งสินค้าจากสาขาลาดพร้าว (2) ไปสาขาสยาม (1) - กำลังส่ง
  (2, 1, 2, 1, 'SENT', 1, NOW() - INTERVAL '12 hours', 'โอนสินค้าระหว่างสาขา', NOW() - INTERVAL '1 day', NOW() - INTERVAL '12 hours'),
  -- Store 2: ส่งสินค้าจากส่วนกลาง (NULL) ไปสาขาเซ็นทรัล (3) - กำลังส่ง
  (3, 2, NULL, 3, 'SENT', 4, NOW() - INTERVAL '6 hours', 'เติมสต็อกเค้กและไอศกรีม', NOW() - INTERVAL '1 day', NOW() - INTERVAL '6 hours'),
  -- Store 1: เบิกสินค้าใหม่ไปสาขาสยาม (1) - รอส่ง
  (4, 1, NULL, 1, 'CREATED', NULL, NULL, 'เบิกสินค้าเพิ่มเติม', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours'),
  -- Store 2: เบิกสินค้าใหม่ไปสาขาเซ็นทรัล (3) - รอส่ง
  (5, 2, NULL, 3, 'CREATED', NULL, NULL, 'เบิกไอศกรีมเพิ่ม', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour');

SELECT setval('stock_transfers_id_seq', 5);

-- =========================================================
-- 14) Stock Transfer Items (items in each transfer)
-- =========================================================
INSERT INTO stock_transfer_items (id, stock_transfer_id, product_id, send_count, receive_count) VALUES
  -- Transfer 1: ส่งไปสาขาสยาม (อเมริกาโน่ร้อน, ลาเต้ร้อน) - กำลังส่ง
  (1, 1, 1, 50, 0),  -- อเมริกาโน่ร้อน 50 ชิ้น
  (2, 1, 2, 30, 0),  -- ลาเต้ร้อน 30 ชิ้น
  -- Transfer 2: โอนจากลาดพร้าวไปสยาม (อเมริกาโน่เย็น, ลาเต้เย็น) - กำลังส่ง
  (3, 2, 3, 20, 0),  -- อเมริกาโน่เย็น 20 ชิ้น
  (4, 2, 4, 15, 0),  -- ลาเต้เย็น 15 ชิ้น
  -- Transfer 3: ส่งไปสาขาเซ็นทรัล (เค้ก, ไอศกรีม) - กำลังส่ง
  (5, 3, 5, 10, 0),  -- ช็อกโกแลตเค้ก 10 ชิ้น
  (6, 3, 6, 10, 0),  -- สตรอว์เบอร์รี่ชีสเค้ก 10 ชิ้น
  (7, 3, 7, 50, 0),  -- ไอศกรีมวานิลลา 50 ชิ้น
  (8, 3, 8, 50, 0),  -- ไอศกรีมช็อกโกแลต 50 ชิ้น
  -- Transfer 4: เบิกสินค้าเพิ่มเติมไปสาขาสยาม - รอส่ง
  (9, 4, 1, 25, 0),  -- อเมริกาโน่ร้อน 25 ชิ้น
  (10, 4, 3, 25, 0), -- อเมริกาโน่เย็น 25 ชิ้น
  -- Transfer 5: เบิกไอศกรีมเพิ่มไปสาขาเซ็นทรัล - รอส่ง
  (11, 5, 7, 100, 0), -- ไอศกรีมวานิลลา 100 ชิ้น
  (12, 5, 8, 100, 0); -- ไอศกรีมช็อกโกแลต 100 ชิ้น

SELECT setval('stock_transfer_items_id_seq', 12);

COMMIT;

-- =========================================================
-- Summary:
-- =========================================================
-- Stores: 2
--   1. ร้านกาแฟ เดอะคอฟฟี่
--   2. ร้านขนมหวาน สวีทมี
--
-- Branches: 4 (2 per store)
--   Store 1: สาขาสยาม, สาขาลาดพร้าว
--   Store 2: สาขาเซ็นทรัล, สาขาเมกาบางนา
--
-- Staff Accounts: 6
--   Store 1: 1 manager + 2 staff (1 per branch)
--   Store 2: 1 manager + 2 staff (1 per branch)
--   Login: manager@thecoffee.com / manager@sweetme.com
--   PINs (unique per staff):
--     - manager@thecoffee.com: 1234
--     - staff1@thecoffee.com: 2345
--     - staff2@thecoffee.com: 3456
--     - manager@sweetme.com: 4567
--     - staff1@sweetme.com: 5678
--     - staff2@sweetme.com: 6789
--
-- Categories: 4 (2 per store)
--   Store 1: เครื่องดื่มร้อน, เครื่องดื่มเย็น
--   Store 2: เค้ก, ไอศกรีม
--
-- Products: 8 (2 per category)
--   Store 1: อเมริกาโน่ร้อน, ลาเต้ร้อน, อเมริกาโน่เย็น, ลาเต้เย็น
--   Store 2: ช็อกโกแลตเค้ก, สตรอว์เบอร์รี่ชีสเค้ก, ไอศกรีมวานิลลา, ไอศกรีมช็อกโกแลต
--
-- Promotion Types: 5 per store (10 total)
--   1. ลดเปอร์เซ็นต์ - ลดราคาเป็น % ตรงๆ
--   2. ลดบาท - ลดราคาเป็นจำนวนบาทตรงๆ
--   3. ซื้อเป็นเซ็ต - ซื้อครบเซ็ตได้ราคาพิเศษ (auto-detect)
--   4. ซื้อครบลดเปอร์เซ็นต์ - ซื้อ N ชิ้น ลด N%
--   5. ซื้อครบลดบาท - ซื้อ N ชิ้น ลด N บาท
--
-- Promotions: 6 per store (12 total)
--   Store 1 (ร้านกาแฟ):
--     - ลาเต้ลด 10% (type 1 - product-level)
--     - อเมริกาโน่ลด 10 บาท (type 2 - product-level)
--     - คู่หูกาแฟ 120 บาท (type 3: อเมริกาโน่ร้อน + ลาเต้เย็น)
--     - ซื้อ 3 แก้ว ลด 15% (type 4)
--     - ลดท้ายบิล 5% (type 1 - bill-level)
--     - ลดท้ายบิล 20 บาท (type 2 - bill-level)
--   Store 2 (ร้านขนมหวาน):
--     - เค้กลด 15% (type 1 - product-level)
--     - ไอศกรีมลด 5 บาท (type 2 - product-level)
--     - เค้ก+ไอศกรีม 150 บาท (type 3)
--     - ไอศกรีม 2 ถ้วย ลด 20 บาท (type 5)
--     - ลดท้ายบิล 10% (type 1 - bill-level)
--     - ลดท้ายบิล 15 บาท (type 2 - bill-level)
