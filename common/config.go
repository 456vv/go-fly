package common

import (
	"encoding/json"
	"io/ioutil"
)

type Mysql struct {
	Server   string
	Port     string
	Database string
	Username string
	Password string
}

func GetMysqlConf() (*Mysql, error) {
	var mysql = &Mysql{}
	
	info, err := ioutil.ReadFile(MysqlConf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(info, mysql)
	return mysql, err
}
