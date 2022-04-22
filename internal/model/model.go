package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/wesleywxie/gogetit/internal/config"
	"github.com/wesleywxie/gogetit/internal/log"
	"go.uber.org/zap"
	"moul.io/zapgorm"
)

var db *gorm.DB

// InitDB init db object
func InitDB() {
	connectDB()
	configDB()
	updateTable()
}

func connectDB() {
	if config.RunMode == config.TestMode {
		return
	}

	var err error
	db, err = gorm.Open("sqlite3", config.SQLitePath)
	if err != nil {
		zap.S().Fatalf("connect db failed, err: %+v", err)
	}
}

// Disconnect disconnects from the database.
func Disconnect() {
	err := db.Close()
	if err != nil {
		fmt.Printf("fatal error while closing db: %v", err)
	}
}

func configDB() {
	db.DB().SetMaxIdleConns(3)
	db.DB().SetMaxOpenConns(20)
	db.LogMode(config.DBLogMode)
	db.SetLogger(zapgorm.New(log.Logger.WithOptions(zap.AddCallerSkip(7))))
}

func updateTable() {
	createOrUpdateTable(&Video{})
	createOrUpdateTable(&Torrent{})
	createOrUpdateTable(&SelectedTorrent{})
}

// createOrUpdateTable create table or Migrate table
func createOrUpdateTable(model interface{}) {
	if !db.HasTable(model) {
		db.CreateTable(model)
	} else {
		db.AutoMigrate(model)
	}
}
