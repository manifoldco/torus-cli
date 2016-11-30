package main

import (
	"bytes"
	"flag"
	"log"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var t *template.Template

var (
	bucket    string
	rawBucket = flag.String("bucket", "", "generate for this bucket")
)

func init() {
	flag.Parse()

	bucket = *rawBucket
	if len(bucket) > 4 && bucket[:5] == "s3://" {
		bucket = bucket[5:]
	}

	prefix := "tools/cdn-indexer/"
	name := filepath.Join(prefix, "index.html.tmpl")
	t = template.Must(template.New("index.html.tmpl").Funcs(template.FuncMap{
		"filterZero": func(v interface{}) interface{} {
			if v == reflect.Zero(reflect.TypeOf(v)).Interface() {
				return ""
			}

			return v
		},
	}).ParseFiles(name))
}

type fileType string

const (
	directoryType fileType = "directory"
	genericType            = "generic"
)

type cdnFile struct {
	Name     string
	Type     fileType
	Size     int64
	Modified time.Time
}

// Listings are sorted lexicographically, with directories coming before
// files.
type sortCDNFile []cdnFile

func (d sortCDNFile) Len() int      { return len(d) }
func (d sortCDNFile) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d sortCDNFile) Less(i, j int) bool {
	if d[i].Type != d[j].Type {
		return d[i].Type == directoryType
	}

	return strings.Compare(d[i].Name, d[j].Name) < 0
}

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("failed to create session,", err)
		return
	}

	svc := s3.New(sess)

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket), // Required
	}

	listings := make(map[string][]cdnFile)

	err = svc.ListObjectsV2Pages(params, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, o := range page.Contents {
			f := cdnFile{
				Name:     filepath.Base(*o.Key),
				Type:     genericType,
				Size:     *o.Size,
				Modified: *o.LastModified,
			}

			// index pages shouldn't list themselves.
			if f.Name == "index.html" {
				continue
			}

			dir := filepath.Dir(*o.Key)
			listings[dir] = append(listings[dir], f)
		}
		return true
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add dirs into their parent listings
	for dir := range listings {
		for {
			f := cdnFile{
				Name: filepath.Base(dir),
				Type: directoryType,
			}
			if f.Name == "." {
				break
			}

			dir = filepath.Dir(dir)

			found := false
			for _, o := range listings[dir] {
				if o.Name == f.Name {
					found = true
					break
				}
			}
			if !found {
				listings[dir] = append(listings[dir], f)
			}
		}
	}

	now := time.Now()
	for dir, listing := range listings {
		if dir == "." {
			dir = "/"
		}

		sort.Sort(sortCDNFile(listing))

		buf := &bytes.Buffer{}
		err = t.Execute(buf, struct {
			Dir       string
			Files     []cdnFile
			Timestamp time.Time
		}{Dir: dir, Files: listing, Timestamp: now})

		if err != nil {
			log.Fatal(err)
		}

		dirPath := dir

		params := &s3.PutObjectInput{
			Bucket:       aws.String(bucket),
			Key:          aws.String(filepath.Join(dirPath, "index.html")),
			Body:         bytes.NewReader(buf.Bytes()),
			ContentType:  aws.String("text/html"),
			CacheControl: aws.String("public, max-age=300"),
		}
		_, err := svc.PutObject(params)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote /%s\n", filepath.Join(dirPath, "index.html"))
	}
}
