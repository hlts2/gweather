package gwether

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

func job(ctx context.Context, url string) error {
	g, err := createGsonFromURL(url)
	if err != nil {
		return errors.Wrapf(err, "faild to create gson from url: %v", URL)
	}

	r, err := g.GetByKeys("feed", "entry")
	if err != nil {
		return errors.Wrapf(err, "faild to get by keys, keys: %v", []string{"feed", "entry"})
	}

	errCh := make(chan error)

	var wg sync.WaitGroup
	for _, v := range r.Slice() {
		wg.Add(1)
		go func(v *gson.Result) {
			defer wg.Done()
			rm := v.Map()

			m := make(map[string]interface{})

			m["title"] = rm["title"].String()
			m["name"] = rm["author"].Map()["name"].String()
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
	return merr
}
