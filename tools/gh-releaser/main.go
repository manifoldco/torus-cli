package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
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
	app := cli.NewApp()
	app.Name = "gh-releaser"
	app.HelpName = "gh-releaser"
	app.Usage = "Simple tooling for creating and modifying github releases"
	app.Commands = []cli.Command{
		{
			Name:   "reindex",
			Usage:  "Create all missing production releases based on the CHANGELOG.md",
			Action: reindex,
		},
		{
			Name:      "upload",
			ArgsUsage: "<version> <asset-folder>",
			Usage:     "Attach files in the given folder to the specified release",
			Action:    upload,
		},
	}
	app.Run(os.Args)
}

var (
	owner = "manifoldco"
	repo  = "torus-cli"
)

func reindex(_ *cli.Context) error {
	c, err := newGithubClient()
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	changelog, err := loadChangelog()
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	tags, err := fetchTags(c)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	releases, err := fetchReleases(c)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	sort.Sort(tagSorter(tags))
	sort.Sort(releaseSorter(releases))

	tagsWithNoRelease := pruneExistingReleases(releases, tags)
	err = createMissingReleases(c, tagsWithNoRelease, changelog)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	return nil
}

func upload(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		return cli.NewExitError("Two arguments must be provided", -1)
	}

	version := args[0]
	folder := args[1]

	if isPrerelease(version) {
		return cli.NewExitError("Must specify valid semver or release cannot be a release candidate", -1)
	}

	c, err := newGithubClient()
	if err != nil {
		return err
	}

	releases, err := fetchReleases(c)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}
	release := findRelease(version, releases)
	if release == nil {
		return cli.NewExitError("Could not find release for tag: "+version, -1)
	}

	err = attachToRelease(c, release, folder)
	if err != nil {
		return cli.NewExitError("Could not upload assets: "+err.Error(), -1)
	}

	return nil
}

func newGithubClient() (*github.Client, error) {
	tok, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return nil, errors.New("Missing GITHUB_TOKEN environment variable")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(tc), nil
}

func loadChangelog() ([]section, error) {
	var changelog []section
	cl, err := ioutil.ReadFile("CHANGELOG.md")
	if err != nil {
		return changelog, err
	}

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

	return changelog, nil
}

func fetchTags(c *github.Client) ([]*github.RepositoryTag, error) {
	var tags []*github.RepositoryTag
	opt := &github.ListOptions{PerPage: 100}
	for {
		tagPage, resp, err := c.Repositories.ListTags(owner, repo, opt)
		if err != nil {
			return tags, err
		}

		tags = append(tags, tagPage...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return tags, nil
}

func fetchReleases(c *github.Client) ([]*github.RepositoryRelease, error) {
	var releases []*github.RepositoryRelease
	opt := &github.ListOptions{PerPage: 100}
	for {
		releasePage, resp, err := c.Repositories.ListReleases(owner, repo, opt)
		if err != nil {
			return releases, err
		}

		releases = append(releases, releasePage...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return releases, nil
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

func findRelease(tag string, releases []*github.RepositoryRelease) *github.RepositoryRelease {
	for _, r := range releases {
		if semverCmp(tag, *r.TagName) == 0 {
			return r
		}
	}

	return nil
}

func createMissingReleases(c *github.Client, tagsWithNoRelease []*github.RepositoryTag, changelog []section) error {
	for _, tag := range tagsWithNoRelease {
		body := "*No changelog entry*"
		for _, sec := range changelog {
			if sec.tag == *tag.Name {
				body = publishedRegex.ReplaceAllString(sec.body, "")
				break
			}
		}

		if isPrerelease(*tag.Name) { // we don't create releases for RCs
			continue
		}

		rel := &github.RepositoryRelease{
			Name:       github.String(*tag.Name),
			TagName:    github.String(*tag.Name),
			Prerelease: github.Bool(false),
			Body:       github.String(body),
		}

		log.Println("Creating release", *rel.Name)
		_, _, err := c.Repositories.CreateRelease(owner, repo, rel)
		if err != nil {
			return err
		}
	}

	return nil
}

func attachToRelease(c *github.Client, release *github.RepositoryRelease, folder string) error {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return err
	}

	for _, f := range files {
		fd, err := os.Open(path.Join(folder, f.Name()))
		if err != nil {
			return err
		}

		opts := &github.UploadOptions{Name: f.Name()}
		_, _, err = c.Repositories.UploadReleaseAsset(owner, repo, *release.ID, opts, fd)
		if err != nil {
			return err
		}
	}

	return nil
}

func isPrerelease(name string) bool {
	ver, err := semver.ParseTolerant(name)
	if err != nil { // if we can't parse then its not a release tag
		return true
	}

	return len(ver.Pre) > 0
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
