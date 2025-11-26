# E-Commerce Platform Integration System

Hệ thống tích hợp đa nền tảng e-commerce sử dụng Kafka và Go Lang, tương tự kiến trúc trong diagram.

## Kiến trúc

### 1. Webhooks Layer
- **webhooks-api** (ant-kafka): Nhận webhooks từ các platform và gửi vào Kafka
- **webhooks-enrich** (ant-enrich): Xử lý và làm giàu dữ liệu từ Kafka

### 2. API Gateway Services
- **order-service**: Quản lý đơn hàng
- **catalog-service**: Quản lý sản phẩm
- **settings-service**: Cấu hình hệ thống
- **report-service**: Báo cáo và thống kê

### 3. Platform Connectors
- **firefly-connector**: Kết nối với Core (MSI)
- **mantis-connector**: Kết nối với BigCommerce
- **ladybug-connector**: Kết nối với Magento
- **hermes-connector**: Kết nối với Shopify
- **dragonfly-connector**: Kết nối với NetSuite
- **locust-connector**: Kết nối với Kidzania

## Cấu trúc Project

```
.
├── cmd/
│   ├── webhooks-api/          # Webhooks API service
│   ├── webhooks-enrich/       # Webhooks enrichment service
│   ├── order-service/          # Order management service
│   ├── catalog-service/       # Catalog management service
│   ├── settings-service/      # Settings service
│   ├── report-service/         # Report service
│   ├── firefly-connector/      # MSI connector
│   ├── mantis-connector/       # BigCommerce connector
│   ├── ladybug-connector/      # Magento connector
│   ├── hermes-connector/       # Shopify connector
│   ├── dragonfly-connector/    # NetSuite connector
│   └── locust-connector/       # Kidzania connector
├── internal/
│   ├── models/                 # Shared data models
│   ├── config/                 # Configuration management
│   └── kafka/                  # Kafka client utilities
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## Cách chạy

### 1. Build và chạy với Docker Compose

```bash
docker compose build
docker compose up -d
```

### 2. Test Webhooks API

Gửi webhook từ Shopify:
```bash
curl -X POST http://localhost:8080/webhooks/shopify \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "order.created",
    "order": {
      "id": "12345",
      "total": 99.99,
      "status": "pending"
    }
  }'
```

Gửi webhook từ Magento:
```bash
curl -X POST http://localhost:8080/webhooks/magento \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "product.updated",
    "product": {
      "id": "prod-001",
      "name": "Test Product",
      "sku": "SKU-001",
      "price": 49.99
    }
  }'
```

### 3. Health Check

```bash
# Webhooks API
curl http://localhost:8080/health

# Order Service
curl http://localhost:8081/health

# Catalog Service
curl http://localhost:8082/health

# Settings Service
curl http://localhost:8083/health

# Report Service
curl http://localhost:8084/health
```

## Luồng dữ liệu

1. **Webhook nhận vào** → `webhooks-api` nhận webhook từ platform
2. **Gửi vào Kafka** → `webhooks-api` gửi event vào topic `webhooks`
3. **Enrichment** → `webhooks-enrich` đọc từ `webhooks`, enrich và gửi vào `webhooks-enriched`
4. **API Services** → Các service (order, catalog, report) đọc từ `webhooks-enriched` và xử lý
5. **Platform Connectors** → Các connector đọc orders và gửi đến platform tương ứng

## Kafka Topics

- `webhooks`: Webhook events từ các platform
- `webhooks-enriched`: Webhook events đã được enrich
- `orders`: Order events đã được xử lý

## Environment Variables

- `KAFKA_BROKER`: Kafka broker address (default: localhost:9092)
- `KAFKA_TOPIC`: Kafka topic name (default: webhooks)
- `HTTP_PORT`: HTTP server port (default: 8080)
- `SERVICE_NAME`: Service name for logging

## Mở rộng

Để thêm platform connector mới:
1. Tạo service mới trong `cmd/`
2. Sử dụng `kafka.NewConsumer()` để đọc từ topic `orders`
3. Implement logic gửi đến platform tương ứng
4. Thêm vào `docker-compose.yml`

