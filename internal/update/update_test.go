package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "JarLens/"+Current {
			t.Errorf("unexpected User-Agent %q", r.UserAgent())
		}
		fmt.Fprint(w, `{"tag_name":"v1.1.0","html_url":"https://github.com/xddcode/JarLens/releases/tag/v1.1.0","body":"- New features","draft":false,"prerelease":false}`)
	}))
	defer server.Close()
	release, err := Latest(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if release.Version != "1.1.0" || release.Notes != "- New features" || !Newer(Current, release.Version) {
		t.Fatalf("release = %+v", release)
	}
}

func TestLatestRejectsDraftAndNonHTTPSPage(t *testing.T) {
	for _, body := range []string{
		`{"tag_name":"v1.0.0","html_url":"https://example.com","draft":true}`,
		`{"tag_name":"v1.0.0","html_url":"http://example.com"}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, body) }))
		if _, err := Latest(context.Background(), server.Client(), server.URL); err == nil {
			t.Fatal("unsafe release was accepted")
		}
		server.Close()
	}
}

func TestLatestNoPublishedRelease(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	if _, err := Latest(context.Background(), server.Client(), server.URL); err != ErrNoRelease {
		t.Fatalf("error = %v, want ErrNoRelease", err)
	}
}

func TestNewerVersion(t *testing.T) {
	for _, test := range []struct {
		current, latest string
		want            bool
	}{
		{"0.1.0", "1.0.0", true}, {"1.0.0", "1.0.1", true}, {"1.0.0", "1.0.0", false},
		{"1.0.1", "1.0.0", false}, {"1.0.0", "bad", false},
	} {
		if got := Newer(test.current, test.latest); got != test.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", test.current, test.latest, got, test.want)
		}
	}
}

func TestCurrentMatchesVivaManifest(t *testing.T) {
	data, err := os.ReadFile("../../viva.toml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `version = "`+Current+`"`) {
		t.Fatalf("viva.toml does not use update.Current %q", Current)
	}
}
