package main

import (
    "fmt"
    "os"
    "github.com/PuerkitoBio/goquery"
    "io"
    "net/http"
    "strings"
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


func downloadImg(Images []string, ch chan bool) {

    for _, url := range Images {
    	parts := strings.Split(url, "/")
		name := parts[len(parts)-1]
		file, _ := os.Create("tmp/" + name)
		resp, _ := http.Get(url)
		io.Copy(file, resp.Body)
		file.Close()
		resp.Body.Close()
    	fmt.Println("====Saving==== " + name)
    }
    ch <- true
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
    
    l := 0
    ch := make(chan bool)
	for i:=len(Images)/4; i < len(Images); i+= len(Images)/20 {
		go downloadImg(Images[l:i], ch)
		l = i
	}


    select {
    case <- ch:
    	fmt.Println("\n\n[ ---- Done! ---- ]")
    }
    close(chImgs)
    close(ch)

}

