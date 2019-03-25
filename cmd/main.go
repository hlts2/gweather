package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	f "github.com/hlts2/gweather/internal/fetcher"
	"github.com/hlts2/gweather/internal/redis"
	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var (
	fetcher f.WetherInfomationFetcher
	pool    redis.Pool
)

func action(cli *cli.Context) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		pool.Close()
	}()

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

			mm, err := fetcher.Fetch(ctx, f.URL)
			if err != nil {
				glg.Errorf("faild to fetch contents: %v", err)
			}

			conn, err := pool.GetContext(ctx)
			if err != nil {
				cancel()
				err = errors.Wrap(err, "faild to get redis connection")
				break
			}

			// e.g) key: 気象特別警報・警報・注意報_鳥取地方気象台
			for key, val := range mm {
				b, _ := json.Marshal(val)
				if err := conn.Send("SET", key, b); err != nil {
					glg.Errorf("faild to send: %v", err)
				}
			}

			if err := conn.Flush(); err != nil {
				glg.Errorf("faild to flush: %v", err)
			}

			glg.Infof("Finish job. time: %v", time.Since(start))
		}
	}
}

func before(cli *cli.Context) error {
	if fetcher == nil {
		fetcher = f.New()
	}

	if pool == nil {
		pool = redis.New(cli.String("host"))
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
			Usage: "Interval to get weather information",
			Value: 180,
		},
		cli.StringFlag{
			Name:  "host",
			Usage: "Host address for Redis",
			Value: "redis://127.0.0.1:6379",
		},
	}

	glg.Info("Start cli application")
	if err := app.Run(os.Args); err != nil {
		glg.Error(err)
	}
	glg.Info("Finish cli application")
}
