# rv

**rv** is a Redis data viewer application with a text-based user interface that runs in your terminal.

> Note: **rv** is currently in development.


## Features

* Repeatedly SCAN keys or key patterns
* Enable/disable scanners
* Inspect keys matching SCAN configurations
* Inspect data structures (single key-value pairs, lists, sets, sorted sets and hashes)


## Usage

1. `go build` from project root
1. Create a `config.toml` (see [example config](#example-minimum-config))
1. `./rv` or `rv.exe` (optionally pass a config file argument, by default `config.toml` will be used)


## Configuration

Besides setting up the connection to your Redis server, you can define an infinite number of *scanners*. Scanners are
small workers in the background which repeatedly [SCAN](https://redis.io/commands/scan) their defined pattern and report
how many matching keys they found. You can then inspect these keys in detail.

First, you have to name your scanner. In the scanner list, scanners are ordered by name.

```toml
[scans.my_scanner]
```

Or you can use spaces in the name:

```toml
[scans."My Scanner"]
```

Next, define the pattern you want to scan. The one below will scan a single key since there are no wildcards characters
(`*`) in the pattern:

```toml
pattern = "example:key"
```

To scan a range of keys, pattern could be something like:

```toml
pattern = "example:*"
```

Your patterns must match a single type of keys. Also, you have to tell the Redis type to **rv**:

```toml
type = "hash"
```

Finally, set the frequency of the scan:

```toml
interval = "20s"
```

To sum up, your scanner config now looks like this:

```toml
[scans."My Scanner"]
pattern = "example:*"
type = "hash"
interval = "20s"
```

This scanner will kick off a SCAN command in every 20 seconds and will look for *hashes* matching the `example:*`
pattern.


#### Example minimum config

```toml
[redis]
server = "localhost:6379"

[scans.customers]
pattern = "customer:*"
type = "hash"
interval = "15s"
```
