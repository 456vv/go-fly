package common

import (
	"encoding/json"
	"os"

	"imaptool/tools"
)

type Mysql struct {
	Server   string
	Port     string
	Database string
	Username string
	Password string
}

func GetMysqlConf() (*Mysql, error) {
	mysql := &Mysql{}
	isExist, err := tools.IsFileExist(MysqlConf)
	if !isExist {
		return nil, err
	}
	info, err := os.ReadFile(MysqlConf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(info, mysql)
	return mysql, err
}
