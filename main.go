package main

import (
    "fmt"
    "os"
    "flag"
    "strings"
    "bufio"
    "sync"
    "context"

    "golang.org/x/sync/semaphore"
    "github.com/valyala/fasthttp"
)

func BustDir(url string, wg *sync.WaitGroup, sem *semaphore.Weighted, client *fasthttp.Client) {
    defer wg.Done()

    if err := sem.Acquire(context.Background(), 1); err != nil {
        fmt.Println("Failed to acquire semaphore:", err)
        return
    }
    defer sem.Release(1)

    req := fasthttp.AcquireRequest()
    defer fasthttp.ReleaseRequest(req)
    req.SetRequestURI(url)

    resp := fasthttp.AcquireResponse()
    defer fasthttp.ReleaseResponse(resp)

    err := client.Do(req, resp)
    if err != nil {
        fmt.Println("Error making request:", err)
        return
    }

    statusCode := resp.StatusCode()
    if statusCode < 400 || statusCode >= 600 {
        fmt.Println("Found:", url, "Status Code:", statusCode)
    }
}

func main() {
    uFlag := flag.String("u", "", "URL to fuzz")
    wFlag := flag.String("w", "", "Wordlist file path")
    flag.Parse()

    if *uFlag == "" {
        fmt.Println("Please provide a URL using the -u flag")
        os.Exit(1)
    }
    if *wFlag == "" {
        fmt.Println("Please provide a wordlist file path using the -w flag")
        os.Exit(1)
    }

    url := *uFlag
    filePath := *wFlag
    if !strings.HasSuffix(url, "/") {
        url += "/"
    }

    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening wordlist file:", err)
        os.Exit(1)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    fmt.Println("Running gofuzzer on", url)

    var wg sync.WaitGroup

    sem := semaphore.NewWeighted(100)
    client := &fasthttp.Client{
        MaxConnsPerHost: 100,
    }

    for scanner.Scan() {
        word := scanner.Text()
        wg.Add(1)
        go BustDir(url + word, &wg, sem, client)
    }

    wg.Wait()
    fmt.Println("Fuzzing completed.")
}
