# FoodHub

ระบบสั่งอาหารแบบ Microservices (โปรเจกต์ MDD)

## สิ่งที่ต้องมี
- Docker Desktop (พร้อม Docker Compose)
- Node.js (สำหรับ frontend)
- ไม่บังคับ: psql หรือ pgAdmin (ถ้าจะ import SQL เอง)

## เริ่มจากศูนย์ (หลัง pull)
1. Clone แล้วเข้าโฟลเดอร์โปรเจกต์
```powershell
git clone <repo-url>
cd FoodHub
```

2. เปิด backend + databases + rabbitmq + monitoring
```powershell
docker compose up -d --build
```

3. เช็คสถานะ container
```powershell
docker compose ps
```

4. รัน frontend
```powershell
cd frontend
npm install
npm run dev
```

Frontend: http://localhost:5173

## พอร์ตที่ใช้
- Auth: http://localhost:3006
- User: http://localhost:3001
- Menu: http://localhost:3002
- Cart: http://localhost:3005
- Order: http://localhost:3003
- Payment: http://localhost:3004
- RabbitMQ UI: http://localhost:15672 (guest/guest)
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## Health Check
ตัวอย่าง:
```
curl http://localhost:3002/health
```

## Seed เมนู (ถ้าต้องการ)
เราเตรียมไฟล์ SQL ไว้ใน `sql/schema/`

รันผ่าน Docker:
```powershell
docker exec -i menu-db psql -U postgres -d menudb < sql/schema/menu.sql
docker exec -i menu-db psql -U postgres -d menudb < sql/schema/menu_seed.sql
```

หรือใช้ pgAdmin:
1. เปิดฐาน `menudb`
2. Query Tool -> รัน `sql/schema/menu.sql`
3. รัน `sql/schema/menu_seed.sql`

## ถ้าต้องการรีเซ็ตใหม่หมด
```powershell
docker compose down -v --remove-orphans
docker compose up -d --build
```

## หมายเหตุ
- เปิด CORS ทุก service แล้ว เพื่อให้ frontend เรียก API ได้ตรง
