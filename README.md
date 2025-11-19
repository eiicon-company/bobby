# Bobo

Bobo is a lightweight, generic repository pattern library for [Bob ORM](https://github.com/stephenafamo/bob), providing common CRUD operations with type-safe generics.

## Installation

```bash
go get github.com/eiicon-company/bobo
```

## Quick Start

```go
import (
    "context"
    "database/sql"

	"go.uber.org/dig"
    "github.com/eiicon-company/bobo"
    "github.com/stephenafamo/scan"

    // Your Bob ORM generated models
    "yourproject/models"
)

type (
	UserRepo interface {
		bobo.BoboRepo[*models.User, models.UserSlice, *models.UserSetter]
		// Custom method for this repository
		ListByState(context.Context, bob.Executor, string, ...bobo.SelMod) (models.UserSlice, error)
	}

	userRepo struct {
		bobo.BoboRepo[*models.User, models.UserSlice, *models.UserSetter]
	}
)

// ListByState retrieves by state
func (b *userRepo) ListByState(ctx context.Context, exec bob.Executor, state string, loads ...bobo.SelMod) (models.UserSlice, error) {
	if state == "" {
		return nil, xerrors.Errorf("no state arg found: %w", sql.ErrNoRows)
	}

	mods := []bobo.SelMod{
		sm.Where(models.Users.Columns.State.EQ(mysql.Arg(state))),
	}

	return b.ListBy(ctx, exec, mods, loads...)
}

type newUserRepoIn struct {
	dig.In
	DB  *sql.DB
}

func newUserRepo(in newUserRepoIn) UserRepo {
	return &userRepo{
		in: in,
		BaseRepo: NewBaseRepo[*models.User, models.UserSlice, *models.UserSetter](
			in.DB,
			models.Users,
			models.Users.Columns.ID,
			scan.StructMapper[*models.User](), // Mapper for cursor operations
			dbinfo.Users,                      // Table info for column defaults
		),
	}
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details

