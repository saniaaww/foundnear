# Migrations FoundNear

Setiap anggota menjalankan migration yang sama di PostgreSQL lokal masing-masing
(GitHub hanya menyimpan kode, bukan database).

## Setup

    psql -U postgres -c "CREATE DATABASE foundnear;"
    psql -U postgres -d foundnear -f migrations\001_init_schema.up.sql
    psql -U postgres -d foundnear -f migrations\002_seed_data.up.sql

## Rollback (urutan terbalik)

    psql -U postgres -d foundnear -f migrations\002_seed_data.down.sql
    psql -U postgres -d foundnear -f migrations\001_init_schema.down.sql

## Test repository (database khusus test, isinya DIHAPUS setiap test)

    psql -U postgres -c "CREATE DATABASE foundnear_test;"
    $env:TEST_DATABASE_URL = "postgres://postgres:PASSWORD@localhost:5432/foundnear_test?sslmode=disable"
    go test ./app/repository/... -v

## Kontrak repository (constructor menerima *pgxpool.Pool)

| Repository | Constructor | Method |
|---|---|---|
| RoleRepository | NewRoleRepository | GetRoleByID, GetRoleByName, ListRoles |
| UserRepository | NewUserRepository | CreateUser, GetUserByEmail, GetUserByID, UpdateUser |
| CategoryRepository | NewCategoryRepository | CreateCategory, GetCategoryByID, ListCategories, UpdateCategory, DeleteCategory |
| ReportRepository | NewReportRepository | CreateReport, GetReportByID, GetPublicReportByID, ListPublicReports, ListReportsByUser, UpdateReport, UpdateReportStatus, DeleteReport |
| ClaimRepository | NewClaimRepository | CreateClaim, GetClaimByID, ListClaimsByReportID, ListClaimsByClaimant, ListClaims, UpdateClaimStatus |
| ClaimVerificationRepository | NewClaimVerificationRepository | VerifyClaim (transaksi), GetVerificationByClaimID |
| ItemReturnRepository | NewItemReturnRepository | CreateReturn (transaksi), GetReturnByID, GetReturnByClaimID |

## Error standar (errors.Is dengan package model)

ErrNotFound, ErrDuplicate, ErrForeignKey, ErrConstraint, ErrBusinessRule, ErrInvalidState.

## Aturan penting

- Respons publik wajib memakai model.PublicItemReport (tanpa private_info).
- Approve/reject klaim hanya lewat VerifyClaim, pengembalian hanya lewat CreateReturn,
  karena status report ikut berubah secara atomik.
- Nama role: admin, officer, user (konstanta di app/model/constants.go).