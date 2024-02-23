package cmd

import (
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
		install()
	},
}

func install() {
	if ok, _ := tools.IsFileNotExist("./install.lock"); !ok {
		log.Println("请先删除./install.lock")
		os.Exit(1)
	}

	if err := models.Connect(); err != nil {
		log.Println("数据库连接失败")
		os.Exit(1)
	}

	isExit, _ := tools.IsFileExist(common.MysqlConf)
	if !isExit {
		log.Println("config/go-fly.sql 数据库配置文件或者数据库文件不存在")
		os.Exit(1)
	}
	sqls, _ := os.ReadFile(common.MysqlConf)
	sqlArr := strings.Split(string(sqls), "|")
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
	installFile.Close()
}
