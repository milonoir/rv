# rv

**rv** is a Redis data viewer application with a text-based user interface that runs in your terminal.

> Note: **rv** is currently in development.


## Features

* Repeatedly SCAN keys or key patterns
* Inspect keys matching SCAN configurations
* Inspect data structures (single key-value pairs, lists, sets, sorted sets and hashes)


## Usage

1. Create a `config.toml` (see example config below)
1. `go build` from project root
1. `./rv` or `rv.exe` (optionally pass a config file argument, by default `config.toml` will be used)


## Configuration


#### Example config

```toml
[redis]
server = "localhost:6379"

[scans.customers]
pattern = "customer:*"
type = "hash"
interval = "15s"
```
