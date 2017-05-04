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



func crawl(url string, ch chan string, chFinished chan bool) {
	fmt.Println("\n\n============ Fetching Page ============\n\n")
    resp, err := goquery.NewDocument(url)

    defer func() {
        // Notify we're done
        chFinished <- true
    }()

    if err != nil {
        fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
        return
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
}

var wg sync.WaitGroup
func downloadImg(Images []string) {

    for _, url := range Images {
        if url[:4] != "http"{
            url = "http:" + url
        }
    	parts := strings.Split(url, "/")
		name := parts[len(parts)-1]
		file, _ := os.Create("tmp/" + name)
		resp, _ := http.Get(url)
		io.Copy(file, resp.Body)
		file.Close()
		resp.Body.Close()
    	fmt.Println("====Saving==== " + name)
    }
    defer wg.Done()
}


func main() {
    Images := make([]string, 0)
    seedUrls := os.Args[1:]

    // Channels
    chImgs := make(chan string)
    chFinished := make(chan bool)

    // Crawl process (concurrently)
    for _, url := range seedUrls {
        go crawl(url, chImgs, chFinished)
    }

    for c := 0; c < len(seedUrls); {
        select {
            case url := <-chImgs:
                    Images = append(Images, url)
            case <-chFinished:
                    c++
            }
    }
    pool := len(Images)/5
    if pool > 30 {
        pool = 30
    }
    l := 0
	for i:=len(Images)/pool; i < len(Images); i += len(Images)/pool {
        wg.Add(1)
		go downloadImg(Images[l:i])
		l = i
	}


    wg.Wait()

    fmt.Println("\n\n[ ---- Done! ---- ]")

    close(chImgs)

}

