# snappy

<p align="center">
<img src="https://stacker.news/favicon.png" width="64" height="64" />
<img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" width="64" height="64" />
</p>

<p align="center">A Go client for the <a href="https://stacker.news" target="_blank">Stacker News</a> GraphQL API</p>

## How to use

As library:

```
$ go get github.com/ekzyis/snappy
```

```go
import sn "github.com/ekzyis/snappy"
```

As command:

```
$ go install github.com/ekzyis/snappy/cmd/snappy
$ snappy
  ___ ___  ___ ____  ___  __ __
 (_-</ _ \/ _ `/ _ \/ _ \/ // /
/___/_//_/\_,_/ .__/ .__/\_, /
             /_/  /_/   /___/

Commands:
  query    Query all items of a user.

Usage: snappy query -author <username> [-type all|posts|comments]
```

`SN_API_KEY` must be set in your environment for authenticated API access.

## How to test

1. Run SN
2. Set `TEST_SN_BASE_URL` and `TEST_SN_API_KEY` in .env
3. Run `go test ./...`
