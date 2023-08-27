package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type postgresql struct {
	db *gorm.DB
}

func newPostgresql() (*postgresql, error) {
	postgresql := postgresql{}
	err := postgresql.init()
	if err != nil {
		return nil, err
	}
	return &postgresql, nil
}

func (p *postgresql) Create(deals []deal) error {
	return p.db.Create(deals).Error
}

func (p *postgresql) init() error {
	dsn := "host=database user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	p.db = db
	err = p.migrate()
	if err != nil {
		return err
	}

	return nil
}

type deal struct {
	Type   int64   `gorm:"column:type"`
	Time   int64   `gorm:"column:time"`
	Price  float64 `gorm:"column:price"`
	Volume float64 `gorm:"column:volume"`
}

func (p *postgresql) migrate() error {
	err := p.db.AutoMigrate(&deal{}) //CreateTable(&deal{})
	if err != nil {
		return err
	}
	// err = p.db.Migrator().CreateIndex(&deal{}, "time")
	// if err != nil {
	// 	return err
	// }
	// err = p.db.Migrator().CreateIndex(&deal{}, "idx_time")
	// if err != nil {
	// 	return err
	// }

	return nil
}
