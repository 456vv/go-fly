package models

import (
	"fmt"
	"log"
	"time"

	"imaptool/common"

	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at"`
}

func Connect() error {
	mysql, err := common.GetMysqlConf()
	if err != nil {
		return err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mysql.Username, mysql.Password, mysql.Server, mysql.Port, mysql.Database)

	if DB != nil {
		DB.Close()
	}

	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return err
	}
	DB.SingularTable(true)
	DB.LogMode(true)
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	DB.DB().SetConnMaxLifetime(59 * time.Second)
	InitConfig()
	return nil
}

func Execute(sql string) error {
	if DB == nil {
		if err := Connect(); err != nil {
			return err
		}
	}
	return DB.Exec(sql).Error
}

func CloseDB() {
	DB.Close()
}
