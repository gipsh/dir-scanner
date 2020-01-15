package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var END_MESSAGE = "byebyebye"

func producer(ch chan<- string, wordlist string, workers int) {

	file, err := os.Open(wordlist)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ch <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < workers; i++ {
		ch <- END_MESSAGE
	}

	fmt.Println("producer done")
}

func consumer(baseurl string, ch <-chan string, id int, wg *sync.WaitGroup) {

	defer wg.Done()
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		elem := <-ch
		if elem == END_MESSAGE {
			break
		}
		fullurl := fmt.Sprintf("%s/%s", baseurl, elem)
		response, err := client.Get(fullurl)
		if err != nil {
			log.Fatal(err)
		}
		//if response.StatusCode != 400 {
		fmt.Printf("[%d] %d %s\n", id, response.StatusCode, fullurl)
		//}
	}

	fmt.Printf("Exit worker %d", id)

}

func main() {

	fmt.Println("--[dir-scanner]--")

	urlPtr := flag.String("url", "http://example.com:80", "url")
	wordlistPtr := flag.String("wordlist", "common.txt", "wordlist file")
	workersPtr := flag.Int("workers", 3, "workers")

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("expected at least 'url' and 'wordlist'")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("[-] using url: ", *urlPtr)
	fmt.Println("[-] wordlist: ", *wordlistPtr)
	fmt.Println("[-] workers:", *workersPtr)

	var wg sync.WaitGroup
	ch := make(chan string, 5)

	for i := 0; i < *workersPtr; i++ {
		wg.Add(1)
		go consumer(*urlPtr, ch, i, &wg)
	}

	go producer(ch, *wordlistPtr, *workersPtr)

	wg.Wait()
}
