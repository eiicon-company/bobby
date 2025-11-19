package bobo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/mysql"
	"github.com/stephenafamo/bob/dialect/mysql/dm"
	sm "github.com/stephenafamo/bob/dialect/mysql/sm"
	"github.com/stephenafamo/scan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/eiicon-company/auba-api/pkg/data/model/bobmodel/enums"
	"github.com/eiicon-company/auba-api/pkg/data/model/bobmodel/models"
)

// Test helper to clean up test data using Bob ORM
func cleanupBanners(t *testing.T, db *sql.DB, ids ...int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}

	// Convert int slice to bob.Expression slice for IN clause
	idExprs := make([]bob.Expression, len(ids))
	for i, id := range ids {
		idExprs[i] = mysql.Arg(id)
	}

	// Use Bob ORM's Delete statement builder with dm package
	ctx := context.Background()
	exec := newDebugDB(db)
	_, err := mysql.Delete(
		dm.From(models.Banners.Name()),
		dm.Where(models.Banners.Columns.ID.In(idExprs...)),
	).Exec(ctx, exec)
	if err != nil {
		t.Logf("Failed to cleanup banners: %v", err)
	}
}

// Test helper to create a test banner
func createTestBanner(id int) *models.Banner {
	return &models.Banner{
		ID:        id,
		Name:      "Test Banner " + string(rune(id)),
		State:     enums.BannersStateMypageTop,
		Sort:      id,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// setupGeneratorTestData creates test banners in the 80000 range
func setupGeneratorTestData(t *testing.T, _ *sql.DB, exec bob.Executor, repo BannerRepo, startID, count int) {
	t.Helper()
	ctx := context.Background()

	for i := 0; i < count; i++ {
		banner := createTestBanner(startID + i)
		err := repo.Create(ctx, exec, banner)
		require.NoError(t, err, "failed to create test banner %d", startID+i)
	}
}

func TestBaseReader_Generator_SmallDataset(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 80000-80009 (10 records)
	startID := 80000
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	// Create repo for setup
	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	// Setup test data
	setupGeneratorTestData(t, db, exec, repo, startID, count)

	// Create base reader
	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test Generator
	ch := reader.Generator(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
		sm.OrderBy(models.Banners.Columns.ID.String()).Asc(),
	})

	// Collect all results
	var allBanners models.BannerSlice
	var batchCount int
	for gen := range ch {
		require.NoError(t, gen.Err, "Generator should not return error")
		allBanners = append(allBanners, gen.Rows...)
		batchCount++
	}

	// Verify results
	assert.Equal(t, 1, batchCount, "Small dataset should be in single batch")
	assert.Len(t, allBanners, count, "Should receive all banners")

	// Verify order
	for i := 0; i < count; i++ {
		assert.Equal(t, startID+i, allBanners[i].ID, "Banners should be in correct order")
	}
}

func TestBaseReader_Generator_LargeDataset(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 81000-83499 (2500 records)
	startID := 81000
	count := 2500
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	// Create repo for setup
	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	// Setup test data
	setupGeneratorTestData(t, db, exec, repo, startID, count)

	// Create base reader
	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test Generator
	ch := reader.Generator(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
		sm.OrderBy(models.Banners.Columns.ID.String()).Asc(),
	})

	// Collect all results
	var allBanners models.BannerSlice
	var batchSizes []int
	for gen := range ch {
		require.NoError(t, gen.Err, "Generator should not return error")
		batchSizes = append(batchSizes, len(gen.Rows))
		allBanners = append(allBanners, gen.Rows...)
	}

	// Verify batching: 1000 + 1000 + 500 = 2500
	assert.Equal(t, 3, len(batchSizes), "Should have 3 batches")
	assert.Equal(t, 1000, batchSizes[0], "First batch should have 1000 rows")
	assert.Equal(t, 1000, batchSizes[1], "Second batch should have 1000 rows")
	assert.Equal(t, 500, batchSizes[2], "Third batch should have 500 rows")

	// Verify total count
	assert.Len(t, allBanners, count, "Should receive all banners")

	// Verify order
	for i := 0; i < count; i++ {
		assert.Equal(t, startID+i, allBanners[i].ID, "Banners should be in correct order")
	}
}

func TestBaseReader_Generator_ContextCancellation(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)

	// Use ID range 84000-88999 (5000 records for 5 batches)
	startID := 84000
	count := 5000
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	// Create repo for setup
	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	// Setup test data
	setupGeneratorTestData(t, db, exec, repo, startID, count)

	// Create base reader
	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test Generator
	ch := reader.Generator(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
		sm.OrderBy(models.Banners.Columns.ID.String()).Asc(),
	})

	// Cancel context after receiving first batch
	var receivedBatches int
	var gotCancelError bool
	var totalRecords int
	for gen := range ch {
		if gen.Err != nil {
			// Context cancellation error expected
			if gen.Err == context.Canceled || xerrors.Is(gen.Err, context.Canceled) {
				gotCancelError = true
			}
			break
		}
		receivedBatches++
		totalRecords += len(gen.Rows)
		if receivedBatches == 1 {
			// Cancel after first batch
			cancel()
		}
	}

	// Verify cancellation worked
	assert.True(t, receivedBatches >= 1, "Should receive at least one batch before cancellation")
	// Either we got cancel error, or we didn't receive all data (due to cancellation)
	if !gotCancelError {
		assert.Less(t, totalRecords, count, "Should not receive all data if cancellation happened during processing")
		t.Logf("Cancellation happened during processing: received %d/%d records in %d batches (no error message)",
			totalRecords, count, receivedBatches)
	} else {
		t.Logf("Received explicit cancellation error after %d batches", receivedBatches)
	}
}

func TestBaseReader_Generator_EmptyDataset(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 85000-85999 (no data)
	startID := 85000
	count := 1000

	// Create base reader
	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test Generator with no data
	ch := reader.Generator(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
		sm.OrderBy(models.Banners.Columns.ID.String()).Asc(),
	})

	// Collect all results
	var allBanners models.BannerSlice
	var hasError bool
	for gen := range ch {
		if gen.Err != nil {
			hasError = true
			t.Logf("Unexpected error: %v", gen.Err)
		}
		allBanners = append(allBanners, gen.Rows...)
	}

	// Verify results
	assert.False(t, hasError, "Empty dataset should not produce error")
	assert.Len(t, allBanners, 0, "Should receive no banners")
}

func TestBaseReader_Generator_WithConditions(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 86000-86099 (100 records with different states)
	startID := 86000
	count := 100
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	// Create repo for setup
	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	// Setup test data with alternating states
	for i := 0; i < count; i++ {
		banner := createTestBanner(startID + i)
		// Alternate between MYPAGE_TOP and NOLOGIN_TOP
		if i%2 == 0 {
			banner.State = enums.BannersStateMypageTop
		} else {
			banner.State = enums.BannersStateNologinTop
		}
		err := repo.Create(ctx, exec, banner)
		require.NoError(t, err)
	}

	// Create base reader
	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test Generator with state filter (MYPAGE_TOP only)
	ch := reader.Generator(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
		sm.Where(models.Banners.Columns.State.EQ(mysql.Arg(enums.BannersStateMypageTop))),
		sm.OrderBy(models.Banners.Columns.ID.String()).Asc(),
	})

	// Collect all results
	var allBanners models.BannerSlice
	for gen := range ch {
		require.NoError(t, gen.Err, "Generator should not return error")
		allBanners = append(allBanners, gen.Rows...)
	}

	// Verify results: should get 50 records (every other one)
	assert.Len(t, allBanners, 50, "Should receive only MYPAGE_TOP banners")

	// Verify all have correct state
	for _, banner := range allBanners {
		assert.Equal(t, enums.BannersStateMypageTop, banner.State, "All banners should have MYPAGE_TOP state")
	}
}

// makeIDRange creates a slice of IDs for cleanup
func makeIDRange(start, count int) []int {
	ids := make([]int, count)
	for i := 0; i < count; i++ {
		ids[i] = start + i
	}
	return ids
}

func TestBaseReader_Conn(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	assert.Equal(t, db, reader.Conn(), "Conn() should return the same db instance")
}

func TestBaseReader_Find(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90000-90009
	startID := 90000
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	// Setup test data
	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test Find existing record
	banner, err := reader.Find(ctx, exec, startID)
	require.NoError(t, err)
	assert.Equal(t, startID, banner.ID)
	assert.Equal(t, "Test Banner "+string(rune(startID)), banner.Name)

	// Test Find non-existent record
	_, err = reader.Find(ctx, exec, 99999)
	assert.Error(t, err, "Should return error for non-existent record")
}

func TestBaseReader_FindBy(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90010-90019
	startID := 90010
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test FindBy with ID condition (should return first matching record, no implicit ordering)
	banner, err := reader.FindBy(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.EQ(mysql.Arg(startID))),
	})
	require.NoError(t, err)
	assert.Equal(t, startID, banner.ID, "FindBy should return the matching record")

	// Test FindBy with no results
	_, err = reader.FindBy(ctx, exec, []SelMod{
		sm.Where(models.Banners.Columns.ID.EQ(mysql.Arg(99999))),
	})
	assert.Error(t, err, "Should return error when no records found")
}

func TestBaseReader_FindPreload(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID 90020
	id := 90020
	defer cleanupBanners(t, db, id)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, id, 1)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test FindPreload
	banner, err := reader.FindPreload(ctx, exec, id)
	require.NoError(t, err)
	assert.Equal(t, id, banner.ID)

	// Test FindPreload non-existent
	_, err = reader.FindPreload(ctx, exec, 99999)
	assert.Error(t, err)
}

func TestBaseReader_FirstBy_LastBy(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90030-90039
	startID := 90030
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	where := []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
	}

	// Test FirstBy - should return lowest ID
	first, err := reader.FirstBy(ctx, exec, where)
	require.NoError(t, err)
	assert.Equal(t, startID, first.ID, "FirstBy should return record with lowest ID")

	// Test LastBy - should return highest ID
	last, err := reader.LastBy(ctx, exec, where)
	require.NoError(t, err)
	assert.Equal(t, startID+count-1, last.ID, "LastBy should return record with highest ID")
}

func TestBaseReader_All(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90040-90049
	startID := 90040
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test All - should return all banners in DESC order
	banners, err := reader.All(ctx, exec)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(banners), count, "All should return at least our test records")

	// Find our test records in results (DESC order, so highest ID first)
	foundCount := 0
	for _, banner := range banners {
		if banner.ID >= startID && banner.ID < startID+count {
			foundCount++
		}
	}
	assert.Equal(t, count, foundCount, "All should include all our test records")
}

func TestBaseReader_AllPreload(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90050-90059
	startID := 90050
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test AllPreload with WHERE condition
	banners, err := reader.AllPreload(ctx, exec,
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID+count))),
	)
	require.NoError(t, err)
	assert.Len(t, banners, count, "AllPreload should return all matching records")

	// Verify DESC order
	for i := 0; i < count; i++ {
		expectedID := startID + count - 1 - i
		assert.Equal(t, expectedID, banners[i].ID, "AllPreload should return records in DESC order")
	}
}

func TestBaseReader_ListBy_SliceBy(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90060-90079
	startID := 90060
	count := 20
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	where := []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
	}

	// Test ListBy - should return records in DESC order
	listBanners, err := reader.ListBy(ctx, exec, where)
	require.NoError(t, err)
	assert.Len(t, listBanners, count)
	// Verify DESC order
	for i := 0; i < count; i++ {
		expectedID := startID + count - 1 - i
		assert.Equal(t, expectedID, listBanners[i].ID, "ListBy should return in DESC order")
	}

	// Test SliceBy - returns records without default ordering (but we can add our own)
	sliceBanners, err := reader.SliceBy(ctx, exec, where, sm.OrderBy(models.Banners.Columns.ID.String()).Asc())
	require.NoError(t, err)
	assert.Len(t, sliceBanners, count)
	// Verify ASC order (because we specified it)
	for i := 0; i < count; i++ {
		expectedID := startID + i
		assert.Equal(t, expectedID, sliceBanners[i].ID, "SliceBy with ASC order should return in ASC order")
	}
}

func TestBaseReader_ListByIDs(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90080-90089
	startID := 90080
	count := 10
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test ListByIDs with subset of IDs
	ids := []int{startID, startID + 2, startID + 5, startID + 9}
	banners, err := reader.ListByIDs(ctx, exec, ids)
	require.NoError(t, err)
	assert.Len(t, banners, len(ids))

	// Verify all requested IDs are present
	foundIDs := make(map[int]bool)
	for _, banner := range banners {
		foundIDs[banner.ID] = true
	}
	for _, id := range ids {
		assert.True(t, foundIDs[id], "ListByIDs should include ID %d", id)
	}

	// Test ListByIDs with empty slice
	emptyBanners, err := reader.ListByIDs(ctx, exec, []int{})
	require.NoError(t, err)
	assert.Len(t, emptyBanners, 0, "ListByIDs with empty IDs should return empty slice")
}

func TestBaseReader_ListPagerBy_SlicePagerBy(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 90090-90129 (40 records)
	startID := 90090
	count := 40
	defer cleanupBanners(t, db, makeIDRange(startID, count)...)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, startID, count)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	where := []SelMod{
		sm.Where(models.Banners.Columns.ID.GTE(mysql.Arg(startID))),
		sm.Where(models.Banners.Columns.ID.LT(mysql.Arg(startID + count))),
	}

	// Test ListPagerBy - First page
	banners1, total1, err := reader.ListPagerBy(ctx, exec, where, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, count, total1, "Total count should be %d", count)
	assert.Len(t, banners1, 10, "First page should have 10 records")

	// Test ListPagerBy - Second page
	banners2, total2, err := reader.ListPagerBy(ctx, exec, where, 10, 10)
	require.NoError(t, err)
	assert.Equal(t, count, total2, "Total count should be %d", count)
	assert.Len(t, banners2, 10, "Second page should have 10 records")

	// Test ListPagerBy - Last page (partial)
	banners4, total4, err := reader.ListPagerBy(ctx, exec, where, 10, 30)
	require.NoError(t, err)
	assert.Equal(t, count, total4, "Total count should be %d", count)
	assert.Len(t, banners4, 10, "Last page should have 10 records")

	// Test SlicePagerBy
	sliceBanners, sliceTotal, err := reader.SlicePagerBy(ctx, exec, where, 15, 0)
	require.NoError(t, err)
	assert.Equal(t, count, sliceTotal, "Total count should be %d", count)
	assert.Len(t, sliceBanners, 15, "First page should have 15 records")
}

func TestBaseReader_Exists(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID 90130
	id := 90130
	defer cleanupBanners(t, db, id)

	repo := newBannerRepo(newBannerRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	setupGeneratorTestData(t, db, exec, repo, id, 1)

	reader := NewBaseReader(
		db,
		models.Banners,
		models.Banners.Columns.ID,
		scan.StructMapper[*models.Banner](),
	)

	// Test Exists - existing record
	exists, err := reader.Exists(ctx, exec, id)
	require.NoError(t, err)
	assert.True(t, exists, "Exists should return true for existing record")

	// Test Exists - non-existent record
	notExists, err := reader.Exists(ctx, exec, 99999)
	require.NoError(t, err)
	assert.False(t, notExists, "Exists should return false for non-existent record")
}
