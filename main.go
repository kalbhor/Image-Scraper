package main

import (
    "fmt"
    "os"
    "github.com/PuerkitoBio/goquery"
)



func crawl(url string, ch chan string, chFinished chan bool) {
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
    resp.Find("div").Each(func(index int, item *goquery.Selection) {
        linkTag := item.Find("img")
        link, _ := linkTag.Attr("src")

        ch <- link
    })
}




func main() {
    foundImages := make(map[string]bool)
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
                    foundImages[url] = true
            case <-chFinished:
                    c++
            }
    }

    for url, _ := range foundImages {
            fmt.Println(" - " + url)
    }

    close(chImgs)

}

