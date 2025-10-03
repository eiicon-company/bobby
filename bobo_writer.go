// Package bobo base writer implementation
package bobo

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/mysql"
	"github.com/stephenafamo/bob/dialect/mysql/dm"
	"golang.org/x/xerrors"

	"github.com/eiicon-company/go-core/util/structs"
)

// Use TableWrapper to wrap Bob ORM tables and provide generic operations without reflection

// BaseWriter provides common write operations for Bob ORM repositories
type BaseWriter[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T]] interface {
	// Create inserts a new record
	Create(context.Context, bob.Executor, T) error
	// Update updates an existing record
	Update(context.Context, bob.Executor, T) error
	// Upsert inserts or updates a record
	Upsert(context.Context, bob.Executor, T) error
	// UpsertLegacy uses Create or Update based on ID
	UpsertLegacy(context.Context, bob.Executor, T) error
	// Delete removes a record by ID
	Delete(context.Context, bob.Executor, int) error
}

// baseWriter implements write operations using Bob ORM tables
// T must be a Bob ORM model (implements BaseModel[TSet])
// With generics constraints T BaseModel[TSet] and TSet BaseSetter[T],
// we can directly call T.Update(ctx, exec, TSet) without type assertion!
type baseWriter[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T], C bob.Expression] struct {
	db       *sql.DB
	table    *mysql.Table[T, S, TSet, C] // Bob ORM table with proper Setter constraint
	idColumn mysql.Expression
	toSetter func(T) TSet
	reader   BaseReader[T, S, T] // For finding records before update/delete
}

// NewBaseWriter creates a new base writer
func NewBaseWriter[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T], C bob.Expression](
	db *sql.DB,
	table *mysql.Table[T, S, TSet, C], // Bob ORM table
	idColumn mysql.Expression,
	toSetter func(T) TSet,
	reader BaseReader[T, S, T],
) BaseWriter[T, S, TSet] {
	return &baseWriter[T, S, TSet, C]{
		db:       db,
		table:    table,
		idColumn: idColumn,
		toSetter: toSetter,
		reader:   reader,
	}
}

// extractID extracts ID field from a model using reflection
func extractID[T any](m T) (int, error) {
	rv := reflect.ValueOf(m)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Try to find ID field
	idField := rv.FieldByName("ID")
	if !idField.IsValid() {
		return 0, xerrors.Errorf("model %T does not have ID field", m)
	}

	// Convert to int
	id, err := cast.ToIntE(idField.Interface())
	if err != nil {
		return 0, xerrors.Errorf("failed to convert ID to int: %w", err)
	}

	return id, nil
}

// Create inserts a new record using generics without reflection or type assertions
func (b *baseWriter[T, S, TSet, C]) Create(ctx context.Context, tx bob.Executor, m T) error {
	setter := b.toSetter(m)

	// Use the table's Insert method directly (no reflection, no type assertion)
	insertQuery := b.table.Insert(setter)

	// Call One to execute and get the created record
	created, err := insertQuery.One(ctx, tx)
	if err != nil {
		return xerrors.Errorf("Create: %w", err)
	}

	// Copy created back to m
	if err := copier.Copy(&m, created); err != nil {
		return xerrors.Errorf("Create copy result: %w", err)
	}

	return nil
}

// Update updates an existing record using generics without reflection or type assertions
func (b *baseWriter[T, S, TSet, C]) Update(ctx context.Context, tx bob.Executor, m T) error {
	id, err := extractID(m)
	if err != nil {
		return xerrors.Errorf("Update extract ID: %w", err)
	}

	// Find existing record with preload
	ex, err := b.reader.Find(ctx, tx, id)
	if err != nil {
		return xerrors.Errorf("Update find existing id=%d: %w", id, err)
	}

	// Merge new values into existing using structs.OverwriteMerge
	// T is already a pointer type (*models.Banner), so don't take address again
	if err := structs.OverwriteMerge(ex, m); err != nil {
		return xerrors.Errorf("Update merge failed id=%d: %w", id, err)
	}

	// Convert to setter
	setter := b.toSetter(ex)

	// Call Update method directly - no type assertion needed!
	// T has BaseModel[TSet] constraint, so ex.Update is guaranteed to exist
	// ex.Update() will update the database AND update the ex object itself via s.Overwrite(o)
	if err := ex.Update(ctx, tx, setter); err != nil {
		return xerrors.Errorf("Update execute id=%d: %w", id, err)
	}

	// Copy the freshly updated record back to m
	if err := copier.Copy(m, ex); err != nil {
		return xerrors.Errorf("Update copy result: %w", err)
	}

	return nil
}

// Upsert inserts or updates a record
func (b *baseWriter[T, S, TSet, C]) Upsert(ctx context.Context, tx bob.Executor, m T) error {
	return b.UpsertLegacy(ctx, tx, m)
}

// UpsertLegacy uses Create or Update based on ID
func (b *baseWriter[T, S, TSet, C]) UpsertLegacy(ctx context.Context, tx bob.Executor, m T) error {
	id, _ := extractID(m) // Ignore error, treat as 0 if no ID

	if id == 0 {
		return b.Create(ctx, tx, m)
	}

	// Check if record exists
	exists, err := b.reader.Exists(ctx, tx, id)
	if err != nil {
		return xerrors.Errorf("UpsertLegacy check exists id=%d: %w", id, err)
	}

	if exists {
		return b.Update(ctx, tx, m)
	}

	return b.Create(ctx, tx, m)
}

// Delete removes a record by ID using generics without reflection or type assertions
func (b *baseWriter[T, S, TSet, C]) Delete(ctx context.Context, tx bob.Executor, id int) error {
	// Use the table's Delete method directly (no reflection, no type assertion)
	_, err := b.table.Delete(dm.Where(b.idColumn.EQ(mysql.Arg(id)))).Exec(ctx, tx)
	if err != nil {
		return xerrors.Errorf("Delete execute id=%d: %w", id, err)
	}

	return nil
}
