package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	f "github.com/hlts2/gweather/internal/fetcher"
	"github.com/kpango/glg"
	"github.com/urfave/cli"
)

var fetcher f.WetherInfomationFetcher

func action(cli *cli.Context) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	t := time.NewTicker(time.Duration(cli.Uint("s")) * time.Second)
	defer t.Stop()

	for {
		select {
		case sig := <-sigCh:
			glg.Warnf("Received os signal: %v", sig)
			cancel()

		case <-ctx.Done():
			signal.Stop(sigCh)
			close(sigCh)
			return

		case <-t.C:
			start := time.Now()
			glg.Info("Start job to get information")

			mm, err := fetcher.Fetch(ctx)
			if err != nil {
				glg.Errorf("faild to fetch contents: %v", err)
			}

			_ = mm

			// // FIXME: 毎回Connectionを生成しない
			// c := redis.NewConn(nil, 1*time.Microsecond, 1*time.Second)
			//
			// // e.g) key: 気象特別警報・警報・注意報_鳥取地方気象台
			// for key, val := range mm {
			// 	c.Send("SET", key, val)
			// }
			//
			// err = c.Flush()
			// if err != nil {
			// 	glg.Errorf("faild to flush: %v", err)
			// }

			glg.Infof("Finish job. time: %v", time.Since(start))
		}
	}
}

func before(cli *cli.Context) error {
	if fetcher == nil {
		fetcher = f.New()
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "gweater"
	app.Usage = "CLI tool for acquiring weather information regularly"
	app.Version = "v1.0,0"
	app.Before = before
	app.Action = action
	app.Flags = []cli.Flag{
		cli.UintFlag{
			Name:  "second, s",
			Usage: "Interval to get information",
			Value: 180,
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Set the configuration file",
			Value: "coofig.yaml",
		},
	}

	glg.Info("Start cli application")
	err := app.Run(os.Args)
	if err != nil {
		glg.Error(err)
	}
	glg.Info("Finish cli application")
}
