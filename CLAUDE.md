# service-template

Go service ของ ZOVA สร้างจาก service-template ใน GitHub org zova-platform

- ภาพรวมระบบ: repo `docs` ไฟล์ `docs/architecture/overview.md`
- spec ของ service นี้: repo `docs` ไฟล์ `docs/repos.md` หัวข้อ service-template
- วิธีรัน env และ API: README.md

กติกา
- Conventional Commits บรรทัดเดียว ไม่ใส่ Co-Authored-By 1 commit ต้อง build และ test ผ่าน (`task check`)
- ไม่ใส่ comment ในโค้ด
- ใช้ platform-go สำหรับ config, logging, tracing, gRPC, health และ graceful shutdown ห้ามเขียนซ้ำใน service
- business logic อยู่ใน `internal/domain` และ `internal/app` เท่านั้น `internal/grpc` แค่แปลง proto กับ domain
- แต่ละ service เป็นเจ้าของ database ของตัวเอง ห้ามต่อ database ของ service อื่น คุยกันผ่าน gRPC หรือ event
- เปลี่ยน proto แบบ breaking ไม่ได้ ถ้าจำเป็นให้ขึ้น package ใหม่
- ตรวจเวอร์ชันล่าสุดของ library และ GitHub Actions ก่อนใช้ทุกครั้ง
- ห้าม commit secret หรือไฟล์ .env
