package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
	"foundnear/app/repository"
)

var (
	bg       = context.Background()
	testPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		fmt.Println("TEST_DATABASE_URL belum di-set, test database dilewati")
		os.Exit(0)
	}
	if !strings.Contains(url, "test") {
		fmt.Println("TEST_DATABASE_URL harus mengarah ke database yang namanya mengandung 'test'")
		os.Exit(1)
	}

	pool, err := pgxpool.New(bg, url)
	if err != nil {
		fmt.Println("gagal membuat pool:", err)
		os.Exit(1)
	}
	if err := resetSchema(pool); err != nil {
		fmt.Println("gagal menyiapkan schema test:", err)
		os.Exit(1)
	}
	testPool = pool

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func resetSchema(pool *pgxpool.Pool) error {
	if _, err := pool.Exec(bg, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		return err
	}
	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil || len(files) == 0 {
		return fmt.Errorf("file migration tidak ditemukan: %v", err)
	}
	for _, f := range files {
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(bg, string(sqlBytes)); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
	}
	return nil
}

func resetData(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(bg,
		`TRUNCATE item_returns, claim_verifications, item_claims, item_reports, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("gagal membersihkan data: %v", err)
	}
}

func mustUser(t *testing.T, email string) *model.User {
	t.Helper()
	role, err := repository.NewRoleRepository(testPool).GetRoleByName(bg, model.RoleUser)
	if err != nil {
		t.Fatalf("role user tidak ditemukan: %v", err)
	}
	u := &model.User{RoleID: role.ID, Name: "Tester", Email: email, PasswordHash: "$2a$10$hashedpasswordexample"}
	if err := repository.NewUserRepository(testPool).CreateUser(bg, u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return u
}

func mustReport(t *testing.T, userID int64, reportType string) *model.ItemReport {
	t.Helper()
	cats, err := repository.NewCategoryRepository(testPool).ListCategories(bg)
	if err != nil || len(cats) == 0 {
		t.Fatalf("kategori seed tidak ada: %v", err)
	}
	priv := "ada stiker bintang di bagian dalam"
	rp := &model.ItemReport{
		UserID: userID, CategoryID: cats[0].ID, ReportType: reportType,
		ItemName: "Dompet kulit", Description: "Dompet coklat", PrivateInfo: &priv,
		Location: "Perpustakaan", IncidentDate: time.Now(),
	}
	if err := repository.NewReportRepository(testPool).CreateReport(bg, rp); err != nil {
		t.Fatalf("CreateReport: %v", err)
	}
	return rp
}

func mustClaim(t *testing.T, reportID, claimantID int64) *model.ItemClaim {
	t.Helper()
	c := &model.ItemClaim{ReportID: reportID, ClaimantID: claimantID, ClaimDescription: "Itu dompet saya"}
	if err := repository.NewClaimRepository(testPool).CreateClaim(bg, c); err != nil {
		t.Fatalf("CreateClaim: %v", err)
	}
	return c
}

func TestConnection(t *testing.T) {
	if err := testPool.Ping(bg); err != nil {
		t.Fatalf("ping gagal: %v", err)
	}
}

func TestSeedData(t *testing.T) {
	roles, err := repository.NewRoleRepository(testPool).ListRoles(bg)
	if err != nil || len(roles) != 3 {
		t.Fatalf("seharusnya ada 3 role, dapat %d (err=%v)", len(roles), err)
	}
}

func TestUserRepository(t *testing.T) {
	resetData(t)
	repo := repository.NewUserRepository(testPool)

	u := mustUser(t, "  Budi@Example.COM ")
	if u.ID == 0 || u.Email != "budi@example.com" {
		t.Fatalf("user tidak tersimpan dengan benar: %+v", u)
	}

	got, err := repo.GetUserByEmail(bg, "BUDI@example.com")
	if err != nil || got.ID != u.ID || got.RoleName != model.RoleUser || got.PasswordHash == "" {
		t.Fatalf("GetUserByEmail: %+v, err=%v", got, err)
	}

	got, err = repo.GetUserByID(bg, u.ID)
	if err != nil || got.Email != "budi@example.com" {
		t.Fatalf("GetUserByID: %+v, err=%v", got, err)
	}

	phone := "0812345678"
	got.Name, got.Phone = "Budi Baru", &phone
	if err := repo.UpdateUser(bg, got); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	again, _ := repo.GetUserByID(bg, u.ID)
	if again.Name != "Budi Baru" || again.Phone == nil || *again.Phone != phone {
		t.Fatalf("update tidak tersimpan: %+v", again)
	}

	dup := &model.User{RoleID: u.RoleID, Name: "Lain", Email: "budi@example.com", PasswordHash: "x"}
	if err := repo.CreateUser(bg, dup); !errors.Is(err, model.ErrDuplicate) {
		t.Fatalf("email ganda harus ErrDuplicate, dapat %v", err)
	}

	bad := &model.User{RoleID: 9999, Name: "X", Email: "x@example.com", PasswordHash: "x"}
	if err := repo.CreateUser(bg, bad); !errors.Is(err, model.ErrForeignKey) {
		t.Fatalf("role_id tidak valid harus ErrForeignKey, dapat %v", err)
	}

	if _, err := repo.GetUserByEmail(bg, "tidakada@example.com"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("harus ErrNotFound, dapat %v", err)
	}
}

func TestReportRepository_CRUD(t *testing.T) {
	resetData(t)
	repo := repository.NewReportRepository(testPool)
	u := mustUser(t, "pelapor@example.com")

	rp := mustReport(t, u.ID, model.ReportTypeFound)
	if rp.Status != model.ReportStatusActive {
		t.Fatalf("status default harus active, dapat %s", rp.Status)
	}

	full, err := repo.GetReportByID(bg, rp.ID)
	if err != nil || full.PrivateInfo == nil || *full.PrivateInfo == "" {
		t.Fatalf("GetReportByID harus menyertakan private_info: %+v, err=%v", full, err)
	}

	pub, err := repo.GetPublicReportByID(bg, rp.ID)
	if err != nil || pub.CategoryName == "" || pub.ItemName != "Dompet kulit" {
		t.Fatalf("GetPublicReportByID: %+v, err=%v", pub, err)
	}

	full.ItemName = "Dompet hitam"
	if err := repo.UpdateReport(bg, full); err != nil {
		t.Fatalf("UpdateReport: %v", err)
	}
	if again, _ := repo.GetReportByID(bg, rp.ID); again.ItemName != "Dompet hitam" {
		t.Fatalf("update tidak tersimpan: %+v", again)
	}

	if err := repo.UpdateReportStatus(bg, rp.ID, model.ReportStatusClosed); err != nil {
		t.Fatalf("UpdateReportStatus: %v", err)
	}
	if err := repo.UpdateReportStatus(bg, rp.ID, "sembarang"); !errors.Is(err, model.ErrConstraint) {
		t.Fatalf("status tidak valid harus ErrConstraint, dapat %v", err)
	}

	list, total, err := repo.ListPublicReports(bg, model.ReportFilter{Keyword: "hitam", ReportType: model.ReportTypeFound})
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("filter pencarian: total=%d len=%d err=%v", total, len(list), err)
	}
	_, total, _ = repo.ListPublicReports(bg, model.ReportFilter{Keyword: "tidak-ada-barang-ini"})
	if total != 0 {
		t.Fatalf("seharusnya tidak ada hasil, dapat %d", total)
	}

	mine, err := repo.ListReportsByUser(bg, u.ID, 10, 0)
	if err != nil || len(mine) != 1 {
		t.Fatalf("ListReportsByUser: len=%d err=%v", len(mine), err)
	}

	if err := repo.DeleteReport(bg, rp.ID); err != nil {
		t.Fatalf("DeleteReport: %v", err)
	}
	if _, err := repo.GetReportByID(bg, rp.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("setelah delete harus ErrNotFound, dapat %v", err)
	}
	if err := repo.DeleteReport(bg, rp.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("delete ulang harus ErrNotFound, dapat %v", err)
	}
}

func TestReportRepository_Constraints(t *testing.T) {
	resetData(t)
	repo := repository.NewReportRepository(testPool)
	u := mustUser(t, "c@example.com")
	cats, _ := repository.NewCategoryRepository(testPool).ListCategories(bg)

	bad := &model.ItemReport{UserID: u.ID, CategoryID: cats[0].ID, ReportType: "dicuri",
		ItemName: "X", Location: "Y", IncidentDate: time.Now()}
	if err := repo.CreateReport(bg, bad); !errors.Is(err, model.ErrConstraint) {
		t.Fatalf("report_type tidak valid harus ErrConstraint, dapat %v", err)
	}

	noUser := &model.ItemReport{UserID: 9999, CategoryID: cats[0].ID, ReportType: model.ReportTypeLost,
		ItemName: "X", Location: "Y", IncidentDate: time.Now()}
	if err := repo.CreateReport(bg, noUser); !errors.Is(err, model.ErrForeignKey) {
		t.Fatalf("user_id tidak valid harus ErrForeignKey, dapat %v", err)
	}
}

func TestClaimRepository(t *testing.T) {
	resetData(t)
	claims := repository.NewClaimRepository(testPool)
	owner := mustUser(t, "owner@example.com")
	claimant := mustUser(t, "claimant@example.com")

	lost := mustReport(t, owner.ID, model.ReportTypeLost)
	bad := &model.ItemClaim{ReportID: lost.ID, ClaimantID: claimant.ID, ClaimDescription: "x"}
	if err := claims.CreateClaim(bg, bad); !errors.Is(err, model.ErrBusinessRule) {
		t.Fatalf("klaim pada report lost harus ErrBusinessRule, dapat %v", err)
	}

	missing := &model.ItemClaim{ReportID: 9999, ClaimantID: claimant.ID, ClaimDescription: "x"}
	if err := claims.CreateClaim(bg, missing); !errors.Is(err, model.ErrForeignKey) {
		t.Fatalf("report tidak ada harus ErrForeignKey, dapat %v", err)
	}

	found := mustReport(t, owner.ID, model.ReportTypeFound)
	c := mustClaim(t, found.ID, claimant.ID)
	if c.Status != model.ClaimStatusPending {
		t.Fatalf("status default harus pending, dapat %s", c.Status)
	}

	dup := &model.ItemClaim{ReportID: found.ID, ClaimantID: claimant.ID, ClaimDescription: "lagi"}
	if err := claims.CreateClaim(bg, dup); !errors.Is(err, model.ErrDuplicate) {
		t.Fatalf("klaim pending ganda harus ErrDuplicate, dapat %v", err)
	}

	if list, err := claims.ListClaimsByReportID(bg, found.ID); err != nil || len(list) != 1 {
		t.Fatalf("ListClaimsByReportID: len=%d err=%v", len(list), err)
	}
	if list, err := claims.ListClaimsByClaimant(bg, claimant.ID, 10, 0); err != nil || len(list) != 1 {
		t.Fatalf("ListClaimsByClaimant: len=%d err=%v", len(list), err)
	}
	if list, err := claims.ListClaims(bg, model.ClaimStatusPending, 10, 0); err != nil || len(list) != 1 {
		t.Fatalf("ListClaims: len=%d err=%v", len(list), err)
	}

	if err := claims.UpdateClaimStatus(bg, c.ID, "sembarang"); !errors.Is(err, model.ErrConstraint) {
		t.Fatalf("status claim tidak valid harus ErrConstraint, dapat %v", err)
	}
	if err := claims.UpdateClaimStatus(bg, c.ID, model.ClaimStatusCancelled); err != nil {
		t.Fatalf("UpdateClaimStatus: %v", err)
	}

	// Setelah dibatalkan, user yang sama boleh mengajukan klaim baru.
	mustClaim(t, found.ID, claimant.ID)

	// Relasi FK: report yang sudah punya klaim tidak bisa dihapus.
	if err := repository.NewReportRepository(testPool).DeleteReport(bg, found.ID); !errors.Is(err, model.ErrForeignKey) {
		t.Fatalf("delete report berklaim harus ErrForeignKey, dapat %v", err)
	}
}

func TestVerifyClaim_Transaction(t *testing.T) {
	resetData(t)
	reports := repository.NewReportRepository(testPool)
	claims := repository.NewClaimRepository(testPool)
	verifs := repository.NewClaimVerificationRepository(testPool)

	owner := mustUser(t, "owner@example.com")
	claimant := mustUser(t, "claimant@example.com")
	officer := mustUser(t, "officer@example.com")
	found := mustReport(t, owner.ID, model.ReportTypeFound)
	c := mustClaim(t, found.ID, claimant.ID)

	// Result tidak valid: seluruh transaksi harus batal, claim tetap pending.
	invalid := &model.ClaimVerification{ClaimID: c.ID, OfficerID: officer.ID, Result: "mungkin"}
	if err := verifs.VerifyClaim(bg, invalid); !errors.Is(err, model.ErrConstraint) {
		t.Fatalf("result tidak valid harus ErrConstraint, dapat %v", err)
	}
	if cur, _ := claims.GetClaimByID(bg, c.ID); cur.Status != model.ClaimStatusPending {
		t.Fatalf("rollback gagal, status claim: %s", cur.Status)
	}

	ok := &model.ClaimVerification{ClaimID: c.ID, OfficerID: officer.ID, VerificationNote: "ciri cocok", Result: model.VerificationApproved}
	if err := verifs.VerifyClaim(bg, ok); err != nil {
		t.Fatalf("VerifyClaim: %v", err)
	}
	if cur, _ := claims.GetClaimByID(bg, c.ID); cur.Status != model.ClaimStatusApproved {
		t.Fatalf("status claim harus approved, dapat %s", cur.Status)
	}
	if rp, _ := reports.GetReportByID(bg, found.ID); rp.Status != model.ReportStatusMatched {
		t.Fatalf("status report harus matched, dapat %s", rp.Status)
	}
	if v, err := verifs.GetVerificationByClaimID(bg, c.ID); err != nil || v.Result != model.VerificationApproved {
		t.Fatalf("GetVerificationByClaimID: %+v, err=%v", v, err)
	}

	again := &model.ClaimVerification{ClaimID: c.ID, OfficerID: officer.ID, Result: model.VerificationRejected}
	if err := verifs.VerifyClaim(bg, again); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("verifikasi ulang harus ErrInvalidState, dapat %v", err)
	}
}

func TestCreateReturn_Transaction(t *testing.T) {
	resetData(t)
	reports := repository.NewReportRepository(testPool)
	verifs := repository.NewClaimVerificationRepository(testPool)
	returns := repository.NewItemReturnRepository(testPool)

	owner := mustUser(t, "owner@example.com")
	claimant := mustUser(t, "claimant@example.com")
	officer := mustUser(t, "officer@example.com")
	found := mustReport(t, owner.ID, model.ReportTypeFound)
	c := mustClaim(t, found.ID, claimant.ID)

	// Claim belum approved: pengembalian ditolak dan status report tidak berubah.
	early := &model.ItemReturn{ClaimID: c.ID, ReturnedBy: officer.ID, ReceivedBy: claimant.ID}
	if err := returns.CreateReturn(bg, early); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("claim pending harus ErrInvalidState, dapat %v", err)
	}
	if rp, _ := reports.GetReportByID(bg, found.ID); rp.Status == model.ReportStatusReturned {
		t.Fatal("status report tidak boleh returned sebelum claim approved")
	}

	v := &model.ClaimVerification{ClaimID: c.ID, OfficerID: officer.ID, Result: model.VerificationApproved}
	if err := verifs.VerifyClaim(bg, v); err != nil {
		t.Fatalf("VerifyClaim: %v", err)
	}

	ret := &model.ItemReturn{ClaimID: c.ID, ReturnedBy: officer.ID, ReceivedBy: claimant.ID, ReturnNote: "diambil langsung"}
	if err := returns.CreateReturn(bg, ret); err != nil {
		t.Fatalf("CreateReturn: %v", err)
	}
	if ret.ID == 0 || ret.ReturnDate.IsZero() {
		t.Fatalf("ID/ReturnDate harus terisi: %+v", ret)
	}
	if rp, _ := reports.GetReportByID(bg, found.ID); rp.Status != model.ReportStatusReturned {
		t.Fatalf("status report harus returned, dapat %s", rp.Status)
	}
	if got, err := returns.GetReturnByClaimID(bg, c.ID); err != nil || got.ID != ret.ID {
		t.Fatalf("GetReturnByClaimID: %+v, err=%v", got, err)
	}
	if got, err := returns.GetReturnByID(bg, ret.ID); err != nil || got.ReturnNote != "diambil langsung" {
		t.Fatalf("GetReturnByID: %+v, err=%v", got, err)
	}

	// Unique constraint: satu claim maksimal satu pengembalian.
	dup := &model.ItemReturn{ClaimID: c.ID, ReturnedBy: officer.ID, ReceivedBy: claimant.ID}
	if err := returns.CreateReturn(bg, dup); !errors.Is(err, model.ErrDuplicate) {
		t.Fatalf("pengembalian ganda harus ErrDuplicate, dapat %v", err)
	}
}
