package whatsmeow_service

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.mau.fi/whatsmeow/store/sqlstore"
	_ "modernc.org/sqlite"

	"github.com/evolution-foundation/evolution-go/pkg/config"
)

func TestGetAuthContainerRetryAndReuse(t *testing.T) {
	service := &whatsmeowService{
		config:        &config.Config{},
		exPath:        filepath.Join(t.TempDir(), "missing"),
		authContainer: &authContainerState{},
	}

	if _, err := service.getAuthContainer(); err == nil {
		t.Fatal("expected an error for a missing SQLite directory")
	}
	if service.authContainer.container != nil {
		t.Fatal("failed container creation must not be memoized")
	}

	service.exPath = sqliteAuthPath(t)
	c1, err := service.getAuthContainer()
	if err != nil {
		t.Fatalf("retry after a transient initialization failure should succeed: %v", err)
	}
	if c1 == nil {
		t.Fatal("expected a container after successful retry")
	}

	c2, err := service.getAuthContainer()
	if err != nil {
		t.Fatalf("second container lookup failed: %v", err)
	}
	if c1 != c2 {
		t.Fatal("successful container must be reused by the same service")
	}
}

func TestAuthContainerStateIsScopedToService(t *testing.T) {
	first := &whatsmeowService{
		config:        &config.Config{},
		exPath:        sqliteAuthPath(t),
		authContainer: &authContainerState{},
	}
	second := &whatsmeowService{
		config:        &config.Config{},
		exPath:        sqliteAuthPath(t),
		authContainer: &authContainerState{},
	}

	firstContainer, err := first.getAuthContainer()
	if err != nil {
		t.Fatal(err)
	}
	secondContainer, err := second.getAuthContainer()
	if err != nil {
		t.Fatal(err)
	}
	if firstContainer == secondContainer {
		t.Fatal("different service instances must not share container state")
	}
}

func TestConcurrentAuthContainerLookupsReuseOneContainer(t *testing.T) {
	service := &whatsmeowService{
		config:        &config.Config{},
		exPath:        sqliteAuthPath(t),
		authContainer: &authContainerState{},
	}

	const callers = 8
	containers := make(chan *sqlstore.Container, callers)
	errors := make(chan error, callers)
	var waitGroup sync.WaitGroup

	for range callers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			container, err := service.getAuthContainer()
			containers <- container
			errors <- err
		}()
	}

	waitGroup.Wait()
	close(containers)
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent lookup failed: %v", err)
		}
	}

	var expected *sqlstore.Container
	for container := range containers {
		if expected == nil {
			expected = container
			continue
		}
		if container != expected {
			t.Fatal("concurrent lookups returned different containers")
		}
	}
}

func TestPostgresAuthContainerUsesExistingPoolAndKeepsItOpenOnUpgradeError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	service := &whatsmeowService{
		config: &config.Config{
			PostgresAuthDB: "postgres://configured",
		},
		authDB:        db,
		authContainer: &authContainerState{},
	}

	resolvedDB, dialect, ownsDB, err := service.authContainerDB()
	if err != nil {
		t.Fatal(err)
	}
	if resolvedDB != db {
		t.Fatal("Postgres path must reuse the service authDB pool")
	}
	if dialect != "postgres" {
		t.Fatalf("expected postgres dialect, got %q", dialect)
	}
	if ownsDB {
		t.Fatal("service authDB must not be marked as owned by the auth container")
	}

	if _, err := service.getAuthContainer(); err == nil {
		t.Fatal("expected sqlmock upgrade to fail without migration expectations")
	}
	if service.authContainer.container != nil {
		t.Fatal("failed Postgres upgrade must not be memoized")
	}

	mock.ExpectPing()
	if err := db.Ping(); err != nil {
		t.Fatalf("shared Postgres pool was closed after upgrade failure: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpgradeFailureClosesOwnedSQLiteDatabase(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	container := sqlstore.NewWithDB(db, "sqlite", nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := upgradeAuthContainer(ctx, container, db, true); err == nil {
		t.Fatal("expected canceled upgrade to fail")
	}
	if err := db.Ping(); err == nil {
		t.Fatal("owned SQLite database must be closed after upgrade failure")
	}
}

func sqliteAuthPath(t *testing.T) string {
	t.Helper()

	path := t.TempDir()
	if err := os.MkdirAll(filepath.Join(path, "dbdata"), 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}
