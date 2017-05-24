package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Sites struct {
	url    string
	images []string
	folder string
}

var crawlers sync.WaitGroup
var downloaders sync.WaitGroup
var verbose bool = false

func (Site *Sites) Crawl() {
	defer crawlers.Done()

	resp, err := goquery.NewDocument(Site.url)
	if err != nil {
		fmt.Printf("ERROR: Failed to crawl \"" + Site.url + "\"\n\n")
		os.Exit(3)
	}
	// use CSS selector found with the browser inspector
	// for each, use index and item
	resp.Find("*").Each(func(index int, item *goquery.Selection) {
		linkTag := item.Find("img")
		link, _ := linkTag.Attr("src")

		if link != "" {
			Site.images = append(Site.images, link)
		}
	})

	fmt.Printf("%s found %d unique images\n", Site.url, len(Site.images))

	pool := len(Site.images) / 3
	if pool > 10 {
		pool = 10
	}

	l := 0
	counter := len(Site.images) / pool

	for i := counter; i < len(Site.images); i += counter {
		downloaders.Add(1)
		go Site.DownloadImg(Site.images[l:i])
		l = i
	}

	downloaders.Wait()
}

func (Site *Sites) DownloadImg(images []string) {

	defer downloaders.Done()

	os.Mkdir(Site.folder, os.FileMode(0777))

	Site.images = SliceUniq(images)

	for _, url := range Site.images {
		if url[:4] != "http" {
			url = "http:" + url
		}
		parts := strings.Split(url, "/")
		name := parts[len(parts)-1]
		file, _ := os.Create(string(Site.folder + "/" + name))
		resp, _ := http.Get(url)
		io.Copy(file, resp.Body)
		file.Close()
		resp.Body.Close()
		if verbose == true {
		fmt.Printf("Saving %s \n", Site.folder+"/"+name)
		}
	}
}

func SliceUniq(s []string) []string {
	for i := 0; i < len(s); i++ {
		for i2 := i + 1; i2 < len(s); i2++ {
			if s[i] == s[i2] {
				// delete
				s = append(s[:i2], s[i2+1:]...)
				i2--
			}
		}
	}
	return s
}

func main() {


	var seedUrls []string

	if len(os.Args) < 2 {
		fmt.Println("ERROR : Less Args\nCommand should be of type : imagescraper [websites]\n\n")
		os.Exit(3)
	}
	if os.Args[1] == "-v" || os.Args[1] == "--verbose" {
			verbose = true
			seedUrls = os.Args[2:]
	}else {
		seedUrls = os.Args[1:]
	}
	Site := make([]Sites, len(seedUrls))

	// Crawl process (concurrently)
	for i, name := range seedUrls {
		if name[:4] != "http" {
			name = "http://" + name
		}
		u, err := url.Parse(name)
		if err != nil {
			fmt.Printf("could not fetch page - %s %v", name, err)
		}
		Site[i].folder = u.Host
		Site[i].url = name
		crawlers.Add(1)
		go Site[i].Crawl()
	}

	crawlers.Wait()

	fmt.Printf("\n\nScraped succesfully\n\n")

}
