# http-dir-scanner

Scans a web server for hidden directories and files.

You can customize the workers for speed.


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


