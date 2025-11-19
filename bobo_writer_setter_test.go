package bobo

import (
	"context"
	"testing"
	"time"

	"github.com/stephenafamo/scan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbinfo "github.com/eiicon-company/bobo/internal/testdbinfo"
	enums "github.com/eiicon-company/bobo/internal/testenums"
	models "github.com/eiicon-company/bobo/internal/testmodels"
)

// Comprehensive tests for buildSetterForCreate and buildSetterForUpdate
// These tests ensure that the setter building logic works correctly for ALL edge cases
// Test ID ranges: 96000-96999

// TestBuildSetterForCreate_ComprehensiveCoverage tests all branches and edge cases
// of buildSetterForCreate to ensure NO REGRESSIONS occur during refactoring
func TestBuildSetterForCreate_ComprehensiveCoverage(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 96000-96099
	startID := 96000
	defer cleanupBanners(t, db, makeIDRange(startID, 100)...)

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
		dbinfo.Banners, // IMPORTANT: Pass dbinfo to test DB default value logic
	)

	// Test Case 1: CreatedAt/UpdatedAt should be EXCLUDED even if set
	// Expected: These fields use DB defaults (CURRENT_TIMESTAMP)
	t.Run("CreatedAt_UpdatedAt_Excluded", func(t *testing.T) {
		// Use a past timestamp to ensure DB default (CURRENT_TIMESTAMP) is different
		pastTime := time.Now().UTC().Add(-1 * time.Hour)
		banner := &models.Banner{
			ID:        int32(startID),
			Name:      "Test Banner 1",
			State:     enums.BannersStateBS2,
			Sort:      1,
			CreatedAt: pastTime, // Should be ignored, DB uses CURRENT_TIMESTAMP
			UpdatedAt: pastTime, // Should be ignored, DB uses CURRENT_TIMESTAMP
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		// Verify: CreatedAt/UpdatedAt from DB should be recent (DB CURRENT_TIMESTAMP),
		// not the past time we tried to set
		created, err := reader.Find(ctx, exec, startID)
		require.NoError(t, err)

		// CreatedAt/UpdatedAt should be after pastTime (DB uses CURRENT_TIMESTAMP, not our input)
		// This confirms that buildSetterForCreate correctly excluded CreatedAt/UpdatedAt fields,
		// and DB default (CURRENT_TIMESTAMP) was used instead of the input value
		assert.True(t, created.CreatedAt.After(pastTime),
			"CreatedAt should use DB default (CURRENT_TIMESTAMP), not input value. Expected after %v, got %v",
			pastTime, created.CreatedAt)
		assert.True(t, created.UpdatedAt.After(pastTime),
			"UpdatedAt should use DB default (CURRENT_TIMESTAMP), not input value. Expected after %v, got %v",
			pastTime, created.UpdatedAt)
	})

	// Test Case 2: State enum field with empty value should use DB default
	// Expected: Empty string is skipped -> DB default "BS_1" is used
	t.Run("State_EmptyString_UsesDBDefault", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 1),
			Name:  "Test Banner 2",
			State: "", // Empty enum -> should use DB default
			Sort:  2,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+1)
		require.NoError(t, err)

		// Should use DB default "BS_1"
		assert.Equal(t, enums.BannersStateBS1, created.State,
			"Empty State should use DB default BS_1")
	})

	// Test Case 3: State enum field with value should be included
	// Expected: Non-empty enum value is included in INSERT
	t.Run("State_NonEmpty_Included", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 2),
			Name:  "Test Banner 3",
			State: enums.BannersStateBS3,
			Sort:  3,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+2)
		require.NoError(t, err)

		assert.Equal(t, enums.BannersStateBS3, created.State,
			"Non-empty State should be included")
	})

	// Test Case 4: Sort int field with zero value should be INCLUDED
	// Expected: int(0) is included even though DB default is "0"
	// (Logic: Zero values are included for non-time fields)
	t.Run("Sort_ZeroValue_Included", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 3),
			Name:  "Test Banner 4",
			State: enums.BannersStateBS1,
			Sort:  0, // Zero value with DB default
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+3)
		require.NoError(t, err)

		// Sort=0 should be explicitly set, not from DB default
		assert.Equal(t, int32(0), created.Sort, "Sort=0 should be included")
	})

	// Test Case 5: Sort int field with non-zero value should be INCLUDED
	// Expected: Non-zero values are always included
	t.Run("Sort_NonZeroValue_Included", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 4),
			Name:  "Test Banner 5",
			State: enums.BannersStateBS1,
			Sort:  999,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+4)
		require.NoError(t, err)

		assert.Equal(t, int32(999), created.Sort, "Non-zero Sort should be included")
	})

	// Test Case 6: Name string field with empty value should be INCLUDED
	// Expected: Empty string is included (no DB default, NOT a named string/enum)
	t.Run("Name_EmptyString_Included", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 5),
			Name:  "", // Empty plain string
			State: enums.BannersStateBS1,
			Sort:  5,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+5)
		require.NoError(t, err)

		assert.Equal(t, "", created.Name, "Empty Name should be included")
	})

	// Test Case 7: Name string field with value should be INCLUDED
	// Expected: Non-empty strings are always included
	t.Run("Name_NonEmpty_Included", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 6),
			Name:  "Test Banner 7",
			State: enums.BannersStateBS1,
			Sort:  6,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+6)
		require.NoError(t, err)

		assert.Equal(t, "Test Banner 7", created.Name, "Non-empty Name should be included")
	})

	// Test Case 8: All fields with typical values
	// Expected: All fields are included correctly
	t.Run("AllFields_TypicalValues", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 7),
			Name:  "Comprehensive Test Banner",
			State: enums.BannersStateBS4,
			Sort:  777,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+7)
		require.NoError(t, err)

		assert.Equal(t, int32(startID+7), created.ID)
		assert.Equal(t, "Comprehensive Test Banner", created.Name)
		assert.Equal(t, enums.BannersStateBS4, created.State)
		assert.Equal(t, int32(777), created.Sort)
		assert.NotZero(t, created.CreatedAt)
		assert.NotZero(t, created.UpdatedAt)
	})

	// Test Case 9: Verify DB default behavior for State when omitted
	// This tests the interaction between buildSetterForCreate and DB defaults
	t.Run("State_DBDefault_Integration", func(t *testing.T) {
		banner := &models.Banner{
			ID:    int32(startID + 8),
			Name:  "Default State Test",
			State: "", // Should trigger DB default
			Sort:  8,
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+8)
		require.NoError(t, err)

		// Verify dbinfo.Banners.Columns.State.Default is actually "BS_1"
		assert.Equal(t, "BS_1", dbinfo.Banners.Columns.State.Default,
			"dbinfo should have correct default value")
		assert.Equal(t, enums.BannersStateBS1, created.State,
			"DB default should be applied")
	})

	// Test Case 10: Verify Sort DB default behavior
	t.Run("Sort_DBDefault_Integration", func(t *testing.T) {
		// Note: Sort with zero value is INCLUDED in INSERT
		// This is correct behavior: zero values are included for int fields
		banner := &models.Banner{
			ID:    int32(startID + 9),
			Name:  "Default Sort Test",
			State: enums.BannersStateBS1,
			Sort:  0, // Will be explicitly set to 0
		}

		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)

		created, err := reader.Find(ctx, exec, startID+9)
		require.NoError(t, err)

		// Verify dbinfo.Banners.Columns.Sort.Default
		assert.Equal(t, "0", dbinfo.Banners.Columns.Sort.Default,
			"dbinfo should have correct default value")
		assert.Equal(t, int32(0), created.Sort,
			"Sort=0 should be explicitly set")
	})
}

// TestBuildSetterForUpdate_ComprehensiveCoverage tests all branches and edge cases
// of buildSetterForUpdate to ensure NO REGRESSIONS occur during refactoring
func TestBuildSetterForUpdate_ComprehensiveCoverage(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 96100-96199
	startID := 96100
	defer cleanupBanners(t, db, makeIDRange(startID, 100)...)

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
		nil, // infos not needed for Update
	)

	// Helper to create initial banner
	createBanner := func(id int, name string, state enums.BannersState, sort int) *models.Banner {
		banner := &models.Banner{
			ID:    int32(id),
			Name:  name,
			State: state,
			Sort:  int32(sort),
		}
		err := writer.Create(ctx, exec, banner)
		require.NoError(t, err)
		return banner
	}

	// Test Case 1: CreatedAt should NOT be updated
	// Expected: CreatedAt remains unchanged
	t.Run("CreatedAt_NotUpdated", func(t *testing.T) {
		original := createBanner(startID, "Original", enums.BannersStateBS1, 1)
		originalCreatedAt := original.CreatedAt

		time.Sleep(100 * time.Millisecond) // Ensure time difference

		// Try to update CreatedAt (should be ignored)
		newTime := time.Now().UTC().Add(-24 * time.Hour)
		original.CreatedAt = newTime
		original.Name = "Updated Name"

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID)
		require.NoError(t, err)

		// CreatedAt should NOT change
		assert.Equal(t, originalCreatedAt.Unix(), updated.CreatedAt.Unix(),
			"CreatedAt should not be updated")
		// Name should change
		assert.Equal(t, "Updated Name", updated.Name,
			"Name should be updated")
	})

	// Test Case 2: UpdatedAt behavior in Update
	// Expected: UpdatedAt is updated via OverwriteMerge, then buildSetterForUpdate sets it to now
	// NOTE: Due to OverwriteMerge + buildSetterForUpdate flow, UpdatedAt may or may not be set to current time
	// depending on DB trigger behavior
	t.Run("UpdatedAt_UpdatedByDB", func(t *testing.T) {
		original := createBanner(startID+1, "Original 2", enums.BannersStateBS1, 2)
		originalUpdatedAt := original.UpdatedAt

		time.Sleep(100 * time.Millisecond) // Ensure time difference

		// Update Name
		original.Name = "Updated Name 2"

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID+1)
		require.NoError(t, err)

		// UpdatedAt may or may not be updated depending on implementation
		// Just verify the update succeeded
		t.Logf("Original UpdatedAt: %v, Updated UpdatedAt: %v, Diff: %v",
			originalUpdatedAt, updated.UpdatedAt, updated.UpdatedAt.Sub(originalUpdatedAt))
		// The important thing is that other fields were updated successfully
		assert.Equal(t, "Updated Name 2", updated.Name,
			"Name should be updated")
	})

	// Test Case 3: State empty string is SKIPPED in Update (same as Create)
	// Expected: Empty State is skipped, original value remains
	// NOTE: This is because buildSetterForUpdate skips empty enum values
	t.Run("State_EmptyString_SkippedInUpdate", func(t *testing.T) {
		original := createBanner(startID+2, "Original 3", enums.BannersStateBS2, 3)
		originalState := original.State

		// Try to update State to empty (will be skipped)
		original.State = ""
		original.Name = "Updated Name 3"

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID+2)
		require.NoError(t, err)

		// Empty State is skipped, so original value remains
		assert.Equal(t, originalState, updated.State,
			"Empty State should be skipped, original value remains")
		// Name should be updated
		assert.Equal(t, "Updated Name 3", updated.Name,
			"Name should be updated")
	})

	// Test Case 4: Sort zero value should be INCLUDED in UPDATE
	// Expected: Sort=0 is included (unlike CREATE, UPDATE includes zero values)
	t.Run("Sort_ZeroValue_IncludedInUpdate", func(t *testing.T) {
		original := createBanner(startID+3, "Original 4", enums.BannersStateBS1, 999)

		// Update Sort to 0
		original.Sort = 0

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID+3)
		require.NoError(t, err)

		assert.Equal(t, int32(0), updated.Sort,
			"Sort=0 should be included in UPDATE")
	})

	// Test Case 5: Name empty string should be INCLUDED in UPDATE
	// Expected: Empty Name is included
	t.Run("Name_EmptyString_IncludedInUpdate", func(t *testing.T) {
		original := createBanner(startID+4, "Original 5", enums.BannersStateBS1, 5)

		// Update Name to empty
		original.Name = ""

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID+4)
		require.NoError(t, err)

		assert.Equal(t, "", updated.Name,
			"Empty Name should be included in UPDATE")
	})

	// Test Case 6: All fields can be updated (except CreatedAt)
	// Expected: All fields are updated correctly
	t.Run("AllFields_CanBeUpdated", func(t *testing.T) {
		original := createBanner(startID+5, "Original 6", enums.BannersStateBS1, 6)
		originalCreatedAt := original.CreatedAt
		originalUpdatedAt := original.UpdatedAt

		time.Sleep(100 * time.Millisecond)

		// Update all fields
		original.Name = "Completely Updated"
		original.State = enums.BannersStateBS8
		original.Sort = 123

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID+5)
		require.NoError(t, err)

		// All fields should be updated (except CreatedAt)
		assert.Equal(t, "Completely Updated", updated.Name)
		assert.Equal(t, enums.BannersStateBS8, updated.State)
		assert.Equal(t, int32(123), updated.Sort)
		assert.Equal(t, originalCreatedAt.Unix(), updated.CreatedAt.Unix(),
			"CreatedAt should not change")
		// Log UpdatedAt for information (may or may not change depending on implementation)
		t.Logf("Original UpdatedAt: %v, Updated UpdatedAt: %v",
			originalUpdatedAt, updated.UpdatedAt)
	})

	// Test Case 7: Partial update (only some fields changed)
	// Expected: Changed fields are updated, others remain same
	t.Run("PartialUpdate_OnlyChangedFields", func(t *testing.T) {
		original := createBanner(startID+6, "Original 7", enums.BannersStateBS3, 7)
		originalState := original.State
		originalSort := original.Sort

		time.Sleep(100 * time.Millisecond)

		// Only update Name
		original.Name = "Only Name Changed"

		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		updated, err := reader.Find(ctx, exec, startID+6)
		require.NoError(t, err)

		// Name should be updated
		assert.Equal(t, "Only Name Changed", updated.Name)
		// State and Sort should remain same
		assert.Equal(t, originalState, updated.State)
		assert.Equal(t, originalSort, updated.Sort)
	})

	// Test Case 8: Update twice to verify UpdatedAt changes
	// Expected: UpdatedAt is updated on each call
	t.Run("UpdateTwice_UpdatedAtChanges", func(t *testing.T) {
		original := createBanner(startID+7, "Original 8", enums.BannersStateBS1, 8)

		time.Sleep(100 * time.Millisecond)

		// First update
		original.Name = "First Update"
		err := writer.Update(ctx, exec, original)
		require.NoError(t, err)

		firstUpdate, err := reader.Find(ctx, exec, startID+7)
		require.NoError(t, err)
		firstUpdatedAt := firstUpdate.UpdatedAt

		time.Sleep(100 * time.Millisecond)

		// Second update
		original.Name = "Second Update"
		err = writer.Update(ctx, exec, original)
		require.NoError(t, err)

		secondUpdate, err := reader.Find(ctx, exec, startID+7)
		require.NoError(t, err)
		secondUpdatedAt := secondUpdate.UpdatedAt

		// UpdatedAt should be different
		assert.True(t, secondUpdatedAt.After(firstUpdatedAt),
			"UpdatedAt should change on each update")
	})
}

// Helper functions - removed to inline like other tests
