// Package bobby base type definitions and interfaces
package bobby

import (
	"context"
	"database/sql"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/mysql"
	"github.com/stephenafamo/bob/dialect/mysql/dialect"
	"github.com/stephenafamo/bob/orm"
	"github.com/stephenafamo/scan"
)

// Type aliases for cleaner code
type (
	// SelMod is a type alias for SELECT query modifier
	SelMod = bob.Mod[*dialect.SelectQuery]
	// InsMod is a type alias for INSERT query modifier
	InsMod = bob.Mod[*dialect.InsertQuery]
	// UpdMod is a type alias for UPDATE query modifier
	UpdMod = bob.Mod[*dialect.UpdateQuery]
	// DelMod is a type alias for DELETE query modifier
	DelMod = bob.Mod[*dialect.DeleteQuery]
)

// BaseSetter is a type alias for Bob ORM Setter with generic type constraints
type BaseSetter[T any] = orm.Setter[T, *dialect.InsertQuery, *dialect.UpdateQuery]

// BaseModel represents a Bob ORM generated model with Update, Delete, and Reload methods
// All Bob ORM models implement this interface
type BaseModel[TSet any] interface {
	Update(context.Context, bob.Executor, TSet) error
	Delete(context.Context, bob.Executor) error
	Reload(context.Context, bob.Executor) error
}

// BaseGenerator is used as channel result for streaming large datasets
// This matches the apirepo.BaseGenerator pattern but uses Bob ORM's Cursor for efficient streaming
type BaseGenerator[S any] struct {
	Rows S     // Batch of rows (typically 1000 rows per batch)
	Err  error // Error if any occurred during generation
}

// BaseRepo combines read and write operations for Bob ORM repositories
type BaseRepo[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T]] interface {
	BaseReader[T, S, T]
	BaseWriter[T, S, TSet]
}

// baseRepo is the concrete implementation combining reader and writer
type baseRepo[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T], C bob.Expression] struct {
	*baseReader[T, S]
	*baseWriter[T, S, TSet, C]
}

// NewBaseRepo creates a new base repository with both read and write capabilities
// table should be a Bob ORM generated table (e.g., models.Banners)
// mapper should be the Bob ORM generated mapper function (e.g., models.BannerMapper())
func NewBaseRepo[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T], C bob.Expression](
	db *sql.DB,
	table *mysql.Table[T, S, TSet, C], // Bob ORM generated table (e.g., models.Banners)
	idColumn mysql.Expression,
	mapper scan.Mapper[T], // Bob ORM generated mapper (e.g., models.BannerMapper())
	infos any, // dbinfo table info (e.g., dbinfo.Banners)
) BaseRepo[T, S, TSet] {
	// Create reader first
	tableQuerier, ok := any(table).(TableQuerier[T, S])
	if !ok {
		panic("table does not implement TableQuerier interface")
	}

	reader := &baseReader[T, S]{
		db:       db,
		table:    tableQuerier,
		idColumn: idColumn,
		mapper:   mapper,
	}

	// Create writer with reader reference
	writer := &baseWriter[T, S, TSet, C]{
		db:       db,
		table:    table,
		idColumn: idColumn,
		reader:   reader,
		infos:    infos,
	}

	return &baseRepo[T, S, TSet, C]{
		baseReader: reader,
		baseWriter: writer,
	}
}
