
# Perfume API – Redis & RabbitMQ Integration

---

## 🧠 Kiến trúc tổng quan

```
           +---------------------+
           |    perfume-api      |     (POST /api/orders)
           +---------------------+
                    |
        [Redis] <---> DB (PostgreSQL)
                    |
              ↑     |
              |  Publish (RabbitMQ)
              ↓     |
           +----------------------+
           |   delivery-service   |  (sub -> ghi log giao hàng)
           +----------------------+
```

---

## 🔐 Redis – Quản lý Session Người Dùng

### ✅ Mục tiêu

- Sau khi đăng nhập, lưu session vào Redis:  
  `session:<JWT>` → user info (name, role, permissions,...)
- Trong middleware, mỗi request:
  - Kiểm tra token hợp lệ
  - Nếu có cache → lấy từ Redis
  - Nếu không → truy DB và cache lại

### ✅ Ví dụ xử lý

```go
// Khi đăng nhập thành công (Login Handler)
sessionData := SessionData{
    ID:    user.ID,
    Name:  user.Name,
    Email: user.Email,
    Role:  user.Role,
}
jsonData, _ := json.Marshal(sessionData)
RedisClient.Set(ctx, "session:"+token, jsonData, 72*time.Hour)
```

```go
// Trong middleware JWT
val := RedisClient.Get(ctx, "session:"+token)
if val != nil {
    // Parse -> lấy sessionData
    json.Unmarshal([]byte(val), &session)
    c.Set("user_id", session.ID)
    c.Next()
}
```

---

## 📨 RabbitMQ – Gửi Đơn Hàng

### ✅ Mục tiêu

- Khi user đặt hàng → publish message lên RabbitMQ
- Microservice `delivery-service` lắng nghe và log thông tin giao hàng

### ✅ Cấu hình RabbitMQ

```go
// perfume-api/utils/rabbitmq/rabbitmq.go
func Publish(exchange string, message []byte) error {
    ch := rabbitmqConn.Channel
    return ch.Publish(exchange, "", false, false, amqp.Publishing{
        ContentType: "application/json",
        Body:        message,
    })
}
```

```go
// delivery-service/main.go
msgs, _ := ch.Consume(queueName, "", true, false, false, false, nil)
for msg := range msgs {
    fmt.Println("📦 Giao đơn hàng:", string(msg.Body))
}
```

---

##  test API

### 1. Đăng nhập

```http
POST /api/auth/login
```

```json
{
  "email": "admin@gmail.com",
  "password": "admin123"
}
```

### 2. Gọi đơn hàng

```http
POST /api/orders
Authorization: Bearer <token>
```

```json
{
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ]
}
```

### 3. Kiểm tra

- Redis có session:
  ```bash
  redis-cli
  > KEYS *
  > GET session:<token>
  ```

- Terminal delivery-service:
  ```bash
  📦 Giao đơn hàng: {"user_id":1, "items": [...]}
  ```

---

## 🐳 Docker Compose

```yaml
services:
  db:
    image: postgres:15

  redis:
    image: redis:alpine

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "15672:15672" # UI: http://localhost:15672

  api:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
      - redis
      - rabbitmq
    environment:
      - DB_HOST=db
      - REDIS_ADDR=redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
      - JWT_SECRET=your-secret-key
```

---

## 🧰 RabbitMQ UI

- Truy cập: http://localhost:15672  
- Tài khoản: `guest / guest`  
- Xem exchange: `delivery-ex`

---

## 📌 Tổng kết

| Tính năng         | Mô tả                                     | Trạng thái |
|------------------|--------------------------------------------|------------|
| Redis session     | Lưu user login, giảm truy DB               | ✅         |
| Middleware JWT    | Check token & phân quyền                   | ✅         |
| RabbitMQ publish  | Gửi message khi tạo đơn                    | ✅         |
| RabbitMQ consume  | Service khác nhận và xử lý đơn hàng        | ✅         |
| Docker Compose    | Quản lý toàn bộ service                    | ✅         |

