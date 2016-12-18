package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Regexps for finding and removing the version tag and published date in
// the changelog
var (
	sectionHeader  = regexp.MustCompile(`(?m:^##\s+(v[0-9]+\.[0-9]+\.[0-9]+)\s*$)`)
	publishedRegex = regexp.MustCompile(`(?m:\n^_[0-9]{4}-[0-9]{1,2}-[0-9]+_\s*\n\s*\n)`)
)

// section holds CHANGELOG.md sections, mapping the release tag to the body
// contents.
type section struct {
	tag  string
	body string
}

func main() {
	c := newGithubClient()
	changelog := loadChangelog()
	tags := fetchTags(c)
	releases := fetchReleases(c)

	sort.Sort(tagSorter(tags))
	sort.Sort(releaseSorter(releases))

	tagsWithNoRelease := pruneExistingReleases(releases, tags)
	createMissingReleases(c, tagsWithNoRelease, changelog)
}

func newGithubClient() *github.Client {
	tok, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatal("Please set GITHUB_TOKEN")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(tc)
}

func loadChangelog() []section {
	cl, err := ioutil.ReadFile("CHANGELOG.md")
	if err != nil {
		log.Fatal(err)
	}

	var changelog []section
	parts := sectionHeader.FindAllSubmatchIndex(cl, -1)
	for i, part := range parts {
		var end int
		if i+1 == len(parts) {
			end = len(cl)
		} else {
			end = parts[i+1][0]
		}

		sec := section{
			tag:  string(cl[part[2]:part[3]]),
			body: string(cl[part[1]:end]),
		}
		changelog = append([]section{sec}, changelog...)
	}

	return changelog
}

func fetchTags(c *github.Client) []*github.RepositoryTag {
	var tags []*github.RepositoryTag
	opt := &github.ListOptions{PerPage: 100}
	for {
		tagPage, resp, err := c.Repositories.ListTags("manifoldco", "torus-cli", opt)
		if err != nil {
			log.Fatal(err)
		}

		tags = append(tags, tagPage...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return tags
}

func fetchReleases(c *github.Client) []*github.RepositoryRelease {
	var releases []*github.RepositoryRelease
	opt := &github.ListOptions{PerPage: 100}
	for {
		releasePage, resp, err := c.Repositories.ListReleases("manifoldco", "torus-cli", opt)
		if err != nil {
			log.Fatal(err)
		}

		releases = append(releases, releasePage...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}
	return releases
}

func pruneExistingReleases(releases []*github.RepositoryRelease, tags []*github.RepositoryTag) []*github.RepositoryTag {
	var tagsWithNoRelease []*github.RepositoryTag

	i := 0
outerLoop:
	for _, release := range releases {
		for i < len(tags) {
			tag := tags[i]

			switch semverCmp(*release.Name, *tag.Name) {
			case 1: // no release for this tag
				tagsWithNoRelease = append(tagsWithNoRelease, tag)
				i++
			case 0: // Found it
				i++
				continue outerLoop
			default: // no tag for this release (we ignore this case)
				continue outerLoop
			}
		}

	}

	tagsWithNoRelease = append(tagsWithNoRelease, tags[i:]...)

	return tagsWithNoRelease
}

func createMissingReleases(c *github.Client, tagsWithNoRelease []*github.RepositoryTag, changelog []section) {
	for _, tag := range tagsWithNoRelease {
		body := "*No changelog entry*"
		for _, sec := range changelog {
			if sec.tag == *tag.Name {
				body = publishedRegex.ReplaceAllString(sec.body, "")
				break
			}
		}

		ver, err := semver.ParseTolerant(*tag.Name)
		if err != nil { // If its not a valid semver, its not a tag we release
			continue
		}

		prerelease := len(ver.Pre) > 0
		if prerelease { // We don't create real releases for release candidates.
			continue
		}

		rel := &github.RepositoryRelease{
			Name:       github.String(*tag.Name),
			TagName:    github.String(*tag.Name),
			Prerelease: github.Bool(prerelease),
			Body:       github.String(body),
		}

		log.Println("Creating release", *rel.Name)
		_, _, err = c.Repositories.CreateRelease("manifoldco", "torus-cli", rel)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func semverCmp(v1s, v2s string) int {
	v1, err1 := semver.ParseTolerant(v1s)
	v2, err2 := semver.ParseTolerant(v2s)

	// unparseable semvers are 'less' than regulars. Two unparseable semvers
	// are compared as regular strings.
	switch {
	case err1 == nil && err2 == nil:
		return v1.Compare(v2)
	case err1 != nil && err2 != nil:
		return strings.Compare(v1s, v2s)
	case err1 != nil:
		return -1
	default:
		return 1
	}
}

// Sorters for tags and github releases. Elements are sorted in ascending order,
// based on semver. tags that can't be parsed as semvers are ordered before
// parseable tags.

type tagSorter []*github.RepositoryTag

func (t tagSorter) Len() int           { return len(t) }
func (t tagSorter) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tagSorter) Less(i, j int) bool { return semverCmp(*t[i].Name, *t[j].Name) < 0 }

type releaseSorter []*github.RepositoryRelease

func (r releaseSorter) Len() int           { return len(r) }
func (r releaseSorter) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r releaseSorter) Less(i, j int) bool { return semverCmp(*r[i].Name, *r[j].Name) < 0 }
