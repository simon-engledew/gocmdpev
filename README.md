# gocmdpev
A command-line GO Postgres query visualizer, heavily inspired by [pev](https://github.com/AlexTatiyants/pev)

![image](https://cloud.githubusercontent.com/assets/14410/15449922/bd129a10-1f83-11e6-9480-b4c103d7c0a5.png)

## Usage

Prefix your query with a full explain:

`EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON)`

Then pipe it into `psql -qAt` to generate a query plan. Finally, pipe this into `gocmdpev`:

`pbpaste | psql -qAt | ./gocmdpev`
