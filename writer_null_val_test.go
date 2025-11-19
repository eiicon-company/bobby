package bobby

import (
	"context"
	"testing"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBaseWriter_Create_NullVal tests that Create correctly handles null.Val fields
// Test ID range: 95200-95249
func TestBaseWriter_Create_NullVal(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95200-95209 for OrganizationPlan
	// Use ID range 95250-95259 for OrganizationPlanBuyer
	planStartID := 95200
	buyerStartID := 95250
	organizationID := 500 // Reuse existing test organization

	cleanupOrganizationPlans(t, db, makeIDRange(planStartID, 10)...)
	cleanupOrganizationPlanBuyers(t, db, makeIDRange(buyerStartID, 10)...)

	defer cleanupOrganizationPlans(t, db, makeIDRange(planStartID, 10)...)
	defer cleanupOrganizationPlanBuyers(t, db, makeIDRange(buyerStartID, 10)...)

	planRepo := newOrganizationPlanRepo(newOrganizationPlanRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	t.Run("Create with null.Val[bool]{} should set NULL in DB", func(t *testing.T) {
		planID := planStartID
		buyerID := buyerStartID

		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID)
		// Set IsEnabled to null.Val[bool]{} - this should result in NULL in DB
		plan.IsEnabled = null.Val[bool]{}

		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err, "CreateWithBuyer should succeed")

		// Read from DB and verify IsEnabled is NULL
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		// CRITICAL TEST: null.Val[bool]{} should result in NULL in DB
		assert.False(t, created.IsEnabled.IsValue(), "IsEnabled should be NULL (IsValue=false)")
		assert.True(t, created.IsEnabled.IsNull(), "IsEnabled should be NULL (IsNull=true)")
	})

	t.Run("Create with null.From(true) should set true in DB", func(t *testing.T) {
		planID := planStartID + 1
		buyerID := buyerStartID + 1
		organizationID2 := organizationID + 1

		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		// Set IsEnabled to null.From(true) - this should result in true in DB
		plan.IsEnabled = null.From(true)

		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err, "CreateWithBuyer should succeed")

		// Read from DB and verify IsEnabled is true
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		assert.True(t, created.IsEnabled.IsValue(), "IsEnabled should have value (IsValue=true)")
		assert.Equal(t, true, created.IsEnabled.GetOrZero(), "IsEnabled should be true")
	})

	t.Run("Create with null.From(false) should set false in DB", func(t *testing.T) {
		planID := planStartID + 2
		buyerID := buyerStartID + 2
		organizationID2 := organizationID + 2

		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		// Set IsEnabled to null.From(false) - this should result in false in DB
		plan.IsEnabled = null.From(false)

		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err, "CreateWithBuyer should succeed")

		// Read from DB and verify IsEnabled is false
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		assert.True(t, created.IsEnabled.IsValue(), "IsEnabled should have value (IsValue=true)")
		assert.Equal(t, false, created.IsEnabled.GetOrZero(), "IsEnabled should be false")
	})

	t.Run("Create with null.Val[time.Time]{} should set NULL in DB", func(t *testing.T) {
		planID := planStartID + 3
		buyerID := buyerStartID + 3
		organizationID2 := organizationID + 3

		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		// Set UnsubscribedRequestedAt to null.Val[time.Time]{} - this should result in NULL in DB
		plan.UnsubscribedRequestedAt = null.Val[time.Time]{}

		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err, "CreateWithBuyer should succeed")

		// Read from DB and verify UnsubscribedRequestedAt is NULL
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		// CRITICAL TEST: null.Val[time.Time]{} should result in NULL in DB
		assert.False(t, created.UnsubscribedRequestedAt.IsValue(), "UnsubscribedRequestedAt should be NULL (IsValue=false)")
		assert.True(t, created.UnsubscribedRequestedAt.IsNull(), "UnsubscribedRequestedAt should be NULL (IsNull=true)")
	})

	t.Run("Create with null.From(time) should set time in DB", func(t *testing.T) {
		planID := planStartID + 4
		buyerID := buyerStartID + 4
		organizationID2 := organizationID + 4

		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		// Set UnsubscribedRequestedAt to null.From(time) - this should result in time in DB
		testTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		plan.UnsubscribedRequestedAt = null.From(testTime)

		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err, "CreateWithBuyer should succeed")

		// Read from DB and verify UnsubscribedRequestedAt has the value
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		assert.True(t, created.UnsubscribedRequestedAt.IsValue(), "UnsubscribedRequestedAt should have value (IsValue=true)")
		assert.Equal(t, testTime.Unix(), created.UnsubscribedRequestedAt.GetOrZero().Unix(), "UnsubscribedRequestedAt should match")
	})
}

// TestBaseWriter_Update_NullVal tests that Update correctly handles null.Val fields
// Test ID range: 95210-95229
func TestBaseWriter_Update_NullVal(t *testing.T) {
	db, err := dbMain.Conn()
	require.NoError(t, err)

	exec := newDebugDB(db)
	ctx := context.Background()

	// Use ID range 95210-95219 for OrganizationPlan
	// Use ID range 95260-95269 for OrganizationPlanBuyer
	planStartID := 95210
	buyerStartID := 95260
	organizationID := 510 // Different from Create tests

	cleanupOrganizationPlans(t, db, makeIDRange(planStartID, 10)...)
	cleanupOrganizationPlanBuyers(t, db, makeIDRange(buyerStartID, 10)...)

	defer cleanupOrganizationPlans(t, db, makeIDRange(planStartID, 10)...)
	defer cleanupOrganizationPlanBuyers(t, db, makeIDRange(buyerStartID, 10)...)

	planRepo := newOrganizationPlanRepo(newOrganizationPlanRepoIn{
		Env: &testEnv{},
		DB:  db,
	})

	t.Run("Update with null.Val[bool]{} should set NULL in DB", func(t *testing.T) {
		planID := planStartID
		buyerID := buyerStartID

		// Create plan with IsEnabled = true
		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID)
		plan.IsEnabled = null.From(true)
		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err)

		// Verify it's true
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)
		require.True(t, created.IsEnabled.IsValue(), "Initial: IsEnabled should have value")
		require.Equal(t, true, created.IsEnabled.GetOrZero(), "Initial: IsEnabled should be true")

		// Update IsEnabled to null.Val[bool]{} - this should set NULL in DB
		updatePlan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID)
		updatePlan.IsEnabled = null.Val[bool]{}
		err = planRepo.Update(ctx, exec, updatePlan)
		require.NoError(t, err, "Update should succeed")

		// Read from DB and verify IsEnabled is now NULL
		updated, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		// CRITICAL TEST: null.Val[bool]{} should result in NULL in DB
		assert.False(t, updated.IsEnabled.IsValue(), "After update: IsEnabled should be NULL (IsValue=false)")
		assert.True(t, updated.IsEnabled.IsNull(), "After update: IsEnabled should be NULL (IsNull=true)")
	})

	t.Run("Update with null.From(false) should set false in DB", func(t *testing.T) {
		planID := planStartID + 1
		buyerID := buyerStartID + 1
		organizationID2 := organizationID + 1

		// Create plan with IsEnabled = true
		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		plan.IsEnabled = null.From(true)
		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err)

		// Update IsEnabled to null.From(false) - this should set false in DB
		updatePlan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		updatePlan.IsEnabled = null.From(false)
		err = planRepo.Update(ctx, exec, updatePlan)
		require.NoError(t, err, "Update should succeed")

		// Read from DB and verify IsEnabled is now false
		updated, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		assert.True(t, updated.IsEnabled.IsValue(), "After update: IsEnabled should have value (IsValue=true)")
		assert.Equal(t, false, updated.IsEnabled.GetOrZero(), "After update: IsEnabled should be false")
	})

	t.Run("Update with null.Val[time.Time]{} should set NULL in DB", func(t *testing.T) {
		planID := planStartID + 2
		buyerID := buyerStartID + 2
		organizationID2 := organizationID + 2

		// Create plan with UnsubscribedRequestedAt = time
		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		testTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		plan.UnsubscribedRequestedAt = null.From(testTime)
		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err)

		// Verify it has value
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)
		require.True(t, created.UnsubscribedRequestedAt.IsValue(), "Initial: UnsubscribedRequestedAt should have value")

		// Update UnsubscribedRequestedAt to null.Val[time.Time]{} - this should set NULL in DB
		updatePlan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		updatePlan.UnsubscribedRequestedAt = null.Val[time.Time]{}
		err = planRepo.Update(ctx, exec, updatePlan)
		require.NoError(t, err, "Update should succeed")

		// Read from DB and verify UnsubscribedRequestedAt is now NULL
		updated, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		// CRITICAL TEST: null.Val[time.Time]{} should result in NULL in DB
		assert.False(t, updated.UnsubscribedRequestedAt.IsValue(), "After update: UnsubscribedRequestedAt should be NULL (IsValue=false)")
		assert.True(t, updated.UnsubscribedRequestedAt.IsNull(), "After update: UnsubscribedRequestedAt should be NULL (IsNull=true)")
	})

	t.Run("Update from NULL to value should work", func(t *testing.T) {
		planID := planStartID + 3
		buyerID := buyerStartID + 3
		organizationID2 := organizationID + 3

		// Create plan with IsEnabled = NULL
		plan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		plan.IsEnabled = null.Val[bool]{}
		err = planRepo.CreateWithBuyer(ctx, exec, plan)
		require.NoError(t, err)

		// Verify it's NULL
		created, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)
		require.False(t, created.IsEnabled.IsValue(), "Initial: IsEnabled should be NULL")

		// Update IsEnabled to null.From(true) - this should set true in DB
		updatePlan := createTestOrganizationPlanWithBuyer(planID, buyerID, organizationID2)
		updatePlan.IsEnabled = null.From(true)
		err = planRepo.Update(ctx, exec, updatePlan)
		require.NoError(t, err, "Update should succeed")

		// Read from DB and verify IsEnabled is now true
		updated, err := planRepo.Find(ctx, exec, planID)
		require.NoError(t, err)

		assert.True(t, updated.IsEnabled.IsValue(), "After update: IsEnabled should have value (IsValue=true)")
		assert.Equal(t, true, updated.IsEnabled.GetOrZero(), "After update: IsEnabled should be true")
	})
}
