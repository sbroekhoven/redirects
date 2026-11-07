# redirects

Follow the redirects of a URL with Go and see every hop: status code,
protocol, TLS version and whether HTTPS was downgraded to HTTP.

## Library

```bash
go get github.com/sbroekhoven/redirects
```

```go
data := redirects.Get("https://example.com/short-link")
if data.Error {
	log.Fatal(data.ErrorMessage)
}
for _, hop := range data.Redirects {
	fmt.Println(hop.Number, hop.StatusCode, hop.URL)
}
```

When the URL comes from untrusted users, for example in a web service, block
requests to internal addresses such as `localhost`, private networks and cloud
metadata endpoints:

```go
data := redirects.GetWithOptions(url, redirects.Options{BlockPrivateIPs: true})
```

## CLI

```bash
go install github.com/sbroekhoven/redirects/cmd/redirects@latest
redirects https://example.com/short-link
redirects -json https://example.com/short-link
redirects -markdown https://example.com/short-link
```
