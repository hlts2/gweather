package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kpango/glg"
	"github.com/urfave/cli"
)

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
			glg.Infof("Received of signal: %v", sig)
			cancel()

		case <-ctx.Done():
			signal.Stop(sigCh)
			close(sigCh)
			return

		case <-t.C:
			start := time.Now()
			glg.Info("Start job to get information")
			glg.Infof("Finish job. time: %v", time.Since(start))
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "gweater"
	app.Usage = "CLI tool for acquiring weather information regularly"
	app.Version = "v1.0,0"
	app.Action = action
	app.Flags = []cli.Flag{
		cli.UintFlag{
			Name:  "second, s",
			Usage: "Interval to get information",
			Value: 180,
		},
	}

	glg.Info("Start CLI Application")
	app.Run(os.Args)
	glg.Info("Finish CLI Application")
}
