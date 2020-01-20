package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

var END_MESSAGE = "byebyebye"
var TOR_PROXY = "socks5://127.0.0.1:9050/"
var UA_FILE = "user_agents.txt"

type Result struct {
	Code int
	Word string
}

type ScannerConfig struct {
	Url      *string
	Wordlist *string
	Workers  *int
	Tor      *bool
	Uarr     *bool
	Insecure *bool
	Fasthttp *bool
}

func producer(ch chan<- string, config ScannerConfig) {

	file, err := os.Open(*config.Wordlist)
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

	for i := 0; i < *config.Workers; i++ {
		ch <- END_MESSAGE
	}

	// fmt.Println("producer done")
}

func consumer(ch <-chan string, rc chan<- Result, id int, config ScannerConfig, wg *sync.WaitGroup) {

	defer wg.Done()

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		elem := <-ch
		if elem == END_MESSAGE {
			break
		}

		fullurl := fmt.Sprintf("%s/%s", *config.Url, elem)

		req, err := http.NewRequest("GET", fullurl, nil)
		if err != nil {
			log.Println(err)
			continue

		}

		if *config.Uarr {
			req.Header.Set("User-Agent", GetUAManager().GetUserAgent())
		}

		response, err := client.Do(req)

		if err != nil {
			log.Println(err)
			continue
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

	sc := ScannerConfig{}

	sc.Url = flag.String("url", "http://example.com:80", "url")
	sc.Wordlist = flag.String("wordlist", "common.txt", "wordlist file")
	sc.Workers = flag.Int("workers", 8, "workers")
	sc.Tor = flag.Bool("tor", false, "use tor proxy on 127.0.0.1:9050")
	sc.Uarr = flag.Bool("uarr", true, "User-Agent round robin")
	sc.Insecure = flag.Bool("insecure", true, "TLS no certificate validation")
	sc.Fasthttp = flag.Bool("fasthttp", true, "User fasthttp library (they claim 10x faster)")

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("expected at least 'url' and 'wordlist'")
		flag.Usage()
		os.Exit(1)
	}

	file, _ := os.Open(*sc.Wordlist)
	lines, _ := lineCounter(file)

	fmt.Println("[-] using url: ", *sc.Url)
	fmt.Printf("[-] wordlist: %s [%d words]\n", *sc.Wordlist, lines)
	fmt.Println("[-] workers:", *sc.Workers)
	if *sc.Tor {
		fmt.Println("[-] using tor ", TOR_PROXY)
		os.Setenv("HTTP_PROXY", TOR_PROXY)
	}
	if *sc.Uarr {
		fmt.Println("[-] User-Agent round robin enabled")
		GetUAManager()
	}
	if *sc.Insecure {
		fmt.Println("[-] TLS certificate validation is disabled")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if *sc.Fasthttp {
		fmt.Println("[-] using fasthttp library instead of net/http")
	}

	var wg sync.WaitGroup
	ch := make(chan string, 5)
	rc := make(chan Result, 5)

	go resultProcessor(rc, lines)

	for i := 0; i < *sc.Workers; i++ {
		wg.Add(1)
		if *sc.Fasthttp {
			go fastconsumer(ch, rc, i, sc, &wg)
		} else {
			go consumer(ch, rc, i, sc, &wg)
		}
	}

	go producer(ch, sc)

	wg.Wait()
}
