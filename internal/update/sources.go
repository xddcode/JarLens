package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Source string

const (
	SourceGitHub Source = "github"
	SourceCNB    Source = "cnb"

	GitHubReleases = "https://api.github.com/repos/xddcode/JarLens/releases"
	CNBReleases    = "https://cnb.cool/xddcode/JarLens/-/releases"
)

type ReleasePage struct {
	Releases []Release
	Page     int
	Total    int
	HasMore  bool
}

func ParseSource(value string) Source {
	if Source(value) == SourceCNB {
		return SourceCNB
	}
	return SourceGitHub
}

func (source Source) ReleasesURL() string {
	if source == SourceCNB {
		return CNBReleases
	}
	return "https://github.com/xddcode/JarLens/releases"
}

func LatestFromSource(ctx context.Context, client *http.Client, source Source) (Release, error) {
	if ParseSource(string(source)) == SourceGitHub {
		return Latest(ctx, client, Feed)
	}
	page, err := Releases(ctx, client, SourceCNB, "", 1, 20)
	if err != nil {
		return Release{}, err
	}
	for _, release := range page.Releases {
		if release.Latest {
			return release, nil
		}
	}
	if len(page.Releases) > 0 {
		return page.Releases[0], nil
	}
	return Release{}, ErrNoRelease
}

func Releases(ctx context.Context, client *http.Client, source Source, endpoint string, page, pageSize int) (ReleasePage, error) {
	source = ParseSource(string(source))
	page = max(1, page)
	pageSize = max(1, min(50, pageSize))
	if endpoint == "" {
		endpoint = GitHubReleases
		if source == SourceCNB {
			endpoint = CNBReleases
		}
	}
	address, err := url.Parse(endpoint)
	if err != nil {
		return ReleasePage{}, err
	}
	query := address.Query()
	query.Set("page", strconv.Itoa(page))
	if source == SourceCNB {
		query.Set("page_size", strconv.Itoa(pageSize))
	} else {
		query.Set("per_page", strconv.Itoa(pageSize))
	}
	address.RawQuery = query.Encode()
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address.String(), nil)
	if err != nil {
		return ReleasePage{}, err
	}
	request.Header.Set("User-Agent", "JarLens/"+Current)
	if source == SourceCNB {
		request.Header.Set("Accept", "text/html")
	} else {
		request.Header.Set("Accept", "application/vnd.github+json")
	}
	response, err := client.Do(request)
	if err != nil {
		return ReleasePage{}, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return ReleasePage{}, ErrNoRelease
	}
	if response.StatusCode != http.StatusOK {
		return ReleasePage{}, fmt.Errorf("update: release feed answered %s", response.Status)
	}
	if source == SourceCNB {
		return readCNBReleases(response.Body, page, pageSize)
	}
	return readGitHubReleases(response, page, pageSize)
}

type releasePayload struct {
	Tag         string `json:"tag_name"`
	URL         string `json:"html_url"`
	Notes       string `json:"body"`
	PublishedAt string `json:"published_at"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	Latest      bool   `json:"is_latest"`
}

func readGitHubReleases(response *http.Response, page, pageSize int) (ReleasePage, error) {
	var body []releasePayload
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&body); err != nil {
		return ReleasePage{}, fmt.Errorf("update: reading release list: %w", err)
	}
	result := ReleasePage{Page: page, HasMore: len(body) == pageSize || strings.Contains(response.Header.Get("Link"), `rel="next"`)}
	for _, item := range body {
		release, ok := makeRelease(item)
		if ok {
			result.Releases = append(result.Releases, release)
		}
	}
	return result, nil
}

func readCNBReleases(reader io.Reader, page, pageSize int) (ReleasePage, error) {
	body, err := io.ReadAll(io.LimitReader(reader, 8<<20))
	if err != nil {
		return ReleasePage{}, fmt.Errorf("update: reading CNB releases: %w", err)
	}
	const marker = `<script id="__NEXT_DATA__"`
	start := strings.Index(string(body), marker)
	if start < 0 {
		return ReleasePage{}, errors.New("update: CNB release data is missing")
	}
	openingEnd := strings.Index(string(body[start:]), ">")
	if openingEnd < 0 {
		return ReleasePage{}, errors.New("update: CNB release data is incomplete")
	}
	contentStart := start + openingEnd + 1
	contentEnd := strings.Index(string(body[contentStart:]), "</script>")
	if contentEnd < 0 {
		return ReleasePage{}, errors.New("update: CNB release data is incomplete")
	}
	var document struct {
		Props struct {
			PageProps struct {
				InitialState struct {
					Slug struct {
						Repo struct {
							Releases struct {
								List struct {
									Data struct {
										Releases     []releasePayload `json:"releases"`
										ReleaseCount int              `json:"release_count"`
									} `json:"data"`
								} `json:"list"`
							} `json:"releases"`
						} `json:"repo"`
					} `json:"slug"`
				} `json:"initialState"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal(body[contentStart:contentStart+contentEnd], &document); err != nil {
		return ReleasePage{}, fmt.Errorf("update: decoding CNB releases: %w", err)
	}
	data := document.Props.PageProps.InitialState.Slug.Repo.Releases.List.Data
	result := ReleasePage{Page: page, Total: data.ReleaseCount, HasMore: page*pageSize < data.ReleaseCount}
	for _, item := range data.Releases {
		item.URL = CNBReleases + "/tag/" + url.PathEscape(item.Tag)
		release, ok := makeRelease(item)
		if ok {
			result.Releases = append(result.Releases, release)
		}
	}
	return result, nil
}

func makeRelease(body releasePayload) (Release, bool) {
	version := strings.TrimPrefix(strings.TrimSpace(body.Tag), "v")
	if _, ok := parseVersion(version); !ok || body.Draft || body.Prerelease {
		return Release{}, false
	}
	if !strings.HasPrefix(body.URL, "https://") {
		return Release{}, false
	}
	publishedAt, _ := time.Parse(time.RFC3339, body.PublishedAt)
	return Release{
		Version: version, URL: body.URL, Notes: cutUTF8(body.Notes, 16<<10),
		PublishedAt: publishedAt, Latest: body.Latest,
	}, true
}
