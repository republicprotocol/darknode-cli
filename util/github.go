package util

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
)

// Initialize the github client. If an access token has been set in a environment,
// it will use it for oauth to avoid rate limiting.
func GithubClient(ctx context.Context) *github.Client {
	accessToken := os.Getenv("ACCESS_TOKEN")
	var client *http.Client
	if accessToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		)
		client = oauth2.NewClient(ctx, ts)
	}

	return github.NewClient(client)
}

// CurrentReleaseVersion queries the Github API and fetch the latest release version of darknode-cli.
func CurrentReleaseVersion(ctx context.Context) (*version.Version, error) {
	client := GithubClient(ctx)
	release, response, err := client.Repositories.GetLatestRelease(ctx, "renproject", "darknode-cli")
	if err != nil {
		return nil, err
	}

	// Verify the status code is 200.
	if err := VerifyStatusCode(response.Response, http.StatusOK); err != nil {
		return nil, err
	}
	return version.NewVersion(release.GetTagName())
}

func CliLatestVersion()(*version.Version, error) {
	resp, err := http.Get("https://www.github.com/renproject/darknode-cli/releases/latest")
	if err != nil {
		return nil, err
	}

	return version.NewVersion(path.Base(resp.Request.URL.String()))
}

// LatestStableRelease checks the node release repo and return the version of the latest release.
func LatestStableRelease() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := GithubClient(ctx)
	releases, response, err := client.Repositories.ListReleases(ctx, "renproject", "darknode-release", nil)
	if err != nil {
		return "", err
	}

	// Verify the status code is 200.
	if err := VerifyStatusCode(response.Response, http.StatusOK); err != nil {
		return "", err
	}

	// Find the latest stable release
	latest, err := version.NewVersion("0.0.0")
	if err != nil {
		return "", err
	}
	stableVer, err := regexp.Compile("^v?[0-9]+\\.[0-9]+\\.[0-9]+$")
	if err != nil {
		return "", err
	}
	for _, release := range releases {
		if stableVer.MatchString(*release.TagName) {
			ver, err := version.NewVersion(*release.TagName)
			if err != nil {
				continue
			}
			if ver.GreaterThan(latest) {
				latest = ver
			}
		}
	}
	if latest.String() == "0.0.0" {
		return "", errors.New("cannot find any stable release")
	}

	return latest.String(), nil
}
