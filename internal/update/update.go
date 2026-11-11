// Package update reads published releases. It downloads and installs only a
// signed package after the user explicitly chooses to update; manual download
// remains available when in-place installation is unsupported.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const Current = "0.2.0"
const Feed = "https://api.github.com/repos/xddcode/JarLens/releases/latest"

var ErrNoRelease = errors.New("update: no published release")

type Release struct {
	Version     string
	URL         string
	Notes       string
	PublishedAt time.Time
	Latest      bool
}

func Latest(ctx context.Context, client *http.Client, feed string) (Release, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, feed, nil)
	if err != nil {
		return Release{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "JarLens/"+Current)
	response, err := client.Do(request)
	if err != nil {
		return Release{}, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return Release{}, ErrNoRelease
	}
	if response.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("update: release feed answered %s", response.Status)
	}
	var body struct {
		Tag         string `json:"tag_name"`
		URL         string `json:"html_url"`
		Notes       string `json:"body"`
		PublishedAt string `json:"published_at"`
		Draft       bool   `json:"draft"`
		Prerelease  bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body); err != nil {
		return Release{}, fmt.Errorf("update: reading release feed: %w", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(body.Tag), "v")
	if _, ok := parseVersion(version); !ok || body.Draft || body.Prerelease {
		return Release{}, errors.New("update: release feed has no usable stable version")
	}
	if !strings.HasPrefix(body.URL, "https://") {
		return Release{}, errors.New("update: release page is not an HTTPS address")
	}
	publishedAt, _ := time.Parse(time.RFC3339, body.PublishedAt)
	return Release{Version: version, URL: body.URL, Notes: cutUTF8(body.Notes, 16<<10), PublishedAt: publishedAt}, nil
}

func Newer(current, latest string) bool {
	a, okA := parseVersion(strings.TrimPrefix(current, "v"))
	b, okB := parseVersion(strings.TrimPrefix(latest, "v"))
	if !okA || !okB {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return b[index] > a[index]
		}
	}
	return false
}

func parseVersion(version string) (parts [3]int, ok bool) {
	fields := strings.Split(version, ".")
	if len(fields) == 0 || len(fields) > 3 {
		return parts, false
	}
	for index, field := range fields {
		number, err := strconv.Atoi(field)
		if err != nil || number < 0 || number > 1<<20 {
			return parts, false
		}
		parts[index] = number
	}
	return parts, true
}

func cutUTF8(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	for limit > 0 && !utf8.RuneStart(text[limit]) {
		limit--
	}
	return text[:limit]
}
