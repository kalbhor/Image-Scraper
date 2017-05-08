package main

import (
    "fmt"
    "os"
    "github.com/PuerkitoBio/goquery"
    "io"
    "net/http"
    "strings"
    "sync"
)


var inserts sync.WaitGroup
var downloads sync.WaitGroup

func CrawlPage(url string, ch chan string) {
	fmt.Printf("\n\nFetching Page..\n")
    resp, err := goquery.NewDocument(url)

    if err != nil {
        fmt.Printf("ERROR: Failed to crawl \"" + url + "\"\n\n")
        os.Exit(3)
    }

    // use CSS selector found with the browser inspector
    // for each, use index and item
    resp.Find("*").Each(func(index int, item *goquery.Selection) {
        linkTag := item.Find("img")
        link, _ := linkTag.Attr("src")

        if link != ""{
        	ch <- link
        }
    })

    inserts.Done()
}

func CrawlPages(urls []string, imageUrls chan string){
    // Crawl process (concurrently)
    for _, url := range urls {
        if url[:4] != "http"{
            url = "http://" + url
        }
        inserts.Add(1)
        go CrawlPage(url, imageUrls)
    }
}

// takes a channel of incoming URLs and outputs a channel of unique URLs
func EnsureUnique(in chan string, out chan string) {   
    allUrls := make(map[string]bool)

    go func() {
        for url := range in {
            if !allUrls[url] {
                allUrls[url] = true
                fmt.Printf("Enqueuing %s \n", url)
                out <- url
            }
        }
        // once in closes and the last url is pushed onto the out
        close(out)
    }()
}

func DownloadImage(url string, folder string, sem chan bool) {
	os.Mkdir(folder, os.FileMode(0777))
    defer downloads.Done()

    if url[:4] != "http"{
        url = "http:" + url
    }
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]
	file, err := os.Create(string(folder + "/" + name))
    defer file.Close()
    if err != nil {
        fmt.Printf("%v", err)
        return
    }
	resp, err := http.Get(url)
    if err != nil {
        fmt.Printf("%v", err)
        return
    }
	io.Copy(file, resp.Body)
	resp.Body.Close()

	fmt.Printf("Saving %s \n", folder + "/" + name)

    <- sem
}

func DownloadImages(in chan string, Folder string) {

    concurrency := 5
    sem := make(chan bool, concurrency)

    go func() {
        for ui := range in {
            sem <- true
            downloads.Add(1)
            go DownloadImage(ui, Folder, sem)
        }
    }()
}

func main() {
    if len(os.Args) < 3 {
    	fmt.Println("ERROR : Less Args\nCommand should be of type : imagescraper [folder to save] [websites]\n\n")
    	os.Exit(3)  	
    }
    
    Folder := os.Args[1]
    seedUrls := os.Args[2:]

    imageUrls := make(chan string)
    uniqueImgUrls := make(chan string)

    // Crawl websites and push image urls onto the imageUrls channel
    CrawlPages(seedUrls, imageUrls)

    // Ingest urls from imageUrls channel and output unique images onto uniqueImageUrls channel
    EnsureUnique(imageUrls, uniqueImgUrls)

    // Ingest urls from uniqueImageUrls channel, download and write into Folder
    DownloadImages(uniqueImgUrls, Folder)

    // inserts waitgroup is incremented by the Crawl
    inserts.Wait()
    close(imageUrls)
    
    downloads.Wait()
    fmt.Printf("\n\nScraped succesfully\n\n")

}
