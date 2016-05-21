# gocmdpev
A command-line GO Postgres query visualizer, heavily inspired by [pev](https://github.com/AlexTatiyants/pev)

![image](https://cloud.githubusercontent.com/assets/14410/15446008/77530120-1f08-11e6-85ec-d39c586547c6.png)

## Usage

Prefix your query with a full explain:

`EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON)`

Then pipe it into `psql -qAt` to generate a query plan. Finally, pipe this into `gocmdpev`:

`pbpaste | psql -qAt | ./gocmdpev`
