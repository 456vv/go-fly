package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"imaptool/common"
	"imaptool/models"
	"imaptool/tools"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装导入数据",
	Run: func(cmd *cobra.Command, args []string) {
		ready()
		install()
	},
}

func install() {
	if ok, _ := tools.IsFileNotExist("./install.lock"); !ok {
		log.Println("请先删除./install.lock")
		os.Exit(1)
	}
	sqlFile := "import.sql"
	isExit, _ := tools.IsFileExist(common.MysqlConf)
	dataExit, _ := tools.IsFileExist(sqlFile)
	if !isExit || !dataExit {
		log.Println("config/mysql.json 数据库配置文件或者数据库文件go-fly.sql不存在")
		os.Exit(1)
	}
	sqls, _ := ioutil.ReadFile(sqlFile)
	sqlArr := strings.Split(string(sqls), ";")
	for _, sql := range sqlArr {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		err := models.Execute(sql)
		if err == nil {
			log.Println(sql, "\t success!")
		} else {
			log.Println(sql, err, "\t failed!", "数据库导入失败")
			os.Exit(1)
		}
	}
	installFile, _ := os.OpenFile("./install.lock", os.O_RDWR|os.O_CREATE, os.ModePerm)
	installFile.WriteString("gofly live chat")
}
