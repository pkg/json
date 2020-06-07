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
BenchmarkDecoder/pkgjson/canada.json.gz-4                           100          11209566 ns/op         200.82 MB/s         152 B/op          3 allocs/op
BenchmarkDecoder/encodingjson/canada.json.gz-4                       12         101334610 ns/op          22.21 MB/s    17740374 B/op     889106 allocs/op
BenchmarkDecoder/pkgjson/citm_catalog.json.gz-4                     286           4799249 ns/op         359.89 MB/s         152 B/op          3 allocs/op
BenchmarkDecoder/encodingjson/citm_catalog.json.gz-4                 43          32148861 ns/op          53.73 MB/s     5665607 B/op     324799 allocs/op
BenchmarkDecoder/pkgjson/twitter.json.gz-4                          508           2083698 ns/op         303.07 MB/s         168 B/op          4 allocs/op
BenchmarkDecoder/encodingjson/twitter.json.gz-4                      66          17502857 ns/op          36.08 MB/s     3660253 B/op     187815 allocs/op
BenchmarkDecoder/pkgjson/code.json.gz-4                             100          10378053 ns/op         186.98 MB/s         264 B/op          6 allocs/op
BenchmarkDecoder/encodingjson/code.json.gz-4                         10         113882319 ns/op          17.04 MB/s    23355952 B/op    1319126 allocs/op
BenchmarkDecoder/pkgjson/example.json.gz-4                        27150             43989 ns/op         296.07 MB/s         168 B/op          4 allocs/op
BenchmarkDecoder/encodingjson/example.json.gz-4                    3105            383158 ns/op          33.99 MB/s       82416 B/op       4325 allocs/op
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
