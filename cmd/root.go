package cmd

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
	"github.com/spf13/cobra"

	f "github.com/hlts2/gweather/internal/fetcher"
)

var roodCmd = &cobra.Command{
	Use:     "gweater",
	Short:   "CLI tool for acquiring weather information regularly",
	Version: "v1.0.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.WithStack(run(cmd, args))
	},
}

// WeatherInfoURL is a variable to mock the URL.
var WeatherInfoURL = f.URL

func run(cmd *cobra.Command, args []string) (rerr error) {
	fetcher, pool := f.New(), redis.New(host)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		pool.Close()
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	t := time.NewTicker(time.Duration(second) * time.Second)
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

			mm, err := fetcher.Fetch(ctx, WeatherInfoURL)
			if err != nil {
				glg.Errorf("faild to fetch contents: %v", err)
			}

			conn, err := pool.GetContext(ctx)
			if err != nil {
				cancel()
				rerr = errors.Wrap(err, "faild to get redis connection")
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

var (
	second uint
	host   string
)

func init() {
	roodCmd.PersistentFlags().UintVarP(&second, "second", "s", 180, "Interval to get weather information")
	roodCmd.PersistentFlags().StringVar(&host, "host", "redis://127.0.0.1:6379", "Host address for Redis")
}

// Execute executes cli application.
func Execute() {
	glg.Info("Start cli application")

	if err := roodCmd.Execute(); err != nil {
		glg.Error("exit app because an error occurred: %v", err)
		os.Exit(1)
	}

	glg.Info("Finish cli application")
}
