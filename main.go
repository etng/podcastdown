package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	//"github.com/mmcdole/gofeed/rss"
)

var podCastMdTpl = `
# [{{.Title}}]({{.Link}})
Last Updated At: {{.Updated}}
[{{.Author.Name}}](mailto:{{.Author.Email}})
{{.Categories}}
![{{.Image.Title}}]({{.Image.URL}})
---
{{.Description}}
`
var itemMdTpl = `
# {{.Title}}
{{ .Author }} {{ .Updated }}

{{ range .Enclosures }}* {{ .URL }} {{ end }}

---
{{.Description}}
`

const downloadEnclouser = false

func AssureDir(path string) {
	if _, e := os.Stat(path); os.IsNotExist(e) {
		fmt.Printf("creating dir %s\n", path)
		os.MkdirAll(path, 0777)
	}
}
func DownloadFile(url, dest string, wg *sync.WaitGroup) {
	defer wg.Done()
	AssureDir(filepath.Dir(dest))
	outFile, _ := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	fmt.Printf("downloading %s to %s\n", url, dest)
	resp, _ := http.Get(url)
	var buffer []byte
	defer resp.Body.Close()
	io.CopyBuffer(outFile, resp.Body, buffer)
	fmt.Printf("downloaded %s to %s\n", url, dest)
}

type DownloadTask struct {
	Url  string
	Dest string
}

func main() {
	fp := gofeed.NewParser()
	var e error
	feed, e := fp.ParseURL("https://feeds.fireside.fm/surplusvalue/rss")
	if e != nil {
		panic(e)
	}
	fmt.Println(feed.Title)
	podCastTpl, _ := template.New("podcast").Parse(podCastMdTpl)
	itemTpl, _ := template.New("item").Parse(itemMdTpl)
	baseDir := "downloads/"
	dirName := filepath.Join(baseDir, feed.Title)
	var downloadTasks []DownloadTask
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mdFile := filepath.Join(dirName, fmt.Sprintf("%s.md", "README"))
		AssureDir(filepath.Dir(mdFile))
		if out, e := os.OpenFile(mdFile, os.O_WRONLY|os.O_CREATE, os.ModePerm); e != nil {
			fmt.Printf("fail to create file %s for %s\n", mdFile, e)
			return
		} else {
			defer out.Close()
			podCastTpl.Execute(out, feed)
			fmt.Printf("readme writed\n")
		}

	}()

	for _, item := range feed.Items {
		fmt.Printf("found item %s %s %s\n", item.Title, item.Updated, item.Description[:100])
		wg.Add(1)
		go func() {
			defer wg.Done()
			mdFile := filepath.Join(dirName, fmt.Sprintf("%s.md", item.Title))
			AssureDir(filepath.Dir(mdFile))
			if out, e := os.OpenFile(mdFile, os.O_WRONLY|os.O_CREATE, os.ModePerm); e != nil {
				fmt.Printf("fail to create file %s for %s\n", mdFile, e)
				return
			} else {
				defer out.Close()
				itemTpl.Execute(out, item)
				fmt.Printf("item desc writed\n")
			}

		}()
		for i, enclosure := range item.Enclosures {
			fmt.Printf("found enclosure %s %s\n", enclosure.Type, enclosure.URL)
			nameParts := strings.Split(filepath.Base(enclosure.URL), ".")
			ext := nameParts[len(nameParts)-1]
			filename := filepath.Join(dirName, fmt.Sprintf("%s_%d.%s", item.Title, i, ext))
			if downloadEnclouser {
				wg.Add(1)
				go DownloadFile(enclosure.URL, filename, &wg)
			}
			downloadTasks = append(downloadTasks, DownloadTask{
				Url:  enclosure.URL,
				Dest: filename,
			})

		}
		//break
	}
	wg.Wait()
	var lines []string
	for _, dt := range downloadTasks {
		lines = append(lines, fmt.Sprintf("%s %s", dt.Url, dt.Dest))
	}
	ioutil.WriteFile(filepath.Join(dirName, "wget.task"), []byte(strings.Join(lines, "\n")), 0666)
	fmt.Printf("done")
}
