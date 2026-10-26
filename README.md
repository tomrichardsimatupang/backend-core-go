# backend-core-go

REST API berbasis Go dengan arsitektur **feature-first**, autentikasi **JWT**, dan otorisasi **RBAC** (role & permission).

- **Go** 1.26+ · **Gin** · **PostgreSQL** (pgx v5 + Squirrel, tanpa ORM)
- Migrasi SQL dengan **golang-migrate**
- Logging terstruktur (`slog`) dengan `request_id` di setiap request
- Fitur bawaan: `auth`, `rbac`, `user`, `product` (contoh ownership check)

Panduan arsitektur dan konvensi kode ada di [CLAUDE.md](CLAUDE.md).

---

## Daftar Isi

- [Menjalankan dengan Docker](#menjalankan-dengan-docker)
- [Menjalankan tanpa Docker](#menjalankan-tanpa-docker)
- [Konfigurasi (Environment Variable)](#konfigurasi-environment-variable)
- [Langkah Pertama: Membuat Admin](#langkah-pertama-membuat-admin)
- [Dokumentasi REST API](#dokumentasi-rest-api)
  - [Format Response](#format-response)
  - [Autentikasi](#autentikasi)
  - [Endpoint](#endpoint)
  - [Kode Error](#kode-error)
- [Permission](#permission)
- [Development](#development)

---

## Menjalankan dengan Docker

Cukup butuh **Docker** + **Docker Compose**. Tidak perlu install Go.

```bash
# 1. Siapkan konfigurasi
cp .env.example .env
sed -i "s/^JWT_SECRET=.*/JWT_SECRET=$(openssl rand -hex 32)/" .env

# 2. Jalankan seluruh stack: postgres → migrate → api
docker compose up -d --build

# 3. Cek
curl localhost:8080/healthz
# {"data":{"status":"ok"}}
```

Urutan start diatur otomatis: `postgres` (tunggu healthy) → `migrate` (jalankan semua migrasi, lalu berhenti) → `api`.

| Perintah | Fungsi |
| --- | --- |
| `docker compose up -d --build` | Build ulang image & jalankan stack |
| `docker compose logs -f api` | Lihat log API |
| `docker compose run --rm seed` | Sinkronisasi permission & role default |
| `docker compose down` | Stop stack (data database tetap tersimpan) |
| `docker compose down -v` | Stop stack **dan hapus data** database |

Port bisa diubah lewat `.env` bila bentrok: `API_PORT` (default `8080`) dan `POSTGRES_PORT` (default `5432`).

---

## Menjalankan tanpa Docker

### Prasyarat

- Go 1.26+
- PostgreSQL 14+ (lokal, atau cukup `docker compose up -d postgres`)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- (Opsional) [air](https://github.com/air-verse/air) untuk hot reload, [golangci-lint](https://golangci-lint.run) untuk lint

```bash
# Install Go (Linux amd64)
curl -LO https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc && source ~/.bashrc

# Install tools
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest
```

### Langkah

```bash
# 1. Konfigurasi (Makefile otomatis membaca .env)
cp .env.example .env
sed -i "s/^JWT_SECRET=.*/JWT_SECRET=$(openssl rand -hex 32)/" .env
# Sesuaikan DATABASE_URL bila Postgres bukan di localhost:5432

# 2. Database
docker compose up -d postgres   # lewati jika memakai Postgres sendiri
make migrate-up

# 3. Jalankan API
make run        # atau: make dev (hot reload)
```

---

## Konfigurasi (Environment Variable)

| Variabel | Default | Keterangan |
| --- | --- | --- |
| `APP_ENV` | `development` | `production` mengaktifkan Gin release mode |
| `HTTP_ADDR` | `:8080` | Alamat listen server |
| `SHUTDOWN_TIMEOUT` | `10s` | Batas waktu graceful shutdown |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` atau `text` |
| `DATABASE_URL` | — (**wajib**) | DSN Postgres. Di Docker Compose dibentuk otomatis |
| `JWT_SECRET` | — (**wajib**) | Minimal 32 karakter |
| `JWT_TTL` | `24h` | Masa berlaku access token |
| `JWT_ISSUER` | `backend-core-go` | Klaim `iss` pada token |
| `RBAC_CACHE_TTL` | `5m` | Lama cache permission per user (`0` = nonaktif) |
| `SEED_ADMIN_EMAIL` | — | Email user yang dijadikan `super_admin` saat seed |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | `postgres` / `postgres` / `backend_core` | Khusus Docker Compose |
| `POSTGRES_PORT` / `API_PORT` | `5432` / `8080` | Port yang dibuka ke host (Docker Compose) |

---

## Langkah Pertama: Membuat Admin

User baru belum punya role apa pun, sehingga hanya bisa mengakses `/users/me`. Untuk membuat admin pertama:

```bash
# 1. Register
curl -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Admin","email":"admin@example.com","password":"secret123"}'

# 2. Seed: sinkronkan permission & jadikan user tersebut super_admin
SEED_ADMIN_EMAIL=admin@example.com docker compose run --rm seed   # dengan Docker
SEED_ADMIN_EMAIL=admin@example.com make seed                      # tanpa Docker

# 3. Login untuk mendapatkan token
curl -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"secret123"}'
```

Seed aman dijalankan berulang kali. Jalankan ulang setiap ada permission baru di kode.

> Role `super_admin` selalu mendapat **semua** permission dan tidak bisa diubah/dihapus lewat API.

---

## Dokumentasi REST API

Base URL: `http://localhost:8080/api/v1`

### Format Response

**Sukses**

```json
{ "data": { ... } }
```

**Sukses dengan pagination** (endpoint list)

```json
{
  "data": [ ... ],
  "meta": { "page": 1, "per_page": 20, "total": 45, "total_pages": 3 }
}
```

**Error**

```json
{ "error": { "code": "USER_NOT_FOUND", "message": "user not found" } }
```

**Error validasi** (`422`)

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "request validation failed",
    "details": [
      { "field": "email", "rule": "email" },
      { "field": "password", "rule": "min", "param": "8" }
    ]
  }
}
```

**Pagination**: endpoint list menerima query `page` (default `1`) dan `per_page` (default `20`, maks `100`).

Setiap response menyertakan header `X-Request-ID`. Kirim header yang sama dari client untuk melacak request di log server.

### Autentikasi

Endpoint terproteksi membutuhkan header:

```
Authorization: Bearer <access_token>
```

Token didapat dari `register` atau `login`. Token hanya berisi `user_id`; role dan permission selalu dicek ke database (dengan cache), sehingga perubahan role langsung berlaku tanpa perlu login ulang.

### Endpoint

Kolom **Permission** menunjukkan permission yang dibutuhkan. Jika lebih dari satu, user cukup memiliki **salah satunya**.

#### Health

| Method | Path | Permission | Keterangan |
| --- | --- | --- | --- |
| `GET` | `/healthz` | publik | Cek API & koneksi database (di luar `/api/v1`) |

#### Auth

| Method | Path | Permission | Keterangan |
| --- | --- | --- | --- |
| `POST` | `/auth/register` | publik | Daftar user baru, langsung mendapat token |
| `POST` | `/auth/login` | publik | Login, mendapat token |

<details>
<summary>Contoh request & response</summary>

`POST /auth/register`

```json
{ "name": "Budi", "email": "budi@example.com", "password": "secret123" }
```

| Field | Aturan |
| --- | --- |
| `name` | wajib, 2–100 karakter |
| `email` | wajib, format email, maks 255 (disimpan lowercase) |
| `password` | wajib, 8–72 karakter |

`POST /auth/login`

```json
{ "email": "budi@example.com", "password": "secret123" }
```

Response (`201` untuk register, `200` untuk login):

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_at": "2026-10-27T03:34:43Z"
  }
}
```

</details>

#### Users

| Method | Path | Permission | Keterangan |
| --- | --- | --- | --- |
| `GET` | `/users/me` | login | Profil user yang sedang login |
| `PUT` | `/users/me` | login | Ubah profil sendiri |
| `GET` | `/users` | `user:read` | List user (paginated) |
| `POST` | `/users` | `user:create` | Buat user |
| `GET` | `/users/:id` | `user:read` | Detail user |
| `PUT` | `/users/:id` | `user:update` | Ubah user |
| `DELETE` | `/users/:id` | `user:delete` | Hapus user (tidak bisa menghapus diri sendiri) |

<details>
<summary>Contoh request & response</summary>

`POST /users`

```json
{ "name": "Ani", "email": "ani@example.com", "password": "secret123" }
```

`PUT /users/me` atau `PUT /users/:id`: semua field opsional, hanya field yang dikirim yang diubah.

```json
{ "name": "Ani Wijaya", "email": "ani.w@example.com" }
```

Response user:

```json
{
  "data": {
    "id": 2,
    "name": "Ani Wijaya",
    "email": "ani.w@example.com",
    "created_at": "2026-10-26T03:34:43Z",
    "updated_at": "2026-10-26T04:10:00Z"
  }
}
```

</details>

#### Products

| Method | Path | Permission | Keterangan |
| --- | --- | --- | --- |
| `GET` | `/products` | `product:read` | List produk (paginated) |
| `POST` | `/products` | `product:create` | Buat produk (pemilik = user login) |
| `GET` | `/products/:id` | `product:read` | Detail produk |
| `PUT` | `/products/:id` | `product:update` / `product:manage` | Ubah produk |
| `DELETE` | `/products/:id` | `product:delete` / `product:manage` | Hapus produk |

**Aturan kepemilikan:** user dengan `product:update` / `product:delete` hanya bisa mengubah/menghapus **produk miliknya sendiri**. User dengan `product:manage` bisa mengubah/menghapus produk milik siapa pun. Selain itu → `403 PRODUCT_FORBIDDEN`.

Query tambahan pada `GET /products`:

| Query | Keterangan |
| --- | --- |
| `q` | Cari berdasarkan nama (case-insensitive) |
| `owner_id` | Filter berdasarkan pemilik |

<details>
<summary>Contoh request & response</summary>

`POST /products`

```json
{ "name": "Kopi Arabika", "description": "250 gram", "price": 85000 }
```

| Field | Aturan |
| --- | --- |
| `name` | wajib, 1–200 karakter |
| `description` | opsional, maks 2000 karakter |
| `price` | wajib, bilangan bulat ≥ 0 (dalam satuan terkecil, mis. rupiah) |

`PUT /products/:id`: semua field opsional.

```json
{ "price": 90000 }
```

Response:

```json
{
  "data": {
    "id": 1,
    "owner_id": 2,
    "name": "Kopi Arabika",
    "description": "250 gram",
    "price": 90000,
    "created_at": "2026-10-26T03:34:43Z",
    "updated_at": "2026-10-26T04:10:00Z"
  }
}
```

</details>

#### RBAC

| Method | Path | Permission | Keterangan |
| --- | --- | --- | --- |
| `GET` | `/rbac/permissions` | `role:read` | Semua permission yang tersedia |
| `GET` | `/rbac/roles` | `role:read` | List role |
| `POST` | `/rbac/roles` | `role:manage` | Buat role |
| `GET` | `/rbac/roles/:id` | `role:read` | Detail role beserta permission-nya |
| `DELETE` | `/rbac/roles/:id` | `role:manage` | Hapus role |
| `PUT` | `/rbac/roles/:id/permissions` | `role:manage` | Ganti seluruh permission role |
| `GET` | `/rbac/users/:id/roles` | `role:read` | Role milik user |
| `POST` | `/rbac/users/:id/roles` | `role:manage` | Berikan role ke user |
| `DELETE` | `/rbac/users/:id/roles/:roleId` | `role:manage` | Cabut role dari user |

<details>
<summary>Contoh: membuat role "member" dan memberikannya ke user</summary>

```bash
TOKEN=<token admin>

# Buat role (nama disimpan lowercase)
curl -X POST localhost:8080/api/v1/rbac/roles \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"member","description":"Kelola produk sendiri"}'
# → {"data":{"id":2,"name":"member",...}}

# Atur permission role (mengganti seluruh daftar)
curl -X PUT localhost:8080/api/v1/rbac/roles/2/permissions \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"permissions":["product:read","product:create","product:update","product:delete"]}'
# → 204 No Content

# Berikan role ke user id 3
curl -X POST localhost:8080/api/v1/rbac/users/3/roles \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"role_id":2}'
# → 204 No Content
```

`GET /rbac/roles/2`:

```json
{
  "data": {
    "id": 2,
    "name": "member",
    "description": "Kelola produk sendiri",
    "permissions": ["product:create", "product:delete", "product:read", "product:update"],
    "created_at": "2026-10-26T03:34:43Z"
  }
}
```

</details>

### Kode Error

| HTTP | Code | Penyebab |
| --- | --- | --- |
| 400 | `BAD_REQUEST` | Body bukan JSON yang valid |
| 400 | `INVALID_ID` | Path param ID bukan angka positif |
| 401 | `MISSING_TOKEN` | Header `Authorization: Bearer` tidak ada |
| 401 | `INVALID_TOKEN` | Token tidak valid atau kedaluwarsa |
| 401 | `INVALID_CREDENTIALS` | Email atau password salah |
| 403 | `FORBIDDEN` | Tidak punya permission yang dibutuhkan |
| 403 | `PRODUCT_FORBIDDEN` | Mengubah produk milik orang lain tanpa `product:manage` |
| 403 | `CANNOT_DELETE_SELF` | Menghapus akun sendiri |
| 403 | `PROTECTED_ROLE` | Mengubah/menghapus role `super_admin` |
| 404 | `USER_NOT_FOUND` / `PRODUCT_NOT_FOUND` / `ROLE_NOT_FOUND` | Data tidak ditemukan |
| 404 | `ROUTE_NOT_FOUND` | Endpoint tidak ada |
| 409 | `EMAIL_TAKEN` | Email sudah terdaftar |
| 409 | `ROLE_EXISTS` | Nama role sudah dipakai |
| 422 | `VALIDATION_ERROR` | Field request tidak valid (lihat `details`) |
| 422 | `UNKNOWN_PERMISSION` | Permission yang dikirim tidak ada |
| 500 | `INTERNAL_ERROR` | Kesalahan server (detail hanya dicatat di log) |
| 503 | `UNHEALTHY` | Database tidak dapat dijangkau (`/healthz`) |

---

## Permission

| Permission | Akses |
| --- | --- |
| `user:read` | Lihat daftar & detail user |
| `user:create` | Buat user |
| `user:update` | Ubah user mana pun |
| `user:delete` | Hapus user |
| `product:read` | Lihat produk |
| `product:create` | Buat produk |
| `product:update` | Ubah produk **milik sendiri** |
| `product:delete` | Hapus produk **milik sendiri** |
| `product:manage` | Ubah/hapus produk **milik siapa pun** |
| `role:read` | Lihat role, permission, dan role user |
| `role:manage` | Kelola role & assign/revoke role ke user |

---

## Development

| Perintah | Fungsi |
| --- | --- |
| `make run` | Jalankan API |
| `make dev` | Jalankan dengan hot reload (air) |
| `make build` | Build binary ke `bin/api` |
| `make test` | Unit test (`go test ./... -race`) |
| `make lint` | `golangci-lint run` |
| `make migrate-up` | Jalankan semua migrasi |
| `make migrate-down` | Rollback 1 migrasi |
| `make migrate-new name=create_xxx` | Buat file migrasi baru |
| `make seed` | Sinkronisasi permission & role default |

### Struktur Singkat

```
cmd/api            entry point API
cmd/seed           sinkronisasi permission & role
internal/app       server, wiring route, daftar permission
internal/config    konfigurasi dari env
internal/platform  database (pgxpool, transaksi), logger
internal/shared    httpx (response/error), middleware, pagination, authctx, apperr
internal/features  auth, rbac, user, product — satu folder per fitur
migrations         file SQL up/down
```

Untuk menambah feature baru, ikuti checklist di [CLAUDE.md](CLAUDE.md#menambah-feature-baru-checklist).
