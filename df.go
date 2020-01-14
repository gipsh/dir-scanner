package main

import (
        "fmt"
        "log"
        "net/http"
        "os"
	"bufio"
	"time"
	"flag"
	"sync"

)

func producer(ch chan<- string) {

    file, err := os.Open("./100.txt")
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

   fmt.Println("producer done")
}

func consumer(baseurl string, ch <-chan string,  id int,  wg *sync.WaitGroup) {

	defer wg.Done()
	client := http.Client{
	    Timeout: 3 * time.Second,
	}

	for elem := range ch {
         fullurl := fmt.Sprintf("%s/%s",baseurl,elem)
         response, err := client.Get(fullurl)
         if err != nil {
                log.Fatal(err)
         }
	if (response.StatusCode != 400) {
         fmt.Printf("[%d] %d %s\n",id, response.StatusCode, fullurl)
	}
      }


}

func fanOutUnbuffered(ch <-chan string, size int) []chan string {
    cs := make([]chan string, size)
    for i, _ := range cs {
        // The size of the channels buffer controls how far behind the recievers
        // of the fanOut channels can lag the other channels.
        cs[i] = make(chan string)
    }
    go func() {
        for i := range ch {
            for _, c := range cs {
                c <- i
            }
        }
        for _, c := range cs {
            // close all our fanOut channels when the input channel is exhausted.
            close(c)
        }
    }()
    return cs
}



func main() {

	fmt.Println("--[dir-scanner]--")

	urlPtr := flag.String("url", "http://example.com:80", "url")
	wordlistPtr := flag.String("wordlist", "common.txt", "wordlist file")
        workersPtr := flag.Int("workers", 3, "workers")

	flag.Parse()

	fmt.Println("word:", *urlPtr)
    	fmt.Println("numb:", *wordlistPtr)
    	fmt.Println("workers:", *workersPtr)

    	if len(os.Args) < 2 {
       	  fmt.Println("expected at least 'url' and 'wordlist'")
      	  os.Exit(1)
    	}
    
	var wg sync.WaitGroup
	ch := make(chan string,5)

	for i := 0; i < *workersPtr; i++ {
		wg.Add(1)
		go consumer("http://18.215.172.153:3001", ch, i, &wg)
		
	}

	go producer(ch)

	wg.Wait()
}
