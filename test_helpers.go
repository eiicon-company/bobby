package bobby

import (
	"context"
	"database/sql"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"go.uber.org/dig"

	dbinfo "github.com/eiicon-company/bobby/internal/testdbinfo"
	"github.com/eiicon-company/bobby/internal/testenv"
	models "github.com/eiicon-company/bobby/internal/testmodels"
)

// Test repository interfaces and implementations

type (
	// BannerRepo is a test repository for Banner model
	BannerRepo interface {
		BaseRepo[*models.Banner, models.BannerSlice, *models.BannerSetter]
	}

	bannerRepo struct {
		BaseRepo[*models.Banner, models.BannerSlice, *models.BannerSetter]
	}

	newBannerRepoIn struct {
		dig.In
		Env testenv.Env
		DB  *sql.DB
	}

	// OrganizationPlanRepo is a test repository for OrganizationPlan model
	OrganizationPlanRepo interface {
		BaseRepo[*models.OrganizationPlan, models.OrganizationPlanSlice, *models.OrganizationPlanSetter]
		CreateWithBuyer(context.Context, bob.Executor, *models.OrganizationPlan) error
		FindWithBuyer(context.Context, bob.Executor, int) (*models.OrganizationPlan, error)
	}

	organizationPlanRepo struct {
		BaseRepo[*models.OrganizationPlan, models.OrganizationPlanSlice, *models.OrganizationPlanSetter]
	}

	newOrganizationPlanRepoIn struct {
		dig.In
		Env testenv.Env
		DB  *sql.DB
	}
)

// newBannerRepo creates a new Banner repository for testing
func newBannerRepo(in newBannerRepoIn) BannerRepo {
	return &bannerRepo{
		BaseRepo: NewBaseRepo[*models.Banner, models.BannerSlice, *models.BannerSetter](
			in.DB,
			models.Banners,
			models.Banners.Columns.ID,
			scan.StructMapper[*models.Banner](),
			dbinfo.Banners,
		),
	}
}

// newOrganizationPlanRepo creates a new OrganizationPlan repository for testing
func newOrganizationPlanRepo(in newOrganizationPlanRepoIn) OrganizationPlanRepo {
	return &organizationPlanRepo{
		BaseRepo: NewBaseRepo[*models.OrganizationPlan, models.OrganizationPlanSlice, *models.OrganizationPlanSetter](
			in.DB,
			models.OrganizationPlans,
			models.OrganizationPlans.Columns.ID,
			scan.StructMapper[*models.OrganizationPlan](),
			dbinfo.OrganizationPlans,
		),
	}
}

// CreateWithBuyer creates an OrganizationPlan along with its OrganizationPlanBuyer
func (r *organizationPlanRepo) CreateWithBuyer(ctx context.Context, exec bob.Executor, plan *models.OrganizationPlan) error {
	// First create the buyer if it exists in R field
	if plan.R.OrganizationPlanBuyer != nil {
		buyer := plan.R.OrganizationPlanBuyer

		// Use Bob ORM's Insert with buildSetterForCreate logic
		buyerSetter := buildSetterForCreate[*models.OrganizationPlanBuyer, *models.OrganizationPlanBuyerSetter](
			buyer,
			dbinfo.OrganizationPlanBuyers,
		)

		_, err := models.OrganizationPlanBuyers.Insert(buyerSetter).Exec(ctx, exec)
		if err != nil {
			return err
		}
	}

	// Then create the plan
	return r.Create(ctx, exec, plan)
}

// FindWithBuyer finds an OrganizationPlan and loads its OrganizationPlanBuyer relation
func (r *organizationPlanRepo) FindWithBuyer(ctx context.Context, exec bob.Executor, id int) (*models.OrganizationPlan, error) {
	plan, err := r.Find(ctx, exec, id)
	if err != nil {
		return nil, err
	}

	// Load the buyer relation
	buyer, err := models.OrganizationPlanBuyers.Query(
		models.SelectWhere.OrganizationPlanBuyers.ID.EQ(plan.OrganizationPlanBuyerID),
	).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	plan.R.OrganizationPlanBuyer = buyer
	return plan, nil
}
