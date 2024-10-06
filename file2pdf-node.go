package main

import (
	"errors"
	"fmt"
	"github.com/abbhb/file2pdf-node/depot"
	"github.com/abbhb/file2pdf-node/mqcode"

	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var Version = "unstable"

func init() {
	log.SetPrefix("file2pdf-node ")
}

func main() {
	app := cli.NewApp()
	app.Name = "file2pdf-node"
	app.Version = Version
	app.Usage = "The simple REST API for unoserver and unoconvert"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Value: "0.0.0.0:2004",
			Usage: "The addr used by the unoserver api server",
		},
		cli.StringFlag{
			Name:   "unoserver-addr",
			Value:  "127.0.0.1:2002",
			Usage:  "The unoserver addr used by the unoconvert",
			EnvVar: "UNOSERVER_ADDR",
		},
		cli.StringFlag{
			Name:   "unoconvert-bin",
			Value:  "unoconvert",
			Usage:  "Set the unoconvert executable path.",
			EnvVar: "UNOCONVERT_BIN",
		},
		cli.DurationFlag{
			Name:   "unoconvert-timeout",
			Value:  10 * time.Minute,
			Usage:  "Set the unoconvert run timeout",
			EnvVar: "UNOCONVERT_TIMEOUT",
		},
	}
	app.Authors = []cli.Author{
		{
			Name:  "libreoffice-docker",
			Email: "https://github.com/libreofficedocker/unoserver-rest-api",
		},
	}
	app.Action = mainAction
	var err = errors.New("nil")
	for {
		err = app.Run(os.Args)
		if err = app.Run(os.Args); err != nil {
			fmt.Printf("重试中，出错了:%s", err)
			time.Sleep(time.Second * 5)
		}
		if err == nil {
			break
		}
	}

}

func mainAction(c *cli.Context) {
	// Create temporary working directory
	depot.MkdirTemp()

	// Cleanup temporary working directory after finished
	defer depot.CleanTemp()

	// Configure unoconvert options
	//unoAddr := c.String("unoserver-addr")
	//host, port, _ := net.SplitHostPort(unoAddr)
	//unoconvert.SetInterface(host)
	//unoconvert.SetPort(port)
	//unoconvert.SetExecutable(c.String("unoconvert-bin"))
	//unoconvert.SetContextTimeout(c.Duration("unoconvert-timeout"))
	var producer = mqcode.CreateProducer()
	// 启动rocketmq消费者开始等待需要转换的文件进入，一次最大允许一个
	var consumers = mqcode.CreateConsumer(producer)
	consumers.StartConsumer()
}
