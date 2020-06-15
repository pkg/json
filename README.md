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
BenchmarkDecoderToken/pkgjson/canada.json.gz-16                      224           5246921 ns/op         429.02 MB/s         152 B/op          3 allocs/op
BenchmarkDecoderToken/encodingjson/canada.json.gz-16                  16          68592743 ns/op          32.82 MB/s    17740647 B/op     889107 allocs/op
BenchmarkDecoderToken/pkgjson/citm_catalog.json.gz-16                514           2309466 ns/op         747.88 MB/s         152 B/op          3 allocs/op
BenchmarkDecoderToken/encodingjson/citm_catalog.json.gz-16            61          19922140 ns/op          86.70 MB/s     5665622 B/op     324799 allocs/op
BenchmarkDecoderToken/pkgjson/twitter.json.gz-16                    1046           1121534 ns/op         563.08 MB/s         168 B/op          4 allocs/op
BenchmarkDecoderToken/encodingjson/twitter.json.gz-16                 96          12734433 ns/op          49.59 MB/s     3660290 B/op     187815 allocs/op
BenchmarkDecoderToken/pkgjson/code.json.gz-16                        216           5540084 ns/op         350.26 MB/s         264 B/op          6 allocs/op
BenchmarkDecoderToken/encodingjson/code.json.gz-16                    14          78507285 ns/op          24.72 MB/s    23355997 B/op    1319126 allocs/op
BenchmarkDecoderToken/pkgjson/example.json.gz-16                   51165             23740 ns/op         548.62 MB/s         168 B/op          4 allocs/op
BenchmarkDecoderToken/encodingjson/example.json.gz-16               4282            271005 ns/op          48.06 MB/s       82416 B/op       4325 allocs/op
BenchmarkDecoderToken/pkgjson/sample.json.gz-16                     2172            544453 ns/op        1262.72 MB/s        1160 B/op          9 allocs/op
BenchmarkDecoderToken/encodingjson/sample.json.gz-16                 331           3680055 ns/op         186.82 MB/s      759686 B/op      26643 allocs/op
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
