// Command landing-page builds _site/index.html: a plain list of live GitHub
// Pages sites for a user.
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type repo struct {
	Name     string `json:"name"`
	HasPages bool   `json:"has_pages"`
	Private  bool   `json:"private"`
}

type site struct {
	Name string
	URL  string
}

var page = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{.User}}</title>
</head>
<body>
<h1>{{.User}}</h1>
<ul>
{{range .Sites}}<li><a href="{{.URL}}">{{.Name}}</a></li>
{{end}}</ul>
</body>
</html>
`))

func main() {
	user := os.Getenv("GH_USER")
	if user == "" {
		user = "adamratson"
	}
	self := strings.ToLower(user + ".github.io")

	repos, err := fetchRepos(user)
	if err != nil {
		log.Fatal(err)
	}

	var sites []site
	for _, r := range repos {
		if !r.HasPages || r.Private || strings.ToLower(r.Name) == self {
			continue
		}
		url := fmt.Sprintf("https://%s.github.io/%s/", user, r.Name)
		if isLive(url) {
			sites = append(sites, site{r.Name, url})
		}
	}

	if err := os.MkdirAll("_site", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(filepath.Join("_site", "index.html"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := page.Execute(f, struct {
		User  string
		Sites []site
	}{user, sites}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote %d sites to _site/index.html\n", len(sites))
}

func fetchRepos(user string) ([]repo, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&type=owner&sort=updated", user)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	var repos []repo
	return repos, json.NewDecoder(resp.Body).Decode(&repos)
}

func isLive(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
