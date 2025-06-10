# 🚀 Perfume API - Redis & RabbitMQ Integration

Hệ thống API bán nước hoa mô phỏng quy trình đặt hàng và giao hàng thực tế, xây dựng theo mô hình **microservice**, tích hợp:

- ✅ Redis – caching phiên đăng nhập & phân quyền người dùng
- ✅ RabbitMQ – giao tiếp giữa các dịch vụ (pub/sub)
- ✅ PostgreSQL – quản lý dữ liệu đơn hàng, người dùng, sản phẩm
- ✅ Gin + GORM – backend API nhanh & rõ ràng
- ✅ Docker Compose – quản lý dịch vụ dễ dàng

---

## 🧠 Kiến trúc tổng quan

lua
Sao chép
Chỉnh sửa
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
yaml
Sao chép
Chỉnh sửa

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
go
Sao chép
Chỉnh sửa
// Middleware xác thực
val := RedisClient.Get(ctx, "session:"+token)
if val != nil {
  // => Gán session vào context
} else {
  // => Truy DB rồi cache lại
}
📦 Mẫu dữ liệu Redis
json
Sao chép
Chỉnh sửa
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

go
Sao chép
Chỉnh sửa
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

json
Sao chép
Chỉnh sửa
{
  "email": "admin@gmail.com",
  "password": "admin123"
}
👉 Trả về token

2. 🛍 Tạo đơn hàng
POST http://localhost:8080/api/orders

json
Sao chép
Chỉnh sửa
{
  "items": [
    { "product_id": 1, "quantity": 2 },
    { "product_id": 2, "quantity": 1 }
  ]
}
Header:

makefile
Sao chép
Chỉnh sửa
Authorization: Bearer <token>
👉 Nếu thành công:

Redis: gia hạn TTL session

RabbitMQ: gửi message delivery

Terminal delivery-service log:

bash
Sao chép
Chỉnh sửa
📦 Đơn hàng #12 - Giao cho user 1: [“Chanel”, “Dior”]
🧰 docker-compose.yml
yaml
Sao chép
Chỉnh sửa
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
🔍 Phân quyền kiểm tra qua Middleware
go
Sao chép
Chỉnh sửa
role := c.GetString("role")
if role != "admin" {
  c.AbortWithStatusJSON(403, gin.H{"error": "Không có quyền"})
}
✅ Tính năng đã hoàn thành
Tính năng	Trạng thái ✅
Redis cache thông tin user	✅
Xác thực JWT + phân quyền	✅
RabbitMQ publish đơn hàng khi tạo	✅
Microservice delivery-service tiêu thụ	✅
Tích hợp Redis, RabbitMQ qua Docker	✅

🧪 Công cụ kiểm tra
Dịch vụ	Link	Ghi chú
RabbitMQ UI	http://localhost:15672	user/pass: guest / guest
Redis CLI	docker exec -it <container> redis-cli	Xem key: keys session:*
PostgreSQL	pgAdmin, TablePlus, psql	DB: go_crud

👨‍💻 Thực hiện bởi
Nguyễn Văn Buri

Thực tập sinh backend – Công ty TNHH Bê Tông Khí ALC

Mục tiêu: Hoàn thiện hệ thống backend + cải tiến hiệu năng với Redis + RabbitMQ

📎 Gợi ý nâng cao
➕ Tích hợp gửi email (gomail)

🔒 Blacklist JWT khi logout (Redis)

📈 Sử dụng Prometheus + Grafana để theo dõi performance

📬 Tách giao tiếp RabbitMQ thành background job

💡 Nếu bạn đang review project này: hãy vào thư mục delivery-service/ và chạy go run main.go để thấy mô phỏng giao hàng realtime nhé!

less
Sao chép
Chỉnh sửa

---

Bạn chỉ cần:

1. Tạo file `README.md` trong thư mục gốc `perfume-api`
2. Paste nội dung trên
3. Commit + push lên GitHub là đẹp 💥

Bạn muốn mình tách thêm README riêng cho từng phần không (`redis.md`, `rabbitmq.md`, `delivery-service.md`)?






