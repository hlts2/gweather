package fetcher

import (
	"context"
	"net/http"
	"sync"

	xj "github.com/basgys/goxml2json"
	"github.com/hlts2/gson"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const (
	// URL is Anytime update URL.
	// see: http://xml.kishou.go.jp/xmlpull.html
	URL = "http://www.data.jma.go.jp/developer/xml/feed/extra.xml"
)

// WeatherInfomationFetcher represents an interface to fetch weather implementation.
type WeatherInfomationFetcher interface {
	Fetch(ctx context.Context, url string) (map[string]map[string]interface{}, error)
}

type wetherInfomationFetcherImpl struct {
	mu sync.Mutex
}

// New returns WetherInfomationFetcher implementation(*wetherInfomationFetcherImpl).
func New() WeatherInfomationFetcher {
	return new(wetherInfomationFetcherImpl)
}

func createGsonFromURL(url string) (*gson.Gson, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "faild to get response, URL: %s", url)
	}

	buf, err := xj.Convert(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "faild to convert")
	}

	g, err := gson.CreateWithBytes(buf.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "faild to create gson object")
	}

	return g, nil
}

func (w *wetherInfomationFetcherImpl) Fetch(ctx context.Context, url string) (map[string]map[string]interface{}, error) {
	g, err := createGsonFromURL(url)
	if err != nil {
		return nil, errors.Wrapf(err, "faild to create gson from url: %v", URL)
	}

	r, err := g.GetByKeys("feed", "entry")
	if err != nil {
		return nil, errors.Wrapf(err, "faild to get by keys: %v", []string{"feed", "entry"})
	}

	mm := make(map[string]map[string]interface{})

	errCh := make(chan error)

	var wg sync.WaitGroup
	for _, v := range r.Slice() {
		wg.Add(1)
		go func(v *gson.Result) {
			defer wg.Done()
			rm := v.Map()

			m := make(map[string]interface{})

			title, name := rm["title"].String(), rm["author"].Map()["name"].String()

			m["title"] = title
			m["name"] = name
			m["updated"] = rm["updated"].String()
			m["content"] = rm["content"].Map()["#content"].String()

			g, err = createGsonFromURL(rm["link"].Map()["-href"].String())
			if err != nil {
				errCh <- err
				return
			}

			r, err = g.GetByKeys("Report", "Body", "Warning")
			if err != nil {
				errCh <- err
				return
			}
			m["body"] = r.Interface()

			w.mu.Lock()
			// e.g) 気象特別警報・警報・注意報_鳥取地方気象台
			mm[title+"_"+name] = m
			w.mu.Unlock()
		}(v)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var merr error
	for err := range errCh {
		merr = multierr.Append(merr, err)
	}
	return mm, merr
}
