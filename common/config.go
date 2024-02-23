package common

import (
	"encoding/json"
	"os"
)

type Mysql struct {
	Server   string
	Port     string
	Database string
	Username string
	Password string
}

func GetMysqlConf() (*Mysql, error) {
	info, err := os.ReadFile(MysqlConn)
	if err != nil {
		return nil, err
	}
	mysql := &Mysql{}
	err = json.Unmarshal(info, mysql)
	return mysql, err
}
