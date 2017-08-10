# gocmdpev
A command-line GO Postgres query visualizer, heavily inspired by the excellent (web-based) [pev](https://github.com/AlexTatiyants/pev)

![image](https://cloud.githubusercontent.com/assets/14410/15449922/bd129a10-1f83-11e6-9480-b4c103d7c0a5.png)

## Usage

```
go get github.com/simon-engledew/gocmdpev
```

or via Homebrew:

```
brew tap simon-engledew/gocmdpev
brew install gocmdpev
```

Generate a query plan with all the trimmings by prefixing your query with:

```pgsql
EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON)
```

Then pipe the resulting query plan into `gocmdpev`.

On MacOS you can just grab a query on your clipboard and run this one-liner:

```bash
pbpaste | sed '1s/^/EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON) /' | psql -qXAt <DATABASE> | gocmdpev
```

## Python 3 Bindings

Tested on python 3.4.2 built with [golang:1.8.3-jessie](https://github.com/docker-library/golang/blob/2a15dfff04accfd31c2a45b3bb7423aa86aa2d60/1.8/jessie/Dockerfile).

Build:

```bash
go build -buildmode=c-shared -o pycmdpev.so pybindings/*
```

```python
import pycmdpev

pycmdpev.visualize("<JSON EXPLAIN STRING>")
```

## Using with Ruby on Rails

Try the [`pg-eyeballs`](https://github.com/bradurani/pg-eyeballs) gem
