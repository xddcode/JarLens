package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubReleaseHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("per_page") != "5" {
			t.Errorf("unexpected pagination: %s", r.URL.RawQuery)
		}
		w.Header().Set("Link", `<https://example.com/releases?page=3>; rel="next"`)
		fmt.Fprint(w, `[
			{"tag_name":"v1.2.0","html_url":"https://github.com/xddcode/JarLens/releases/tag/v1.2.0","body":"Improvements","published_at":"2026-11-07T08:00:00Z"},
			{"tag_name":"v1.3.0-rc.1","html_url":"https://github.com/xddcode/JarLens/releases/tag/v1.3.0-rc.1","prerelease":true}
		]`)
	}))
	defer server.Close()
	page, err := Releases(context.Background(), server.Client(), SourceGitHub, server.URL, 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 2 || !page.HasMore || len(page.Releases) != 1 {
		t.Fatalf("page = %+v", page)
	}
	release := page.Releases[0]
	if release.Version != "1.2.0" || release.Notes != "Improvements" || release.PublishedAt.Format(time.DateOnly) != "2026-11-07" {
		t.Fatalf("release = %+v", release)
	}
}

func TestCNBReleaseHistoryFromPublicPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("page_size") != "5" {
			t.Errorf("unexpected pagination: %s", r.URL.RawQuery)
		}
		fmt.Fprint(w, `<html><head><script id="__NEXT_DATA__" type="application/json" nonce="test">{"props":{"pageProps":{"initialState":{"slug":{"repo":{"releases":{"list":{"data":{"release_count":8,"releases":[{"tag_name":"v1.2.0","body":"CNB improvements","published_at":"2026-11-07T08:00:00Z","is_latest":true},{"tag_name":"v1.3.0-rc.1","body":"Preview","prerelease":true}]}}}}}}}}}</script></head></html>`)
	}))
	defer server.Close()
	page, err := Releases(context.Background(), server.Client(), SourceCNB, server.URL, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 8 || !page.HasMore || len(page.Releases) != 1 {
		t.Fatalf("page = %+v", page)
	}
	release := page.Releases[0]
	if release.Version != "1.2.0" || !release.Latest || release.Notes != "CNB improvements" || release.URL != CNBReleases+"/tag/v1.2.0" {
		t.Fatalf("release = %+v", release)
	}
}

func TestSourceReleasePages(t *testing.T) {
	if SourceCNB.ReleasesURL() != CNBReleases || SourceGitHub.ReleasesURL() == CNBReleases {
		t.Fatal("source-specific release pages are incorrect")
	}
	if ParseSource("unknown") != SourceGitHub || ParseSource("cnb") != SourceCNB {
		t.Fatal("source normalization is incorrect")
	}
}
