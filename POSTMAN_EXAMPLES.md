# Postman Examples - Transaction API

## Base URL
```
http://localhost:8080
```

## Authentication
All protected endpoints require JWT token in the Authorization header:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## 1. เพิ่มแต้ม - สินค้าอย่างเดียว (1.0 ลิตร)

### Endpoint
```
POST /api/v1/transactions
```

### Headers
```
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

### Body
```json
{
  "member_id": "123e4567-e89b-12d3-a456-426614174000",
  "action": "EARN",
  "products": [
    {
      "product_type": "1_0_LITER",
      "points": 10
    }
  ],
  "receipt_text": "ซื้อน้ำมัน 1.0L จำนวน 1 ขวด"
}
```

### Response
```json
{
  "transactions": [
    {
      "id": "987fcdeb-51a2-43e1-b456-789012345678",
      "member_id": "123e4567-e89b-12d3-a456-426614174000",
      "staff_user_id": "staff-uuid-here",
      "action": "EARN",
      "product_type": "1_0_LITER",
      "points": 10,
      "receipt_text": "ซื้อน้ำมัน 1.0L จำนวน 1 ขวด",
      "created_at": "2026-01-16T15:30:00Z"
    }
  ],
  "total_points": 10,
  "message": "Successfully processed 1 product(s) with 10 total points"
}
```

---

## 2. เพิ่มแต้ม - สินค้า 2 ประเภทพร้อมกัน

### Endpoint
```
POST /api/v1/transactions
```

### Headers
```
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

### Body
```json
{
  "member_id": "123e4567-e89b-12d3-a456-426614174000",
  "action": "EARN",
  "products": [
    {
      "product_type": "1_0_LITER",
      "points": 10
    },
    {
      "product_type": "1_5_LITER",
      "points": 15
    }
  ],
  "receipt_text": "ซื้อน้ำมัน 1.0L x1 และ 1.5L x1"
}
```

### Response
```json
{
  "transactions": [
    {
      "id": "trans-1-uuid",
      "member_id": "123e4567-e89b-12d3-a456-426614174000",
      "staff_user_id": "staff-uuid-here",
      "action": "EARN",
      "product_type": "1_0_LITER",
      "points": 10,
      "receipt_text": "ซื้อน้ำมัน 1.0L x1 และ 1.5L x1",
      "created_at": "2026-01-16T15:30:00Z"
    },
    {
      "id": "trans-2-uuid",
      "member_id": "123e4567-e89b-12d3-a456-426614174000",
      "staff_user_id": "staff-uuid-here",
      "action": "EARN",
      "product_type": "1_5_LITER",
      "points": 15,
      "receipt_text": "ซื้อน้ำมัน 1.0L x1 และ 1.5L x1",
      "created_at": "2026-01-16T15:30:00Z"
    }
  ],
  "total_points": 25,
  "message": "Successfully processed 2 product(s) with 25 total points"
}
```

---

## 3. เพิ่มแต้ม - สินค้าเดียวกันหลายจำนวน (1.5 ลิตร)

### Endpoint
```
POST /api/v1/transactions
```

### Body
```json
{
  "member_id": "123e4567-e89b-12d3-a456-426614174000",
  "action": "EARN",
  "products": [
    {
      "product_type": "1_5_LITER",
      "points": 50
    }
  ],
  "receipt_text": "ซื้อน้ำมัน 1.5L จำนวน 5 ขวด (10 points/ขวด)"
}
```

---

## 4. แลกแต้ม - สินค้า 1 ประเภท

### Endpoint
```
POST /api/v1/transactions
```

### Body
```json
{
  "member_id": "123e4567-e89b-12d3-a456-426614174000",
  "action": "REDEEM",
  "products": [
    {
      "product_type": "1_0_LITER",
      "points": 100
    }
  ],
  "receipt_text": "แลกของรางวัล"
}
```

### Response
```json
{
  "transactions": [
    {
      "id": "redeem-trans-uuid",
      "member_id": "123e4567-e89b-12d3-a456-426614174000",
      "staff_user_id": "staff-uuid-here",
      "action": "REDEEM",
      "product_type": "1_0_LITER",
      "points": 100,
      "receipt_text": "แลกของรางวัล",
      "created_at": "2026-01-16T15:35:00Z"
    }
  ],
  "total_points": 100,
  "message": "Successfully processed 1 product(s) with 100 total points"
}
```

---

## 5. แลกแต้ม - หลายประเภทพร้อมกัน

### Endpoint
```
POST /api/v1/transactions
```

### Body
```json
{
  "member_id": "123e4567-e89b-12d3-a456-426614174000",
  "action": "REDEEM",
  "products": [
    {
      "product_type": "1_0_LITER",
      "points": 50
    },
    {
      "product_type": "1_5_LITER",
      "points": 75
    }
  ],
  "receipt_text": "แลกของรางวัลพิเศษ"
}
```

### Response
```json
{
  "transactions": [
    {
      "id": "redeem-1-uuid",
      "member_id": "123e4567-e89b-12d3-a456-426614174000",
      "staff_user_id": "staff-uuid-here",
      "action": "REDEEM",
      "product_type": "1_0_LITER",
      "points": 50,
      "receipt_text": "แลกของรางวัลพิเศษ",
      "created_at": "2026-01-16T15:40:00Z"
    },
    {
      "id": "redeem-2-uuid",
      "member_id": "123e4567-e89b-12d3-a456-426614174000",
      "staff_user_id": "staff-uuid-here",
      "action": "REDEEM",
      "product_type": "1_5_LITER",
      "points": 75,
      "receipt_text": "แลกของรางวัลพิเศษ",
      "created_at": "2026-01-16T15:40:00Z"
    }
  ],
  "total_points": 125,
  "message": "Successfully processed 2 product(s) with 125 total points"
}
```

---

## สรุป Format

### Request Format
```json
{
  "member_id": "UUID",
  "action": "EARN" หรือ "REDEEM",
  "products": [
    {
      "product_type": "1_0_LITER" หรือ "1_5_LITER",
      "points": จำนวนแต้ม (integer > 0)
    }
  ],
  "receipt_text": "ข้อความบนใบเสร็จ (optional)"
}
```

### Response Format
```json
{
  "transactions": [
    {
      "id": "UUID",
      "member_id": "UUID",
      "staff_user_id": "UUID",
      "action": "EARN" หรือ "REDEEM",
      "product_type": "1_0_LITER" หรือ "1_5_LITER",
      "points": จำนวนแต้ม,
      "receipt_text": "ข้อความ",
      "created_at": "timestamp"
    }
  ],
  "total_points": รวมแต้มทั้งหมด,
  "message": "ข้อความแจ้งผล"
}
```

---

## การคำนวณแต้ม

### EARN (เพิ่มแต้ม)
- `total_points` += แต้มรวมจากทุก product
- `milestone_score` += แต้มรวมจากทุก product
- `points_1_0_liter` += แต้มจาก product type `1_0_LITER`
- `points_1_5_liter` += แต้มจาก product type `1_5_LITER`

### REDEEM (แลกแต้ม)
- `total_points` -= แต้มรวมจากทุก product
- `milestone_score` ไม่เปลี่ยนแปลง
- `points_1_0_liter` ไม่เปลี่ยนแปลง
- `points_1_5_liter` ไม่เปลี่ยนแปลง

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Key: 'TransactionCreateRequest.Products' Error:Field validation for 'Products' failed on the 'required' tag"
}
```

### 401 Unauthorized
```json
{
  "error": "authorization header required"
}
```

### 404 Not Found
```json
{
  "error": "member not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "database error message"
}
```
