package saga

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, db.AutoMigrate(&Log{}))

	return db
}

func teardownTestDB(t *testing.T, db *gorm.DB) {
	require.NoError(t, db.Migrator().DropTable(&Log{}))
	sqlDB, _ := db.DB()
	require.NoError(t, sqlDB.Close())
}

func TestAppendLog(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()
	store := NewGormStore(db)

	log := &Log{ExecutionID: gofakeit.UUID(), Type: LogTypeSagaStepExec}
	err := store.AppendLog(ctx, log)

	assert.NoError(t, err)
}

func TestGetAllLogsByExecutionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()
	store := NewGormStore(db)

	exID := gofakeit.UUID()

	log1 := &Log{ExecutionID: exID, Type: LogTypeSagaStepExec}
	log2 := &Log{ExecutionID: exID, Type: LogTypeSagaStepExec}
	require.NoError(t, store.AppendLog(ctx, log1))
	require.NoError(t, store.AppendLog(ctx, log2))

	logs, err := store.GetAllLogsByExecutionID(ctx, exID)

	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}

func TestGetAllLogsByExecutionID_NoLogs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()
	store := NewGormStore(db)

	logs, err := store.GetAllLogsByExecutionID(ctx, "nonexistent")

	assert.NoError(t, err)
	assert.Len(t, logs, 0)
}

func TestGetStepLogsToCompensate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()
	store := NewGormStore(db)

	exID := gofakeit.UUID()

	log1 := &Log{ExecutionID: exID, Type: LogTypeSagaStepExec}
	log2 := &Log{ExecutionID: exID, Type: LogTypeSagaStepExec}
	require.NoError(t, store.AppendLog(ctx, log1))
	require.NoError(t, store.AppendLog(ctx, log2))

	logs, err := store.GetStepLogsToCompensate(ctx, exID)

	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}

func TestGetStepLogsToCompensate_NoLogs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()
	store := NewGormStore(db)

	logs, err := store.GetStepLogsToCompensate(ctx, "nonexistent")

	assert.NoError(t, err)
	assert.Len(t, logs, 0)
}
