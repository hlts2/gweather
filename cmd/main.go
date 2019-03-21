package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func getSugeredLogger() (*zap.SugaredLogger, func()) {
	logger, _ := zap.NewProduction()
	suger := logger.Sugar()
	return suger, func() {
		logger.Sync()
	}
}

func action(cli *cli.Context) (err error) {
	sugger, flush := getSugeredLogger()
	sugger.Info("Start cli application")
	defer func() {
		sugger.Info("Finish cli application")
		flush()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	t := time.NewTicker(time.Duration(cli.Uint("s")) * time.Second)
	defer t.Stop()

	for {
		select {
		case sig := <-sigCh:
			sugger.Infow("Received os signal",
				"signal", sig)
			cancel()

		case <-ctx.Done():
			signal.Stop(sigCh)
			close(sigCh)
			return

		case <-t.C:
			start := time.Now()
			sugger.Info("Start job to get information")
			sugger.Infow("Finish job", "time", time.Since(start))
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

	app.Run(os.Args)
}
