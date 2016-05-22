# gocmdpev
A command-line GO Postgres query visualizer, heavily inspired by the excellent (web-based) [pev](https://github.com/AlexTatiyants/pev)

![image](https://cloud.githubusercontent.com/assets/14410/15449922/bd129a10-1f83-11e6-9480-b4c103d7c0a5.png)

## Usage

```
go get github.com/simon-engledew/gocmdpev
```

Generate a query plan with all the trimmings by prefixing your query with:

```pgsql
EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON)
```

Then pipe the resulting query plan into `gocmdpev`.

On MacOS you can just grab a query on your clipboard and run this one-liner:

```bash
pbpaste | sed 's/^/EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON) /' | psql -qAt <DATABASE> | gocmdpev
```

## Python 3 Bindings

Build:

```bash
go build -buildmode=c-shared -o pycmdpev.so
```


```python
import pycmdpev

pycmdpev.visualize("<JSON EXPLAIN STRING>")
```
