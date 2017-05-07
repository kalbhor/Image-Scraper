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

var wg sync.WaitGroup // Used for waiting for channels
const MaxThreads = 30 

func Crawl(url string, ch chan string, chFinished chan bool) {

	fmt.Printf("\n\nFetching Page..\n")
    resp, err := goquery.NewDocument(url)

    defer func() {
        // Notify we're done
        chFinished <- true
        fmt.Printf("Done Crawling...")
    }()

    if err != nil {
        fmt.Printf("ERROR: Failed to crawl \"" + url + "\"\n\n")
        os.Exit(3)
    }

    resp.Find("*").Each(func(index int, item *goquery.Selection) {
        linkTag := item.Find("img")
        link, _ := linkTag.Attr("src")

        if link != ""{
        	ch <- link
        }
    })
}


func DownloadImg(Images []string) {

    os.Mkdir("tmp", os.FileMode(0522))

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
    	fmt.Printf("Saving %s \n", name)
    }
    defer wg.Done()
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
    Images := make([]string, 0)
    seedUrls := os.Args[1:]

    // Channels
    chImgs := make(chan string)
    chFinished := make(chan bool)

    // Crawl process (concurrently)
    for _, url := range seedUrls {
        go Crawl(url, chImgs, chFinished)
    }

    for c := 0; c < len(seedUrls); {
        select {
            case url := <-chImgs:
                    Images = append(Images, url)
            case <-chFinished:
                    c++
            }
    }
    close(chImgs)

    Images = SliceUniq(Images)
    fmt.Printf("\n\n========= Found %d Unique Images =========\n\n", len(Images))

    var i,j,ThreadCount int

    if len(Images)/3 > MaxThreads {
        ThreadCount = MaxThreads
    }else if ThreadCount <= 0 {
        ThreadCount = 1
        i,j = 1,0
    }else {
        ThreadCount = len(Images)/3
        i,j = len(Images)/ThreadCount,0
    }

	for ; i < len(Images) ; i += len(Images)/ThreadCount {
        wg.Add(1)
		go DownloadImg(Images[j:i])
		j = i
	}


    wg.Wait()
    fmt.Printf("\n\nScraped succesfully\n\n")

}

