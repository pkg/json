# pkg/json [![GoDoc](https://godoc.org/github.com/pkg/json?status.svg)](https://godoc.org/github.com/pkg/json)

An alternative JSON decoder for Go.

## Features

- `json.Scanner`, when provided an external buffer, does not allocate.
- `io.Reader` friendly; you don't need to buffer your input in memory.
- `json.Decoder.Token()` is _almost_ allocation free. It allocates a heck of a lot less than `encoding/json.Decoder.Token()`

## Is it faster than fastjson/ultrajson/megajson/fujson?

Honestly, I don't know.
I have some benchmarks that show that `pkg/json` is faster than `encoding/json` for tokenisation, but this package isn't finished yet.

```
BenchmarkDecoder/pkgjson/canada.json.gz-4                    151           7678709 ns/op         293.16 MB/s         152 B/op          3 allocs/op
BenchmarkDecoder/encodingjson/canada.json.gz-4                12         102050197 ns/op          22.06 MB/s    17740358 B/op     889106 allocs/op
BenchmarkDecoder/pkgjson/citm_catalog.json.gz-4              346           3396931 ns/op         508.46 MB/s         152 B/op          3 allocs/op
BenchmarkDecoder/encodingjson/citm_catalog.json.gz-4          40          28496966 ns/op          60.61 MB/s     5665615 B/op     324799 allocs/op
BenchmarkDecoder/pkgjson/twitter.json.gz-4                   708           1676174 ns/op         376.76 MB/s         168 B/op          4 allocs/op
BenchmarkDecoder/encodingjson/twitter.json.gz-4               66          17879604 ns/op          35.32 MB/s     3660255 B/op     187815 allocs/op
BenchmarkDecoder/pkgjson/code.json.gz-4                      144           8403582 ns/op         230.91 MB/s         264 B/op          6 allocs/op
BenchmarkDecoder/encodingjson/code.json.gz-4                   9         123089228 ns/op          15.76 MB/s    23355896 B/op    1319125 allocs/op
BenchmarkDecoder/pkgjson/example.json.gz-4                 31387             37471 ns/op         347.58 MB/s         168 B/op          4 allocs/op
BenchmarkDecoder/encodingjson/example.json.gz-4             2779            415232 ns/op          31.37 MB/s       82416 B/op       4325 allocs/op
```

## Should I use this?

Right now? No.
In the future, maybe.

## I've found a bug!

Great! Please follow this two step process:

1. Raise an issue
2. Send a PR (unless I fix it first)

## Standing on the shoulders of giants

This project is heavily influenced by Steven Schveighoffer's [`iopipe`](https://www.youtube.com/watch?v=un-bZdyumog).
