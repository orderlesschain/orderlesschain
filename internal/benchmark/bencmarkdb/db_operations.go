package bencmarkdb

// The GORM query language: https://gorm.io/docs/create.html

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

const dbLocation = "./data/benchmark.db"

type ExperimentStatus int

const (
	Running ExperimentStatus = iota
	Failed
	Done
)

type ExperimentsORM struct {
	gorm.Model
	ExperimentID string `gorm:"index"`
	Status       ExperimentStatus
	ResultPath   string
}

type ExperimentOperations struct {
	db *gorm.DB
}

func NewExperimentOperations() *ExperimentOperations {
	var err error
	// On performance of the DB and ORM: https://gorm.io/docs/performance.html
	DB, err := gorm.Open(sqlite.Open(dbLocation), &gorm.Config{
		SkipDefaultTransaction:   true,
		PrepareStmt:              true,
		DisableNestedTransaction: true,
		Logger:                   gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxIdleTime(time.Hour)
	sqlDB.SetConnMaxLifetime(24 * time.Hour)

	err = DB.AutoMigrate(&ExperimentsORM{})
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return &ExperimentOperations{
		db: DB,
	}
}

func (exop *ExperimentOperations) AddBenchmarkToDB(benchmark *ExperimentsORM) (uint, error) {
	if err := exop.db.Create(benchmark).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return 0, err
	}
	return benchmark.ID, nil
}

func (exop *ExperimentOperations) GetOneExperiment(id uint) (*ExperimentsORM, error) {
	experiment := &ExperimentsORM{}
	if err := exop.db.First(&experiment, id).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return nil, err
	}
	return experiment, nil
}

func (exop *ExperimentOperations) GetExperimentStatus(id uint) (ExperimentStatus, error) {
	experiment := &ExperimentsORM{}
	if err := exop.db.First(&experiment, id).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return Failed, err
	}
	return experiment.Status, nil
}

func (exop *ExperimentOperations) UpdateExperimentStatus(id uint, status ExperimentStatus) error {
	experiment := &ExperimentsORM{}
	if err := exop.db.First(&experiment, id).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	experiment.Status = status
	if err := exop.db.Save(&experiment).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	return nil
}

func (exop *ExperimentOperations) UpdateExperimentReportPath(id uint, path string) error {
	experiment := &ExperimentsORM{}
	if err := exop.db.First(&experiment, id).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	experiment.ResultPath = path
	if err := exop.db.Save(&experiment).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	return nil
}

func (exop *ExperimentOperations) GetExperimentStatusWithExperimentID(experimentID string) (ExperimentStatus, error) {
	experiment := &ExperimentsORM{}
	if err := exop.db.Where("experiment_id = ?", experimentID).First(&experiment).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return Failed, err
	}
	return experiment.Status, nil
}

func (exop *ExperimentOperations) GetExperimentResultPathWithExperimentID(experimentID string) (string, error) {
	experiment := &ExperimentsORM{}
	if err := exop.db.Where("experiment_id = ?", experimentID).First(&experiment).Error; err != nil {
		logger.ErrorLogger.Println(err)
		return "", err
	}
	return experiment.ResultPath, nil
}
