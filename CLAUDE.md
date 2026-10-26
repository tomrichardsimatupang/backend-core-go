# CLAUDE.md

Panduan untuk Claude saat bekerja di repository ini. Ikuti konvensi di bawah untuk setiap perubahan kode.

## Ringkasan Project

REST API berbasis Go dengan arsitektur **feature-first** (satu folder per fitur).

- **Bahasa:** Go 1.26+ (versi minimum dari dependency terbaru: gin 1.12, validator 10.30)
- **HTTP framework:** Gin (`github.com/gin-gonic/gin`)
- **Database:** PostgreSQL via `github.com/jackc/pgx/v5` (`pgxpool`)
- **Query builder:** Squirrel (`github.com/Masterminds/squirrel`) — **tanpa ORM**
- **Migrasi:** `golang-migrate` (SQL manual)
- **Auth:** JWT (feature `auth`) + RBAC (feature `rbac`)
- **Hot reload:** `air`

## Perintah

```bash
make run            # jalankan API
make dev            # jalankan dengan hot reload (air)
make test           # go test ./... -race
make lint           # golangci-lint run
make migrate-up     # jalankan semua migrasi
make migrate-down   # rollback 1 migrasi
make migrate-new name=create_xxx   # buat file migrasi baru
make seed           # sinkronisasi permission & role default
docker compose up -d postgres      # Postgres lokal
docker compose up -d --build      # full stack: postgres → migrate → api (:8080)
docker compose run --rm seed       # seed di dalam compose (SEED_ADMIN_EMAIL dari .env)
```

Selalu jalankan `make test` dan `make lint` setelah mengubah kode.

## Struktur Folder

```
cmd/api/main.go                 # entry point: config, DB, wiring, start server
internal/
├── app/
│   ├── server.go               # setup Gin, middleware global, graceful shutdown
│   ├── routes.go               # wiring semua feature + middleware auth/rbac
│   └── permissions.go          # daftar Permissions semua feature → rbac.AllPermissions
├── config/config.go            # baca env
├── platform/                   # infrastruktur
│   ├── database/
│   │   ├── postgres.go         # pgxpool
│   │   └── tx.go               # interface DBTX + helper transaksi
│   └── logger/logger.go
├── shared/                     # dipakai lintas feature, TANPA business logic
│   ├── apperr/apperr.go        # tipe error domain (Kind + Code + Message)
│   ├── authctx/context.go      # SetUserID / UserIDFrom
│   ├── httpx/
│   │   ├── response.go         # OK, Created, NoContent, BadRequest
│   │   └── errors.go           # mapping error domain → HTTP status
│   ├── middleware/             # request_id, logger, recovery
│   └── pagination/
└── features/
    ├── auth/                   # autentikasi (JWT)
    │   ├── dto.go, token.go, service.go, handler.go
    │   ├── middleware.go       # Authenticate(): verifikasi JWT → authctx
    │   └── routes.go
    ├── rbac/                   # otorisasi
    │   ├── model.go, dto.go, errors.go
    │   ├── permissions.go      # role:read, role:manage
    │   ├── registry.go         # AllPermissions(groups...): gabung + validasi format
    │   ├── repository.go, cache.go, service.go
    │   ├── middleware.go       # Require(perms...)
    │   ├── seeder.go
    │   ├── handler.go, routes.go
    ├── user/
    └── product/
migrations/                     # NNNNNN_nama.up.sql / .down.sql
```

### Isi standar satu feature

| File | Tanggung jawab |
| --- | --- |
| `model.go` | Entity, representasi tabel (tag `db:"..."`) |
| `dto.go` | Request/response struct + tag `binding` untuk validasi |
| `errors.go` | Error domain (`ErrXxxNotFound`, dll) |
| `permissions.go` | Konstanta permission feature ini + slice `Permissions` |
| `repository.go` | Interface `Repository` + implementasi Squirrel/pgx |
| `service.go` | Interface `Service` + business logic |
| `handler.go` | Handler Gin: bind, panggil service, kirim response |
| `routes.go` | `RegisterRoutes(rg, db, authz)` |
| `*_test.go` | Unit test service (mock repository) |

## Aturan Arsitektur

1. **Alur dependensi satu arah:** `handler → service → repository → database`. Tidak boleh dilompati (handler tidak memanggil repository).
2. **Service dan repository diekspos sebagai interface**, constructor mengembalikan interface: `func NewService(repo Repository) Service`.
3. **Feature tidak saling import implementasi.** Jika butuh data feature lain, definisikan interface kecil di feature pemakai. Pengecualian: konstanta `Permissions` dan `*rbac.Middleware`.
4. **`shared/` tidak boleh import `features/`.**
5. **Handler tipis:** tanpa SQL, tanpa business logic. Hanya bind/validasi, panggil service, map response.
6. **Business logic hanya di service**, termasuk ownership check ("user hanya boleh mengubah datanya sendiri").
7. Semua method repository dan service menerima `ctx context.Context` sebagai parameter pertama; di handler gunakan `c.Request.Context()`.
8. Jangan menaruh dependency di variabel global; semua di-inject lewat constructor dan dirangkai di `internal/app/routes.go`.

## Konvensi Database & Query Builder

- **Dilarang memakai ORM** (GORM, ent, dll). Gunakan Squirrel + pgx.
- Builder wajib memakai placeholder Postgres:
  ```go
  qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
  ```
- Selalu gunakan `ToSql()` lalu eksekusi dengan pgx; jangan gabungkan string SQL dengan input user.
- Scan dengan helper pgx:
  - satu baris: `pgx.CollectOneRow(rows, pgx.RowToStructByName[T])`
  - banyak baris: `pgx.CollectRows(rows, pgx.RowToStructByName[T])`
  - satu kolom: `pgx.RowTo[string]`
- Ubah `pgx.ErrNoRows` menjadi error domain (`ErrXxxNotFound`) di repository.
- Insert yang butuh ID: `Suffix("RETURNING id, created_at")`.
- Sebutkan kolom secara eksplisit; hindari `SELECT *`.
- **Transaksi:** repository menerima interface `database.DBTX` (dipenuhi `*pgxpool.Pool` dan `pgx.Tx`). Transaksi dibuka di service lewat helper di `platform/database/tx.go`.
- **Migrasi:** SQL manual di `migrations/`, setiap `up` wajib punya `down`. Jangan pernah mengedit migrasi yang sudah di-merge; buat migrasi baru.

Contoh repository:

```go
func (r *repository) FindByID(ctx context.Context, id int64) (*User, error) {
    query, args, err := r.qb.Select("id", "name", "email", "created_at").
        From("users").
        Where(sq.Eq{"id": id}).
        ToSql()
    if err != nil {
        return nil, err
    }
    rows, _ := r.db.Query(ctx, query, args...)
    u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrUserNotFound
    }
    return &u, err
}
```

## Auth & RBAC

- **Autentikasi** (`features/auth`): `auth.Authenticate()` memverifikasi JWT dan menyimpan `user_id` lewat `authctx.SetUserID`. JWT **hanya berisi `user_id`** (dan klaim standar), **bukan** role/permission.
- **Otorisasi** (`features/rbac`): `authz.Require(perm...)` lolos jika user punya **salah satu** permission yang disebut.
- Tabel: `roles`, `permissions`, `role_permissions`, `user_roles`.
- Format kode permission: `resource:action` (huruf kecil), contoh `user:create`, `product:manage`.
- Permission dicache per user (`rbac/cache.go`). **Wajib invalidasi cache** setiap assign/revoke role atau perubahan permission pada role. Jika multi-instance, gunakan Redis.
- Role `super_admin` selalu mendapat semua permission melalui seeder.

### Menambah permission baru

1. Tambahkan konstanta di `features/<feature>/permissions.go` dan masukkan ke slice `Permissions`.
2. Pastikan slice tersebut terdaftar di `internal/app/permissions.go` (diteruskan ke `rbac.AllPermissions`). Daftarnya tidak bisa di `rbac/registry.go` karena feature lain mengimpor `rbac` (import cycle).
3. Pasang di route: `rg.POST("", authz.Require(PermCreate), h.Create)`.
4. Jalankan `make seed`.

## Menambah Feature Baru (checklist)

1. Buat folder `internal/features/<nama>/` dengan file standar di atas.
2. Buat migrasi `make migrate-new name=create_<nama>` (up + down).
3. Definisikan permission di `permissions.go` dan daftarkan di `internal/app/permissions.go`.
4. Implementasikan repository (Squirrel) → service → handler.
5. Buat `RegisterRoutes(rg *gin.RouterGroup, db *pgxpool.Pool, authz *rbac.Middleware)`.
6. Daftarkan di `internal/app/routes.go` di bawah group `protected` (kecuali endpoint publik).
7. Definisikan error domain di `errors.go` feature dengan `apperr.New(apperr.KindXxx, "KODE", "pesan")`; tambahkan mapping di `shared/httpx/errors.go` hanya jika butuh `Kind` baru.
8. Tulis unit test service dengan mock repository.

## Response & Error

- Gunakan helper `httpx` (`httpx.OK`, `httpx.Created`, `httpx.Error`, ...); jangan menulis `c.JSON` manual dengan format berbeda.
- Format response:
  ```json
  { "data": {}, "meta": {} }
  { "error": { "code": "USER_NOT_FOUND", "message": "user not found" } }
  ```
- Mapping status: not found → 404, validasi → 400/422, unauthenticated → 401, forbidden → 403, conflict → 409, lainnya → 500.
- Jangan bocorkan pesan error internal/SQL ke client; log di server.

## Gaya Kode

- `gofmt`/`goimports` wajib; lolos `golangci-lint`.
- Penamaan package huruf kecil satu kata, sama dengan nama folder feature.
- Bungkus error dengan konteks: `fmt.Errorf("find user %d: %w", id, err)`.
- Logging dengan `slog` (structured), sertakan `request_id`.
- Konfigurasi hanya lewat env (`internal/config`); jangan hardcode secret.

## Jangan

- Menambahkan ORM atau auto-migrate.
- Menulis SQL di handler atau service.
- Menyimpan permission/role di JWT.
- Mengimpor implementasi feature lain secara langsung.
- Mengedit migrasi lama yang sudah dijalankan.
- Melewatkan `authz.Require(...)` pada endpoint yang terproteksi.