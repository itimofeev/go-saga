package saga

import (
	"context"
	"gorm.io/gorm"
)

type gormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
	return &gormStore{db: db}
}

func (g *gormStore) AppendLog(ctx context.Context, log *Log) error {
	return g.db.WithContext(ctx).Create(log).Error
}

func (g *gormStore) GetAllLogsByExecutionID(ctx context.Context, executionID string) ([]*Log, error) {
	logs := make([]*Log, 0)

	err := g.db.WithContext(ctx).Where("execution_id = ?", executionID).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (g *gormStore) GetStepLogsToCompensate(ctx context.Context, executionID string) ([]*Log, error) {
	logs := make([]*Log, 0)

	err := g.db.WithContext(ctx).Where("execution_id = ? AND type = ? AND step_error IS NULL", executionID, LogTypeSagaStepExec).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}
