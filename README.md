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
BenchmarkDecoder/pkgjson/canada.json.gz-4                    138           8547348 ns/op         263.36 MB/s         152 B/op          3 allocs/op
BenchmarkDecoder/encodingjson/canada.json.gz-4                12         100103279 ns/op          22.49 MB/s    17740353 B/op     889106 allocs/op
BenchmarkDecoder/pkgjson/citm_catalog.json.gz-4              313           3805954 ns/op         453.82 MB/s         152 B/op          3 allocs/op
BenchmarkDecoder/encodingjson/citm_catalog.json.gz-4          40          28544015 ns/op          60.51 MB/s     5665614 B/op     324799 allocs/op
BenchmarkDecoder/pkgjson/twitter.json.gz-4                   624           1888557 ns/op         334.39 MB/s         168 B/op          4 allocs/op
BenchmarkDecoder/encodingjson/twitter.json.gz-4               66          17803324 ns/op          35.47 MB/s     3660282 B/op     187815 allocs/op
BenchmarkDecoder/pkgjson/code.json.gz-4                      126           9443142 ns/op         205.49 MB/s         264 B/op          6 allocs/op
BenchmarkDecoder/encodingjson/code.json.gz-4                   9         112034109 ns/op          17.32 MB/s    23355939 B/op    1319125 allocs/op
BenchmarkDecoder/pkgjson/example.json.gz-4                 29368             40249 ns/op         323.58 MB/s         168 B/op          4 allocs/op
BenchmarkDecoder/encodingjson/example.json.gz-4             3010            396151 ns/op          32.88 MB/s       82416 B/op       4325 allocs/op
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
