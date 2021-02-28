# checker

Simple command line tool to perform universal header and path checks written in Go.

## Wordlist format

When creating the wordlist, any property from a [url.URL](https://golang.org/pkg/net/url/#URL) object in Go can be used, for example the below is an example header wordlist:

```
X-Original-URL: {{.Path}}
Referer: {{.}}
Host: {{.Host}}
```
