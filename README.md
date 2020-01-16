# http-dir-scanner

Scans a web server for hidden directories and files.



## Features

You can customize the workers for speed with `-workers`.

You can route the requests throught TOR proxy with `-tor`

The User-Agent changes for each request, you can disable this behaviour with `-uarr=false`

Wordlists were stolen from OWASP 

## Build 

first get the deps

```bash
go get
```


the build 

```bash
go build
```

## Run


```bash
./dir-finder -url http://example.com:80 -wordlist wordlists/ror.txt -workers 8
```


