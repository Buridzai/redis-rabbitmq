# 🚀 Perfume API - Redis & RabbitMQ Integration

Hệ thống API bán nước hoa mô phỏng quy trình đặt hàng và giao hàng thực tế, xây dựng theo mô hình **microservice**, tích hợp:

- ✅ Redis – caching phiên đăng nhập & phân quyền người dùng
- ✅ RabbitMQ – giao tiếp giữa các dịch vụ (pub/sub)
- ✅ PostgreSQL – quản lý dữ liệu đơn hàng, người dùng, sản phẩm
- ✅ Gin + GORM – backend API nhanh & rõ ràng
- ✅ Docker Compose – quản lý dịch vụ dễ dàng

---

## 🧠 Kiến trúc tổng quan

            +----------------+
            |  perfume-api   |    (POST /api/orders)
            +----------------+
                    |
         [Redis] <--|--> DB (PostgreSQL)
                    |
                    | Publish (RabbitMQ)
                    ↓
          +---------------------+
          |   delivery-service  |
          +---------------------+
                (log giao hàng)

---

## 📦 Redis - Lưu & kiểm tra phiên đăng nhập

### ✅ Mục tiêu

- Khi user đăng nhập, lưu session vào Redis: `session:<token>`
- Các request sau kiểm tra quyền user bằng cách:
  - Xác thực JWT
  - Nếu token hợp lệ → lấy thông tin user từ Redis
  - Nếu không có → truy DB rồi cache lại

### 🔐 Cách hoạt động

```go
// Khi login
token := GenerateJWT(user.ID, user.Email, user.Role)
session := SessionData{ID: user.ID, Email: user.Email, Role: user.Role}
RedisClient.Set(ctx, "session:"+token, json.Marshal(session), 72*time.Hour)



// Middleware xác thực
val := RedisClient.Get(ctx, "session:"+token)
if val != nil {
  // => Gán session vào context
} else {
  // => Truy DB rồi cache lại
}
📦 Mẫu dữ liệu Redis
{
  "ID": 1,
  "Email": "admin@gmail.com",
  "Role": "admin"
}

📬 RabbitMQ - Giao tiếp giữa các service (Đơn hàng → Giao hàng)
✅ Mục tiêu
Khi đặt hàng thành công → gửi message vào RabbitMQ

delivery-service sẽ nhận message và log giả lập giao hàng

🔄 Cách hoạt động
Bên API (perfume-api):
type DeliveryPayload struct {
  OrderID uint
  UserID  uint
  Items   []string
}

rabbitmq.Publish("delivery-ex", payload)

Bên Microservice (delivery-service):

go
Sao chép
Chỉnh sửa

msg := <-channel.Consume(...)
json.Unmarshal(msg.Body, &payload)
fmt.Println("📦 Giao đơn hàng:", payload)

🧪 Ví dụ test API
1. 🔑 Đăng nhập
POST http://localhost:8080/api/auth/login
{
  "email": "admin@gmail.com",
  "password": "admin123"
}

Trả về token

2. 🛍 Tạo đơn hàng
POST http://localhost:8080/api/orders
{
  "items": [
    { "product_id": 1, "quantity": 2 },
    { "product_id": 2, "quantity": 1 }
  ]
}

Header:
Authorization: Bearer <token>
Nếu thành công:

Redis: gia hạn TTL session

RabbitMQ: gửi message delivery

Terminal delivery-service log:
📦 Đơn hàng #12 - Giao cho user 1: [“Chanel”, “Dior”]

docker-compose.yml
services:
  db:
    image: postgres:15
  redis:
    image: redis:alpine
  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "15672:15672"  # Web UI
  api:
    build: .
    environment:
      - REDIS_ADDR=redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
Phân quyền kiểm tra qua Middleware
role := c.GetString("role")
if role != "admin" {
  c.AbortWithStatusJSON(403, gin.H{"error": "Không có quyền"})
}


