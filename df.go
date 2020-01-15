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
	"io"
	"bytes"
	"github.com/cheggaaa/pb/v3"
)

var END_MESSAGE = "byebyebye"
var TOR_PROXY   = "socks5://127.0.0.1:9050/"

type Result struct {
	Code int
	Word string
}

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

	// fmt.Println("producer done")
}

func consumer(baseurl string, ch <-chan string, rc chan<- Result, id int, wg *sync.WaitGroup) {

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
		rc <- Result{response.StatusCode, fullurl}
	}

//	fmt.Printf("Exit worker %d", id)

}

func resultProcessor(rc <-chan Result, words int) {
	bar := pb.StartNew(words)

	for {
		result := <-rc
		bar.Increment()
		if result.Code == 200 {
		  fmt.Printf("[%d] %s \n", result.Code, result.Word)
		}
	}
	bar.Finish()
}

func lineCounter(r io.Reader) (int, error) {
    buf := make([]byte, 32*1024)
    count := 0
    lineSep := []byte{'\n'}

    for {
        c, err := r.Read(buf)
        count += bytes.Count(buf[:c], lineSep)

        switch {
        case err == io.EOF:
            return count, nil

        case err != nil:
            return count, err
        }
    }
}

func main() {

	fmt.Println("--[dir-scanner]--\n")

	urlPtr := flag.String("url", "http://example.com:80", "url")
	wordlistPtr := flag.String("wordlist", "common.txt", "wordlist file")
	workersPtr := flag.Int("workers", 8, "workers")
	torPtr := flag.Bool("tor",false,"use tor proxy on 127.0.0.1:9050")

	flag.Parse()


	if len(os.Args) < 2 {
		fmt.Println("expected at least 'url' and 'wordlist'")
		flag.Usage()
		os.Exit(1)
	}

	file, _ := os.Open(*wordlistPtr)
	lines, _ := lineCounter(file)


	fmt.Println("[-] using url: ", *urlPtr)
	fmt.Printf("[-] wordlist: %s [%d words]\n", *wordlistPtr, lines)
	fmt.Println("[-] workers:", *workersPtr)
	if (*torPtr)  {
	  fmt.Println("[-] using tor ", TOR_PROXY)
	  os.Setenv("HTTP_PROXY", TOR_PROXY)
	}

	var wg sync.WaitGroup
	ch := make(chan string, 5)
	rc := make(chan Result, 5)

	go resultProcessor(rc, lines)

	for i := 0; i < *workersPtr; i++ {
		wg.Add(1)
		go consumer(*urlPtr, ch, rc, i, &wg)
	}

	go producer(ch, *wordlistPtr, *workersPtr)

	wg.Wait()
}
