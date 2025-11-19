package bobo

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stephenafamo/bob"
	"golang.org/x/xerrors"

	"github.com/eiicon-company/go-core/util/logger"
	"github.com/eiicon-company/go-core/util/testdb"

	"github.com/eiicon-company/auba-api/pkg/environ"
)

var (
	dbMain           testdb.DBTester
	rgxMySQLkey      = regexp.MustCompile(`(?m)\s+CONSTRAINT\s+\S+\s+FOREIGN KEY[^\n]+\n`)
	rgxTrailingComma = regexp.MustCompile(`,\s*\)`)
)

type (
	testEnv struct {
		environ.Env
	}
)

func (e *testEnv) Tenant() (string, error) {
	return "test", nil
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getSchema(filename string) ([]byte, error) {
	schema, err := os.ReadFile(filename)
	if err == nil {
		return schema, nil
	}

	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return nil, xerrors.Errorf("no Go-Modules found: %w", err)
	}

	file := filepath.Join(path.Dir(string(out)), filename)
	schema, err = os.ReadFile(file)
	if err != nil {
		return nil, xerrors.Errorf("no Go-Modules found: %w", err)
	}

	schema = bytes.ReplaceAll(schema, []byte{'\r', '\n'}, []byte{'\n'})
	schema = rgxMySQLkey.ReplaceAll(schema, []byte{})
	schema = rgxTrailingComma.ReplaceAll(schema, []byte("\n)"))
	return schema, nil
}

func TestMain(m *testing.M) {
	dsn := fmt.Sprintf("mysql://%s", getenv("AUBA_API_DSN", "root:@tcp(127.0.0.1:3306)/auba_test?parseTime=true"))

	schema, err := getSchema(getenv("AUBA_API_DDL", "modules/auba-dbmigration/schema.sql"))
	if err != nil {
		logger.Printf("no dbMain tester: %+v", err)
		os.Exit(1)
	}

	db, err := testdb.NewDBTester(dsn, schema)
	if err != nil {
		logger.Printf("no dbMain tester: %+v", err)
		os.Exit(10)
	}

	// Set Global for test database
	dbMain = db

	if err := dbMain.Setup(); err != nil {
		logger.Printf("Unable to execute setup: %+v", err)
		os.Exit(20)
	}

	if _, err := dbMain.Conn(); err != nil {
		logger.Printf("failed to get connection: %+v", err)
		os.Exit(30)
	}

	code := m.Run()

	if err = dbMain.Teardown(); err != nil {
		logger.Printf("Unable to execute teardown: %+v", err)
		os.Exit(40)
	}

	os.Exit(code)
}

// newDebugDB creates a bob.Executor with SQL logging enabled
func newDebugDB(db *sql.DB) bob.Executor {
	// Check if SQL_DEBUG environment variable is set
	if os.Getenv("SQL_DEBUG") != "" {
		// Wrap with Debug to print SQL to stdout
		return bob.Debug(bob.NewDB(db))
	}
	return bob.NewDB(db)
}
