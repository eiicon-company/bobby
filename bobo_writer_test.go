package bobo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/stephenafamo/bob"
	bobmysql "github.com/stephenafamo/bob/dialect/mysql"
	"github.com/stephenafamo/bob/dialect/mysql/dm"
	"github.com/stephenafamo/scan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eiicon-company/auba-api/pkg/data/model/bobmodel/enums"
	"github.com/eiicon-company/auba-api/pkg/data/model/bobmodel/models"
)

// BaseWriter Method Tests
// Test ID ranges: 95000-95200 (to avoid conflicts with BaseReader tests: 90000-90130)

func TestBaseWriter_Create(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95000-95009
	startID := 95000
	defer cleanupBanners(t, db, makeIDRange(startID, 10)...)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	writer := NewBaseWriter(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		reader,
		nil, // infos not needed for basic test
	)

	// Test Create new record
	banner := createTestBanner(startID)
	err = writer.Create(ctx, exec, banner)
	require.NoError(t, err, "Create should succeed")

	// IMPORTANT: Verify that banner object itself was updated with DB values
	assert.Equal(t, startID, banner.ID, "banner.ID should be set")
	assert.NotZero(t, banner.CreatedAt, "banner.CreatedAt should be set")
	assert.NotZero(t, banner.UpdatedAt, "banner.UpdatedAt should be set")

	// Also verify by reading from DB (data consistency check)
	created, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, startID, created.ID)
	assert.Equal(t, banner.Name, created.Name)
	// Verify banner object matches DB
	assert.Equal(t, banner.CreatedAt.Unix(), created.CreatedAt.Unix(), "banner.CreatedAt should match DB")
	assert.Equal(t, banner.UpdatedAt.Unix(), created.UpdatedAt.Unix(), "banner.UpdatedAt should match DB")

	// Test Create another record
	banner2 := createTestBanner(startID + 1)
	err = writer.Create(ctx, exec, banner2)
	require.NoError(t, err, "Create second record should succeed")

	// Verify second record
	created2, err := reader.Find(ctx, exec, startID+1)
	require.NoError(t, err)
	assert.Equal(t, startID+1, created2.ID)

	// Test Create duplicate ID (should fail due to primary key constraint)
	duplicateBanner := createTestBanner(startID)
	err = writer.Create(ctx, exec, duplicateBanner)
	assert.Error(t, err, "Create with duplicate ID should fail")
}

func TestBaseWriter_Update(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95010-95019
	startID := 95010
	defer cleanupBanners(t, db, makeIDRange(startID, 10)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	writer := NewBaseWriter(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		reader,
		nil, // infos not needed for basic test
	)

	// Create initial record
	banner := createTestBanner(startID)
	err = repo.Create(ctx, exec, banner)
	require.NoError(t, err)

	// initialUpdatedAt := banner.UpdatedAt

	// Update the record
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "Updated Banner Name"
	updatedBanner.Sort = 999
	err = writer.Update(ctx, exec, updatedBanner)
	require.NoError(t, err, "Update should succeed")

	// IMPORTANT: Verify that updatedBanner object itself was updated with latest DB values
	assert.Equal(t, "Updated Banner Name", updatedBanner.Name, "updatedBanner.Name should be updated")
	assert.Equal(t, 999, updatedBanner.Sort, "updatedBanner.Sort should be updated")
	// assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")
	assert.NotZero(t, updatedBanner.CreatedAt, "updatedBanner.CreatedAt should be set")

	// Also verify by reading from DB (data consistency check)
	updated, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Banner Name", updated.Name)
	assert.Equal(t, 999, updated.Sort)
	// Verify updatedBanner object matches DB
	assert.Equal(t, updatedBanner.Name, updated.Name, "updatedBanner should match DB")
	// Use Unix() comparison to avoid timezone/precision issues between Go and MySQL
	assert.Equal(t, updatedBanner.UpdatedAt.Unix(), updated.UpdatedAt.Unix(), "updatedBanner.UpdatedAt should match DB")

	// Test Update non-existent record (should fail)
	nonExistentBanner := createTestBanner(99999)
	err = writer.Update(ctx, exec, nonExistentBanner)
	assert.Error(t, err, "Update non-existent record should fail")
}

func TestBaseWriter_Delete(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95020-95029
	startID := 95020
	defer cleanupBanners(t, db, makeIDRange(startID, 10)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	writer := NewBaseWriter(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		reader,
		nil, // infos not needed for basic test
	)

	// Create record
	banner := createTestBanner(startID)
	err = repo.Create(ctx, exec, banner)
	require.NoError(t, err)

	// Verify record exists
	exists, err := reader.Exists(ctx, exec, startID)
	require.NoError(t, err)
	assert.True(t, exists, "Record should exist before deletion")

	// Delete the record
	err = writer.Delete(ctx, exec, startID)
	require.NoError(t, err, "Delete should succeed")

	// Verify record no longer exists
	exists, err = reader.Exists(ctx, exec, startID)
	require.NoError(t, err)
	assert.False(t, exists, "Record should not exist after deletion")

	// Test Delete non-existent record (should not fail, just no-op)
	err = writer.Delete(ctx, exec, 99999)
	assert.NoError(t, err, "Delete non-existent record should not fail")
}

func TestBaseWriter_Upsert(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95030-95039
	startID := 95030
	defer cleanupBanners(t, db, makeIDRange(startID, 10)...)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	writer := NewBaseWriter(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		reader,
		nil, // infos not needed for basic test
	)

	// Test Upsert as insert (record doesn't exist)
	banner := createTestBanner(startID)
	err = writer.Upsert(ctx, exec, banner)
	require.NoError(t, err, "Upsert as insert should succeed")

	// IMPORTANT: Verify that banner object itself was updated (INSERT case)
	assert.Equal(t, startID, banner.ID, "banner.ID should be set after upsert insert")
	assert.NotZero(t, banner.CreatedAt, "banner.CreatedAt should be set after upsert insert")
	assert.NotZero(t, banner.UpdatedAt, "banner.UpdatedAt should be set after upsert insert")

	// Also verify by reading from DB
	created, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, startID, created.ID)
	assert.Equal(t, banner.Name, created.Name)

	// initialUpdatedAt := banner.UpdatedAt

	// Test Upsert as update (record exists)
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "Upserted Banner Name"
	updatedBanner.Sort = 888
	err = writer.Upsert(ctx, exec, updatedBanner)
	require.NoError(t, err, "Upsert as update should succeed")

	// IMPORTANT: Verify that updatedBanner object itself was updated (UPDATE case)
	assert.Equal(t, "Upserted Banner Name", updatedBanner.Name, "updatedBanner.Name should be updated after upsert update")
	assert.Equal(t, 888, updatedBanner.Sort, "updatedBanner.Sort should be updated after upsert update")
	// assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")

	// Also verify by reading from DB
	updated, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, "Upserted Banner Name", updated.Name)
	assert.Equal(t, 888, updated.Sort)

	// Test Upsert with ID=0 (should insert)
	bannerZeroID := createTestBanner(0)
	bannerZeroID.ID = 0
	err = writer.Upsert(ctx, exec, bannerZeroID)
	require.NoError(t, err, "Upsert with ID=0 should insert")
}

func TestBaseWriter_UpsertLegacy(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95040-95049
	startID := 95040
	defer cleanupBanners(t, db, makeIDRange(startID, 10)...)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	writer := NewBaseWriter(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		reader,
		nil, // infos not needed for basic test
	)

	// Test UpsertLegacy as insert (record doesn't exist, has ID)
	banner := createTestBanner(startID)
	err = writer.UpsertLegacy(ctx, exec, banner)
	require.NoError(t, err, "UpsertLegacy as insert should succeed")

	// IMPORTANT: Verify that banner object itself was updated (INSERT case)
	assert.Equal(t, startID, banner.ID, "banner.ID should be set after upsertlegacy insert")
	assert.NotZero(t, banner.CreatedAt, "banner.CreatedAt should be set after upsertlegacy insert")

	// Also verify by reading from DB
	created, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, startID, created.ID)

	// initialUpdatedAt := banner.UpdatedAt

	// Test UpsertLegacy as update (record exists)
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "UpsertLegacy Updated Name"
	updatedBanner.Sort = 777
	err = writer.UpsertLegacy(ctx, exec, updatedBanner)
	require.NoError(t, err, "UpsertLegacy as update should succeed")

	// IMPORTANT: Verify that updatedBanner object itself was updated (UPDATE case)
	assert.Equal(t, "UpsertLegacy Updated Name", updatedBanner.Name, "updatedBanner.Name should be updated")
	assert.Equal(t, 777, updatedBanner.Sort, "updatedBanner.Sort should be updated")
	// assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")

	// Also verify by reading from DB
	updated, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, "UpsertLegacy Updated Name", updated.Name)
	assert.Equal(t, 777, updated.Sort)

	// Test UpsertLegacy with ID=0 (should insert)
	bannerZeroID := createTestBanner(0)
	bannerZeroID.ID = 0
	err = writer.UpsertLegacy(ctx, exec, bannerZeroID)
	require.NoError(t, err, "UpsertLegacy with ID=0 should insert")
}

func TestBaseWriter_Create_Update_Delete_Sequence(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95050-95059
	startID := 95050
	defer cleanupBanners(t, db, makeIDRange(startID, 10)...)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	writer := NewBaseWriter(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		reader,
		nil, // infos not needed for basic test
	)

	// Step 1: Create record
	banner := createTestBanner(startID)
	err = writer.Create(ctx, exec, banner)
	require.NoError(t, err, "Create should succeed")

	// Step 2: Verify created record - both object and DB
	assert.Equal(t, startID, banner.ID, "banner.ID should be set after create")
	assert.NotZero(t, banner.CreatedAt, "banner.CreatedAt should be set after create")
	created, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, banner.Name, created.Name)

	// initialUpdatedAt := banner.UpdatedAt

	// Step 3: Update record
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "Sequence Updated Name"
	err = writer.Update(ctx, exec, updatedBanner)
	require.NoError(t, err, "Update should succeed")

	// Step 4: Verify updated record - both object and DB
	assert.Equal(t, "Sequence Updated Name", updatedBanner.Name, "updatedBanner.Name should be updated")
	// assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")
	updated, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, "Sequence Updated Name", updated.Name)

	// Step 5: Delete record
	err = writer.Delete(ctx, exec, startID)
	require.NoError(t, err, "Delete should succeed")

	// Step 6: Verify record is deleted
	exists, err := reader.Exists(ctx, exec, startID)
	require.NoError(t, err)
	assert.False(t, exists, "Record should not exist after deletion")
}

// TestBaseWriter_Update_PreservesRelations tests that Update preserves R field (relations)
// This is critical for models with relations like OrganizationPlan with OrganizationPlanBuyer
func TestBaseWriter_Update_PreservesRelations(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95060-95069 for OrganizationPlan
	// Use ID range 95160-95169 for OrganizationPlanBuyer
	planStartID := 95060
	buyerStartID := 95160
	organizationID := 500 // Reuse existing test organization

	// Cleanup any leftover data from previous test runs
	cleanupOrganizationPlans(t, db, makeIDRange(planStartID, 10)...)
	cleanupOrganizationPlanBuyers(t, db, makeIDRange(buyerStartID, 10)...)

	defer cleanupOrganizationPlans(t, db, makeIDRange(planStartID, 10)...)
	defer cleanupOrganizationPlanBuyers(t, db, makeIDRange(buyerStartID, 10)...)

	// Create OrganizationPlanRepo which handles CreateWithBuyer
	planRepo := newOrganizationPlanRepo(newOrganizationPlanRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	// Test Case 1: Update with preloaded relation should preserve R.OrganizationPlanBuyer
	t.Run("Update preserves R.OrganizationPlanBuyer", func(t *testing.T) {
		// Create plan with buyer using CreateWithBuyer
		planWithBuyer := createTestOrganizationPlanWithBuyer(planStartID, buyerStartID, organizationID)
		err = planRepo.CreateWithBuyer(ctx, exec, planWithBuyer)
		require.NoError(t, err, "CreateWithBuyer should succeed")

		// Load plan with buyer relation using FindWithBuyer
		loadedPlan, err := planRepo.FindWithBuyer(ctx, exec, planStartID)
		require.NoError(t, err, "FindWithBuyer should succeed")
		require.NotNil(t, loadedPlan.R, "R should be populated")
		require.NotNil(t, loadedPlan.R.OrganizationPlanBuyer, "R.OrganizationPlanBuyer should be populated")

		// Save buyer info for verification
		buyerEmail := loadedPlan.R.OrganizationPlanBuyer.Email
		buyerCompanyName := loadedPlan.R.OrganizationPlanBuyer.CompanyName
		buyerID := loadedPlan.R.OrganizationPlanBuyer.ID
		require.NotEmpty(t, buyerEmail, "Buyer email should be set")
		require.NotEmpty(t, buyerCompanyName, "Buyer company name should be set")

		// Update the plan (change a field)
		loadedPlan.IsPending = false
		loadedPlan.ContractPrice = 12345
		err = planRepo.Update(ctx, exec, loadedPlan)
		require.NoError(t, err, "Update should succeed")

		// CRITICAL: Verify R.OrganizationPlanBuyer is STILL populated after Update
		assert.NotNil(t, loadedPlan.R, "R should still be populated after Update")
		assert.NotNil(t, loadedPlan.R.OrganizationPlanBuyer, "R.OrganizationPlanBuyer should still be populated after Update")
		assert.Equal(t, buyerID, loadedPlan.R.OrganizationPlanBuyer.ID, "Buyer ID should be preserved")
		assert.Equal(t, buyerEmail, loadedPlan.R.OrganizationPlanBuyer.Email, "Buyer email should be preserved")
		assert.Equal(t, buyerCompanyName, loadedPlan.R.OrganizationPlanBuyer.CompanyName, "Buyer company name should be preserved")

		// Verify the plan fields were actually updated in DB
		updatedPlan, err := planRepo.Find(ctx, exec, planStartID)
		require.NoError(t, err)
		assert.Equal(t, false, updatedPlan.IsPending, "IsPending should be updated in DB")
		assert.Equal(t, 12345, updatedPlan.ContractPrice, "ContractPrice should be updated in DB")
	})

	// Test Case 2: Multiple updates should still preserve relations
	t.Run("Multiple updates preserve R.OrganizationPlanBuyer", func(t *testing.T) {
		planID := planStartID + 1
		buyerID := buyerStartID + 1
		// Use different organization_id to avoid unique constraint violation
		organizationID2 := organizationID + 1

		// Create plan with buyer
		planWithBuyer := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		err = planRepo.CreateWithBuyer(ctx, exec, planWithBuyer)
		require.NoError(t, err)

		// Load plan with buyer
		loadedPlan, err := planRepo.FindWithBuyer(ctx, exec, planID)
		require.NoError(t, err)
		require.NotNil(t, loadedPlan.R.OrganizationPlanBuyer, "Initial: R.OrganizationPlanBuyer should be populated")

		buyerEmail := loadedPlan.R.OrganizationPlanBuyer.Email

		// First update
		loadedPlan.ContractPrice = 10000
		err = planRepo.Update(ctx, exec, loadedPlan)
		require.NoError(t, err)
		assert.NotNil(t, loadedPlan.R.OrganizationPlanBuyer, "After 1st update: R.OrganizationPlanBuyer should be preserved")
		assert.Equal(t, buyerEmail, loadedPlan.R.OrganizationPlanBuyer.Email, "After 1st update: Buyer should be same")

		// Second update
		loadedPlan.ContractPrice = 20000
		err = planRepo.Update(ctx, exec, loadedPlan)
		require.NoError(t, err)
		assert.NotNil(t, loadedPlan.R.OrganizationPlanBuyer, "After 2nd update: R.OrganizationPlanBuyer should be preserved")
		assert.Equal(t, buyerEmail, loadedPlan.R.OrganizationPlanBuyer.Email, "After 2nd update: Buyer should be same")

		// Third update
		loadedPlan.ContractPrice = 30000
		err = planRepo.Update(ctx, exec, loadedPlan)
		require.NoError(t, err)
		assert.NotNil(t, loadedPlan.R.OrganizationPlanBuyer, "After 3rd update: R.OrganizationPlanBuyer should be preserved")
		assert.Equal(t, buyerEmail, loadedPlan.R.OrganizationPlanBuyer.Email, "After 3rd update: Buyer should be same")
	})
}

// Helper functions for OrganizationPlan tests

// cleanupOrganizationPlans removes test organization_plans
func cleanupOrganizationPlans(t *testing.T, db *sql.DB, ids ...int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}

	idExprs := make([]bob.Expression, len(ids))
	for i, id := range ids {
		idExprs[i] = bobmysql.Arg(id)
	}

	ctx := context.Background()
	exec := newDebugDB(db)
	_, err := bobmysql.Delete(
		dm.From(models.OrganizationPlans.Name()),
		dm.Where(models.OrganizationPlans.Columns.ID.In(idExprs...)),
	).Exec(ctx, exec)
	if err != nil {
		t.Logf("Failed to cleanup organization_plans: %v", err)
	}
}

// cleanupOrganizationPlanBuyers removes test organization_plan_buyers
func cleanupOrganizationPlanBuyers(t *testing.T, db *sql.DB, ids ...int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}

	idExprs := make([]bob.Expression, len(ids))
	for i, id := range ids {
		idExprs[i] = bobmysql.Arg(id)
	}

	ctx := context.Background()
	exec := newDebugDB(db)
	_, err := bobmysql.Delete(
		dm.From(models.OrganizationPlanBuyers.Name()),
		dm.Where(models.OrganizationPlanBuyers.Columns.ID.In(idExprs...)),
	).Exec(ctx, exec)
	if err != nil {
		t.Logf("Failed to cleanup organization_plan_buyers: %v", err)
	}
}

// createTestOrganizationPlanWithBuyer creates a test OrganizationPlan with OrganizationPlanBuyer relation
func createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID int) *models.OrganizationPlan {
	buyer := &models.OrganizationPlanBuyer{
		ID:              buyerID,
		CompanyName:     "Test Company",
		PhoneNumber:     "03-1234-5678",
		Email:           "test@example.com",
		Representative:  "Test Representative",
		UserName:        "Test User",
		Postcode:        "100-0001",
		Address1:        "Tokyo",
		Address2:        "Chiyoda",
		InvoicePostcode: "100-0001",
		InvoiceAddress1: "Tokyo Invoice",
		InvoiceAddress2: "Chiyoda Invoice",
		InvoiceUserName: "Invoice User",
		InvoiceEmail:    "invoice@example.com",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	plan := &models.OrganizationPlan{
		ID:                      planID,
		OrganizationID:          organizationID,
		OrganizationPlanBuyerID: buyerID,
		Plan:                    enums.OrganizationPlansPlanBasic,
		PaymentMethod:           enums.OrganizationPlansPaymentMethodApplication,
		IsPending:               true,
		IsEnabled:               null.From(false),
		ContractPrice:           10000,
		ContractVersion:         3,
		ContractPeriod:          30,
		ContractExpiredAt:       time.Now().UTC().AddDate(0, 0, 30),
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	// Set the buyer relation for CreateWithBuyer
	// The R field is a value type (not pointer), so we can access its fields directly
	plan.R.OrganizationPlanBuyer = buyer

	return plan
}
