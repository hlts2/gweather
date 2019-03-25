package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hlts2/gweather/internal/redis"
	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	f "github.com/hlts2/gweather/internal/fetcher"
)

type action struct {
	fetcher f.WetherInfomationFetcher
	pool    redis.Pool
}

func (a *action) do(cli *cli.Context) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		a.pool.Close()
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

			mm, err := a.fetcher.Fetch(ctx, f.URL)
			if err != nil {
				glg.Errorf("faild to fetch contents: %v", err)
			}

			conn, err := a.pool.GetContext(ctx)
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

type option func(*action)

// WithFetcher returns an option that sets the fetcher.WetherInfomationFetcher implementation.
func WithFetcher(fetcher f.WetherInfomationFetcher) func(*action) {
	return func(a *action) {
		a.fetcher = fetcher
	}
}

// WithPool returns an option that sets the redis.Pool implementation.
func WithPool(pool redis.Pool) func(*action) {
	return func(a *action) {
		a.pool = pool
	}
}

// GetApp returns cli application.
func GetApp(ops ...option) *cli.App {
	app := cli.NewApp()
	app.Name = "gweater"
	app.Usage = "CLI tool for acquiring weather information regularly"
	app.Version = "v1.0,0"

	action := new(action)

	app.Before = func(cli *cli.Context) error {
		action.fetcher = f.New()
		action.pool = redis.New(cli.String("h"))

		for _, op := range ops {
			op(action)
		}

		return nil
	}

	app.Action = action.do
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
	return app
}

func main() {
	app := GetApp()

	glg.Info("Start cli application")
	if err := app.Run(os.Args); err != nil {
		glg.Error(err)
	}
	glg.Info("Finish cli application")
}
