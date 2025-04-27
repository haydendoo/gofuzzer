package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
	"bufio"
	"sync"

	"github.com/valyala/fasthttp"
)

func BustDir(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
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
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "#") {
			continue
		}
		wg.Add(1)
		go BustDir(url + word, &wg)
	}

	wg.Wait()
	fmt.Println("Fuzzing completed.")
}