package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/hlts2/gweather/cmd"
	"github.com/kpango/glg"
)

func main() {
	srv := http.Server{
		Addr:    ":1102",
		Handler: http.FileServer(http.Dir("test_datas")),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			glg.Fail(err)
		}
	}()

	// set mock data.
	os.Args = []string{"", "-s", "1", "--host", "redis://127.0.0.1:1111"}
	cmd.WeatherInfoURL = "http://127.0.0.1:1102/test_1.xml"

	cmd.Execute()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		glg.Errorf("faild to shutdown: %v", err)
	}
}
