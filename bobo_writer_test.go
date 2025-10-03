package bobo

import (
	"context"
	"testing"

	"github.com/stephenafamo/scan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		toBannerSetter,
		reader,
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
		toBannerSetter,
		reader,
	)

	// Create initial record
	banner := createTestBanner(startID)
	err = repo.Create(ctx, exec, banner)
	require.NoError(t, err)

	initialUpdatedAt := banner.UpdatedAt

	// Update the record
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "Updated Banner Name"
	updatedBanner.Sort = 999
	err = writer.Update(ctx, exec, updatedBanner)
	require.NoError(t, err, "Update should succeed")

	// IMPORTANT: Verify that updatedBanner object itself was updated with latest DB values
	assert.Equal(t, "Updated Banner Name", updatedBanner.Name, "updatedBanner.Name should be updated")
	assert.Equal(t, 999, updatedBanner.Sort, "updatedBanner.Sort should be updated")
	assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")
	assert.NotZero(t, updatedBanner.CreatedAt, "updatedBanner.CreatedAt should be set")

	// Also verify by reading from DB (data consistency check)
	updated, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Banner Name", updated.Name)
	assert.Equal(t, 999, updated.Sort)
	// Verify updatedBanner object matches DB
	assert.Equal(t, updatedBanner.Name, updated.Name, "updatedBanner should match DB")
	assert.True(t, updatedBanner.UpdatedAt.Equal(updated.UpdatedAt), "updatedBanner.UpdatedAt should match DB")

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
		toBannerSetter,
		reader,
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
		toBannerSetter,
		reader,
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

	initialUpdatedAt := banner.UpdatedAt

	// Test Upsert as update (record exists)
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "Upserted Banner Name"
	updatedBanner.Sort = 888
	err = writer.Upsert(ctx, exec, updatedBanner)
	require.NoError(t, err, "Upsert as update should succeed")

	// IMPORTANT: Verify that updatedBanner object itself was updated (UPDATE case)
	assert.Equal(t, "Upserted Banner Name", updatedBanner.Name, "updatedBanner.Name should be updated after upsert update")
	assert.Equal(t, 888, updatedBanner.Sort, "updatedBanner.Sort should be updated after upsert update")
	assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")

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
		toBannerSetter,
		reader,
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

	initialUpdatedAt := banner.UpdatedAt

	// Test UpsertLegacy as update (record exists)
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "UpsertLegacy Updated Name"
	updatedBanner.Sort = 777
	err = writer.UpsertLegacy(ctx, exec, updatedBanner)
	require.NoError(t, err, "UpsertLegacy as update should succeed")

	// IMPORTANT: Verify that updatedBanner object itself was updated (UPDATE case)
	assert.Equal(t, "UpsertLegacy Updated Name", updatedBanner.Name, "updatedBanner.Name should be updated")
	assert.Equal(t, 777, updatedBanner.Sort, "updatedBanner.Sort should be updated")
	assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")

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
		toBannerSetter,
		reader,
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

	initialUpdatedAt := banner.UpdatedAt

	// Step 3: Update record
	updatedBanner := createTestBanner(startID)
	updatedBanner.Name = "Sequence Updated Name"
	err = writer.Update(ctx, exec, updatedBanner)
	require.NoError(t, err, "Update should succeed")

	// Step 4: Verify updated record - both object and DB
	assert.Equal(t, "Sequence Updated Name", updatedBanner.Name, "updatedBanner.Name should be updated")
	assert.True(t, updatedBanner.UpdatedAt.After(initialUpdatedAt), "updatedBanner.UpdatedAt should be newer (MySQL ON UPDATE CURRENT_TIMESTAMP)")
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
