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

func Crawl(url string, ch chan string, chFinished chan bool) {
	fmt.Printf("\n\nFetching Page..\n")
    resp, err := goquery.NewDocument(url)

    defer func() {
        // Notify we're done
        chFinished <- true
    }()

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

    fmt.Printf("Done Crawling...")
}


func DownloadImg(Images []string, folder string) {

	os.Mkdir(folder, os.FileMode(0777))

    for _, url := range Images {
        if url[:4] != "http"{
            url = "http:" + url
        }
    	parts := strings.Split(url, "/")
		name := parts[len(parts)-1]
		file, _ := os.Create(string(folder + "/" + name))
		resp, _ := http.Get(url)
		io.Copy(file, resp.Body)
		file.Close()
		resp.Body.Close()
    	fmt.Printf("Saving %s \n", folder + "/" + name)
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
    
    if len(os.Args) < 3 {
    	fmt.Println("ERROR : Less Args\nCommand should be of type : imagescraper [folder to save] [websites]\n\n")
    	os.Exit(3)  	
    }
    
    Images := make([]string, 0)
    Folder := os.Args[1]
    seedUrls := os.Args[2:]

    // Channels
    chImgs := make(chan string)
    chFinished := make(chan bool)

    // Crawl process (concurrently)
    for _, url := range seedUrls {
        if url[:4] != "http"{
            url = "http://" + url
        }
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
    pool := len(Images)/3
    if pool > 30 {
        pool = 30
    }
    l := 0
	for i:=len(Images)/pool; i < len(Images); i += len(Images)/pool {
        wg.Add(1)
		go DownloadImg(Images[l:i], Folder)
		l = i
	}


    wg.Wait()
    fmt.Printf("\n\nScraped succesfully\n\n")

}
