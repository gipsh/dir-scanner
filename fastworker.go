package main

import (
	"fmt"
	"sync"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

func fastconsumer(ch <-chan string, rc chan<- Result, id int, config ScannerConfig, wg *sync.WaitGroup) {

	defer wg.Done()

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	var c *fasthttp.Client

	if *config.Tor {
		c = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpSocksDialer(TOR_PROXY),
		}
	} else {
		c = &fasthttp.Client{}

	}

	for {
		elem := <-ch
		if elem == END_MESSAGE {
			break
		}

		fullurl := fmt.Sprintf("%s/%s", *config.Url, elem)

		req.Header.SetMethod(fasthttp.MethodGet)
		req.SetRequestURI(fullurl)
		//req.SetBodyString("test")

		if *config.Uarr {
			req.Header.Set("User-Agent", GetUAManager().GetUserAgent())
		}

		err := c.Do(req, res)
		if err != nil {
			fmt.Println(err)
		}
		if len(res.Body()) == 0 {
			fmt.Println("missing request body")
		}

		rc <- Result{res.StatusCode(), fullurl}

	}

	//	fmt.Printf("Exit worker %d", id)

}
