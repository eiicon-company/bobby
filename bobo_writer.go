// Package bobo base writer implementation
package bobo

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/gigawattio/metaflector"
	"github.com/spf13/cast"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/mysql"
	"github.com/stephenafamo/bob/dialect/mysql/dm"
	"github.com/volatiletech/strmangle"
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
	reader   BaseReader[T, S, T] // For finding records before update/delete
	infos    any                 // dbinfo table info for accessing column defaults
}

// NewBaseWriter creates a new base writer
func NewBaseWriter[T BaseModel[TSet], S ~[]T, TSet BaseSetter[T], C bob.Expression](
	db *sql.DB,
	table *mysql.Table[T, S, TSet, C], // Bob ORM table
	idColumn mysql.Expression,
	reader BaseReader[T, S, T],
	infos any, // dbinfo table info (e.g., dbinfo.Users)
) BaseWriter[T, S, TSet] {
	return &baseWriter[T, S, TSet, C]{
		db:       db,
		table:    table,
		idColumn: idColumn,
		reader:   reader,
		infos:    infos,
	}
}

// extractID extracts ID field from a model using reflection
func extractID[T any](m T) (int, error) {
	id, err := cast.ToIntE(metaflector.Get(m, "ID"))
	if err != nil {
		return 0, xerrors.Errorf("failed to convert ID to int: %w", err)
	}

	return id, nil
}

// copyWithPreservedR copies all fields from src to dst while preserving dst's R field
func copyWithPreservedR[T any](dst, src T) {
	savedR := metaflector.Get(dst, "R")

	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()
	dstVal.Set(srcVal)

	if savedR != nil {
		if dstRField := dstVal.FieldByName("R"); dstRField.IsValid() && dstRField.CanSet() {
			dstRField.Set(reflect.ValueOf(savedR))
		}
	}
}

// isOptionalSet checks if an optional type (null.Val, omit.Val, omitnull.Val) is set
func isOptionalSet(val reflect.Value) bool {
	isValueMethod := val.MethodByName("IsValue")
	if !isValueMethod.IsValid() {
		return true // Not an optional type, consider it set
	}

	results := isValueMethod.Call([]reflect.Value{})
	if len(results) > 0 && results[0].Kind() == reflect.Bool {
		return results[0].Bool()
	}
	return true
}

// unwrapOptional recursively unwraps optional types to get the base value
func unwrapOptional(val reflect.Value) reflect.Value {
	for {
		getOrZeroMethod := val.MethodByName("GetOrZero")
		if !getOrZeroMethod.IsValid() {
			break
		}

		results := getOrZeroMethod.Call([]reflect.Value{})
		if len(results) == 0 {
			break
		}
		val = results[0]
	}
	return val
}

// callSetterMethod calls the Set method on a setter field with the appropriate value
func callSetterMethod(setterField, modelFieldVal reflect.Value) bool {
	if !setterField.IsValid() || !setterField.CanAddr() {
		return false
	}

	// Check if modelFieldVal is null.Val with IsNull()=true
	// If so, call Null() method on the setter instead of Set()
	if isNull := modelFieldVal.MethodByName("IsNull"); isNull.IsValid() {
		results := isNull.Call([]reflect.Value{})
		if len(results) > 0 && results[0].Kind() == reflect.Bool && results[0].Bool() {
			// This is a NULL value, call Null() method on the setter
			if m := setterField.Addr().MethodByName("Null"); m.IsValid() {
				m.Call([]reflect.Value{})
				return true
			}
		}
	}

	setMethod := setterField.Addr().MethodByName("Set")
	if !setMethod.IsValid() {
		return false
	}

	// Check if we need to unwrap the value for omitnull.Val setters
	valueToPass := modelFieldVal
	if setMethod.Type().NumIn() > 0 {
		expectedType := setMethod.Type().In(0)
		if expectedType != modelFieldVal.Type() {
			valueToPass = unwrapOptional(modelFieldVal)
		}
	}

	setMethod.Call([]reflect.Value{valueToPass})
	return true
}

// shouldSkipFieldForCreate determines if a field should be skipped during Create
func shouldSkipFieldForCreate(field reflect.StructField, modelFieldVal reflect.Value, infos any) bool {
	fieldName := field.Name

	// Skip CreatedAt/UpdatedAt
	//nolint:goconst
	if fieldName == "CreatedAt" || fieldName == "UpdatedAt" {
		return true
	}

	// Skip unexported fields
	if !field.IsExported() {
		return true
	}

	// Check if optional value is not set
	if !isOptionalSet(modelFieldVal) {
		return true
	}

	// Unwrap for condition checking
	unwrappedVal := unwrapOptional(modelFieldVal)

	// Skip time.Time zero values
	if unwrappedVal.Type() == reflect.TypeOf(time.Time{}) && unwrappedVal.Interface().(time.Time).IsZero() {
		return true
	}

	// Check DB defaults using metaflector
	defaultValue, err := cast.ToStringE(metaflector.Get(infos, fmt.Sprintf("Columns.%s.Default", strmangle.TitleCase(fieldName))))
	if err == nil && defaultValue != "" && unwrappedVal.IsZero() {
		return true
	}

	// Skip empty ENUM values
	isNamedString := unwrappedVal.Type().Kind() == reflect.String && (unwrappedVal.Type() != reflect.TypeOf(""))
	if isNamedString && unwrappedVal.String() == "" {
		return true
	}

	return false
}

// shouldSkipFieldForUpdate determines if a field should be skipped during Update
func shouldSkipFieldForUpdate(field reflect.StructField, modelFieldVal reflect.Value) bool {
	fieldName := field.Name

	// Skip CreatedAt
	if fieldName == "CreatedAt" {
		return true
	}

	// Skip unexported fields
	if !field.IsExported() {
		return true
	}

	// UpdatedAt is handled specially, not skipped
	if fieldName == "UpdatedAt" {
		return false
	}

	// Check if optional value is not set
	// Exception: For UPDATE, if IsNull()=true, we want to update to NULL, so don't skip
	if !isOptionalSet(modelFieldVal) {
		// Check if this is a null.Val with IsNull()=true (explicitly NULL)
		if isNull := modelFieldVal.MethodByName("IsNull"); isNull.IsValid() {
			results := isNull.Call([]reflect.Value{})
			if len(results) > 0 && results[0].Kind() == reflect.Bool && results[0].Bool() {
				return false // This is explicitly NULL, include it in UPDATE (don't skip)
			}
		}
		return true
	}

	// Unwrap for condition checking
	unwrappedVal := unwrapOptional(modelFieldVal)

	// Skip time.Time zero values
	if unwrappedVal.Type() == reflect.TypeOf(time.Time{}) && unwrappedVal.Interface().(time.Time).IsZero() {
		return true
	}

	// Skip empty ENUM values
	isNamedString := unwrappedVal.Type().Kind() == reflect.String && (unwrappedVal.Type() != reflect.TypeOf(""))
	if isNamedString && unwrappedVal.String() == "" {
		return true
	}

	return false
}

// buildSetterForCreate creates a Setter for Create operation
// Following apirepo/base.go logic (boil.Infer() + boil.Greylist() for booleans):
// - Include non-zero fields
// - Always include Bool fields (even if false)
// - Exclude CreatedAt/UpdatedAt (use DB defaults)
// - Skip zero-value fields that have DB defaults
func buildSetterForCreate[T BaseModel[TSet], TSet BaseSetter[T]](model T, infos any) TSet {
	modelVal := reflect.ValueOf(model).Elem()
	var zeroSetter TSet
	setterVal := reflect.New(reflect.TypeOf(zeroSetter).Elem()).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		field := modelVal.Type().Field(i)
		if shouldSkipFieldForCreate(field, modelVal.Field(i), infos) {
			continue
		}
		callSetterMethod(setterVal.FieldByName(field.Name), modelVal.Field(i))
	}

	return setterVal.Addr().Interface().(TSet)
}

// buildSetterForUpdate creates a Setter for Update operation
// Following apirepo/base.go logic (boil.Blacklist("updated_at", "created_at")):
// - Include ALL fields except CreatedAt (even if zero)
// - Set UpdatedAt to current UTC time
func buildSetterForUpdate[T BaseModel[TSet], TSet BaseSetter[T]](model T) TSet {
	modelVal := reflect.ValueOf(model).Elem()
	var zeroSetter TSet
	setterVal := reflect.New(reflect.TypeOf(zeroSetter).Elem()).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		field := modelVal.Type().Field(i)
		if field.Name == "UpdatedAt" {
			if setMethod := setterVal.FieldByName("UpdatedAt").Addr().MethodByName("Set"); setMethod.IsValid() {
				setMethod.Call([]reflect.Value{reflect.ValueOf(time.Now().UTC())})
			}
			continue
		}
		if shouldSkipFieldForUpdate(field, modelVal.Field(i)) {
			continue
		}
		callSetterMethod(setterVal.FieldByName(field.Name), modelVal.Field(i))
	}

	return setterVal.Addr().Interface().(TSet)
}

// Create inserts a new record using generics without reflection or type assertions
func (b *baseWriter[T, S, TSet, C]) Create(ctx context.Context, tx bob.Executor, m T) error {
	// Use generic setter builder following apirepo/base.go logic
	setter := buildSetterForCreate[T, TSet](m, b.infos)

	// Use the table's Insert method directly (no reflection, no type assertion)
	insertQuery := b.table.Insert(setter)

	// Call One to execute and get the created record
	created, err := insertQuery.One(ctx, tx)
	if err != nil {
		return xerrors.Errorf("Create: %w", err)
	}

	// Copy created record back to m while preserving R field
	copyWithPreservedR(m, created)

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

	// Use generic setter builder following apirepo/base.go logic
	setter := buildSetterForUpdate[T, TSet](ex)

	// Call Update method directly - no type assertion needed!
	// T has BaseModel[TSet] constraint, so ex.Update is guaranteed to exist
	// Note: ex.Update() will call s.Overwrite(o) which only overwrites fields in the setter,
	// but does NOT include DB-generated fields like UpdatedAt (MySQL ON UPDATE CURRENT_TIMESTAMP)
	if err := ex.Update(ctx, tx, setter); err != nil {
		return xerrors.Errorf("Update execute id=%d: %w", id, err)
	}

	// Reload from DB to get all DB-generated values (UpdatedAt, etc.)
	// This is necessary because Bob ORM's Update() only overwrites setter fields
	if err := ex.Reload(ctx, tx); err != nil {
		return xerrors.Errorf("Update reload after update id=%d: %w", id, err)
	}

	// Finalize: Copy updated ex back to m to reflect all DB-generated values
	// BUT preserve the R field (relations) from m if it exists
	copyWithPreservedR(m, ex)
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
