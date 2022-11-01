package gitea

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"sync"

	gclient "code.gitea.io/sdk/gitea"
	"github.com/spf13/viper"
)

type Client struct {
	serverURL  string
	token      string
	giteapages string
	gc         *gclient.Client
}

func NewClient(serverURL, token, giteapages string) (*Client, error) {
	if giteapages == "" {
		giteapages = "gitea-pages"
	}

	gc, err := gclient.NewClient(serverURL, gclient.SetToken(token), gclient.SetGiteaVersion(""))
	if err != nil {
		return nil, err
	}

	return &Client{
		serverURL:  serverURL,
		token:      token,
		gc:         gc,
		giteapages: giteapages,
	}, nil
}

func (c *Client) Open(name, ref string) (fs.File, error) {
	if ref == "" {
		ref = c.giteapages
	}

	owner, repo, filepath, err := splitName(name)
	if err != nil {
		return nil, err
	}

	if !c.allowsPages(owner, repo) {
		return nil, fs.ErrNotExist
	}

	if err := c.readConfig(owner, repo); err != nil {
		return nil, err
	}

	if !validRefs(ref) {
		return nil, fs.ErrNotExist
	}

	res, err := c.getRawFileOrLFS(owner, repo, filepath, ref)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(filepath, ".md") {
		res, err = handleMD(res)
		if err != nil {
			return nil, err
		}
	}

	return &openFile{
		content: res,
		name:    filepath,
	}, nil
}

func (c *Client) getRawFileOrLFS(owner, repo, filepath, ref string) ([]byte, error) {
	var (
		giteaURL string
		err      error
	)

	// TODO: make pr for go-sdk
	// gitea sdk doesn't support "media" type for lfs/non-lfs
	giteaURL, err = url.JoinPath(c.serverURL+"/api/v1/repos/", owner, repo, "media", filepath)
	if err != nil {
		return nil, err
	}

	giteaURL += "?ref=" + url.QueryEscape(ref)

	req, err := http.NewRequest(http.MethodGet, giteaURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "token "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fs.ErrNotExist
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("unexpected status code '%d'", resp.StatusCode)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return res, nil

}

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func handleMD(res []byte) ([]byte, error) {
	meta, resbody, err := extractFrontMatter(string(res))
	if err != nil {
		return nil, err
	}

	resmd, err := markdown([]byte(resbody))
	if err != nil {
		return nil, err
	}

	res = append([]byte("<!DOCTYPE html>\n<html>\n<body>\n<h1>"), []byte(meta["title"].(string))...)
	res = append(res, []byte("</h1>")...)
	res = append(res, resmd...)
	res = append(res, []byte("</body></html>")...)

	return res, nil
}

func (c *Client) repoTopics(owner, repo string) ([]string, error) {
	repos, _, err := c.gc.ListRepoTopics(owner, repo, gclient.ListRepoTopicsOptions{})
	return repos, err
}

func (c *Client) allowsPages(owner, repo string) bool {
	topics, err := c.repoTopics(owner, repo)
	if err != nil {
		return false
	}

	for _, topic := range topics {
		if topic == c.giteapages {
			return true
		}
	}

	return false
}

func (c *Client) readConfig(owner, repo string) error {
	cfg, err := c.getRawFileOrLFS(owner, repo, c.giteapages+".toml", c.giteapages)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	viper.SetConfigType("toml")
	viper.ReadConfig(bytes.NewBuffer(cfg))

	return nil
}

func splitName(name string) (string, string, string, error) {
	parts := strings.Split(name, "/")

	// parts contains: ["owner", "repo", "filepath"]
	// return invalid path if not enough parts
	if len(parts) < 3 {
		return "", "", "", fs.ErrInvalid
	}

	owner := parts[0]
	repo := parts[1]
	filepath := strings.Join(parts[2:], "/")

	return owner, repo, filepath, nil
}

func validRefs(ref string) bool {
	validrefs := viper.GetStringSlice("allowedrefs")
	for _, r := range validrefs {
		if r == ref {
			return true
		}

		if r == "*" {
			return true
		}
	}

	return false
}
