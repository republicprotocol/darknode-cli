package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

// CurrentReleaseVersion queries the Github API and fetch the latest release version of nodectl.
func CurrentReleaseVersion(ctx context.Context) (*version.Version, error){
	// Initialize the github client
	var hClient *http.Client
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken != "" {
		token := &oauth2.Token{
			AccessToken: accessToken,
		}
		hClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	}
	client := github.NewClient(hClient)

	// Fetch the latest release
	release, response, err := client.Repositories.GetLatestRelease(ctx, "renproject", "nodectl")
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK{
		message, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("code = %v, err = %s", response.StatusCode, message)
	}
	return version.NewVersion(release.GetTagName())
}