# logcounter

Stage 1 (baseline) of an educational project: count web-server requests per IP
address from CSV logs.

This is deliberately the **naive** implementation:

- single-threaded, sequential file traversal;
- all results accumulated in a single `map[string]int` held in RAM;
- no streaming, no sharding, no persistence during the run.

It exists as a starting point for a later refactor.

## Input format

```
IP,timestamp,status_code,path
192.168.1.10,2024-03-15T10:23:11Z,200,/api/users
```

- `timestamp` is RFC 3339 (`time.RFC3339`).
- Empty lines are skipped.
- Malformed lines abort the run with a file/line-prefixed error.

## Layout

```
.
├── go.mod
├── cmd/logcounter/main.go        # CLI entry point
├── internal/parser/parser.go     # line -> Record
├── internal/reader/reader.go     # sequential file reading
├── internal/counter/counter.go   # map[string]int counter
├── testdata/sample.log
└── README.md
```

## Build and run

```sh
go build ./...

go run ./cmd/logcounter -input testdata/sample.log -out counts.json -verbose
```

Using a directory (all `*.log` files) or a glob:

```sh
go run ./cmd/logcounter -input ./logs -out counts.json
go run ./cmd/logcounter -input './logs/*.log' -out counts.json
```

Output `counts.json`:

```json
{
  "10.0.0.5": 3,
  "172.16.0.42": 2,
  "192.168.1.10": 3
}
```

Progress is written to stderr via the standard logger after every file, and
every 100000 lines in `-verbose` mode:

```
progress: files=1 lines=8 unique_ips=3
done: files=1 lines=8 unique_ips=3 out=counts.json
```

## Known Limitations

On large files (tens of GB) the application crashes with OOM because all unique
IPs are held in a single map in RAM.

Additional known limitations of this stage:

- throughput is limited to a single CPU core;
- no checkpointing: a crash loses all progress;
- the whole result set must fit in memory before it can be written to JSON.
