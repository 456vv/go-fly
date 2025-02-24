package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"imaptool/common"
	"imaptool/middleware"
	"imaptool/models"
	"imaptool/router"
	"imaptool/static"
	"imaptool/tools"
	"imaptool/ws"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zh-five/xdaemon"
)

var (
	addr     string
	daemon   bool
	certFile string
	keyFile  string
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "启动http服务",
	Example: "go-fly server -c config/",
	Run: func(cmd *cobra.Command, args []string) {
		os.Chdir(dir)
		run()
	},
}

func init() {
	serverCmd.PersistentFlags().StringVarP(&addr, "addr", "", ":8081", "监听地址")
	serverCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "是否为守护进程模式")
	serverCmd.PersistentFlags().StringVarP(&certFile, "certFile", "", "", "tls证书文件")
	serverCmd.PersistentFlags().StringVarP(&keyFile, "keyFile", "", "", "tls密钥文件")
}

func run() {
	if daemon {
		logFilePath := ""
		if dir, err := os.Getwd(); err == nil {
			logFilePath = dir + "/logs/"
		}
		_, err := os.Stat(logFilePath)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(logFilePath, 0o777); err != nil {
				log.Println(err.Error())
			}
		}
		d := xdaemon.NewDaemon(logFilePath + "go-fly.log")
		d.MaxCount = 10
		d.Run()
	}

	log.Println("start server...\r\ngo：http://" + addr)
	tools.Logger().Println("start server...\r\ngo：http://" + addr)
	fmt.Println(dir)

	engine := gin.Default()
	if common.IsCompireTemplate {
		templ := template.Must(template.New("").ParseFS(static.TemplatesEmbed, "templates/*.html"))
		engine.SetHTMLTemplate(templ)
		engine.StaticFS("/assets", http.FS(static.JsEmbed))
	} else {
		engine.LoadHTMLGlob("static/templates/*")
		engine.Static("/assets", "./static")
	}

	engine.Static("/static", "./static")
	engine.Use(tools.Session("gofly"))
	engine.Use(middleware.CrossSite)
	// 性能监控
	pprof.Register(engine)

	// 连接数据库
	if err := models.Connect(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if noExist, _ := tools.IsFileNotExist("./install.lock"); !noExist {
				log.Println("config/mysql.json 配置文件不存在, 如需要！可能要删除 ./install.lock 文件重新安装。")
				return
			}
			log.Println("config/mysql.json 配置文件不存在, 需要WEB调用 /install 安装才能使用")
		} else {
			// 其它错误
			log.Println(err)
			return
		}
	}
	// 定时给更新数据库状态
	go ws.UpdateVisitorStatusCron()

	// 记录日志
	engine.Use(middleware.NewMidLogger())
	router.InitViewRouter(engine)
	router.InitApiRouter(engine)
	// 记录pid
	os.WriteFile("gofly.sock", []byte(fmt.Sprintf("%d,%d", os.Getppid(), os.Getpid())), 0o666)
	// 限流类
	tools.NewLimitQueue()
	// 清理
	ws.CleanVisitorExpire()
	// 后端websocket
	go ws.WsServerBackend()

	if certFile != "" && keyFile != "" {
		engine.RunTLS(addr, certFile, keyFile)
		return
	}
	engine.Run(addr)
}
