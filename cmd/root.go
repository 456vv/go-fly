package cmd

import (
	"errors"
	"fmt"
	"os"

	"imaptool/models"
	"imaptool/ws"

	"github.com/spf13/cobra"
)

var (
	dir     string
	rootCmd = &cobra.Command{
		Use:   "go-fly",
		Short: "go-fly",
		Long:  `简洁快速的GO语言WEB在线客服 https://gofly.sopans.com`,
		Args:  args,
		Run: func(cmd *cobra.Command, args []string) {
			ready()
		},
	}
)

func args(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("至少需要一个参数")
	}
	return nil
}

func ready() {
	err := os.Chdir(dir)
	exit(err)

	dir, err = os.Getwd()
	exit(err)

	err = models.Connect()
	exit(err)

	go ws.UpdateVisitorStatusCron()
}

func exit(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func Execute() {
	err := rootCmd.Execute()
	exit(err)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "", "./", "程序目录")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(stopCmd)
}
