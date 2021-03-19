package redirector

import (
	denet "github.com/hlandau/goutils/net"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestRedirector(t *testing.T) {
	dir, err := ioutil.TempDir("", "acme-redirector-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)

	r, err := New(Config{
		Bind:          ":9847",
		ChallengePath: dir,
		StatusCode:    308,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer r.Stop()

	req, err := http.NewRequest("FROBNICATE", "http://127.0.0.1:9847/foo/bar?alpha=beta", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()
	loc := res.Header.Get("Location")
	if loc != "https://127.0.0.1:9847/foo/bar?alpha=beta" {
		t.Fatalf("wrong Location: %v", loc)
	}

	err = ioutil.WriteFile(filepath.Join(dir, "foo"), []byte("bar"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", "http://127.0.0.1:9847/.well-known/acme-challenge/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(denet.LimitReader(res.Body, 1*1024*1024))
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "bar" {
		t.Fatal("wrong response")
	}
}
