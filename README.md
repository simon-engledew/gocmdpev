# gocmdpev
A command-line GO Postgres query visualizer, heavily inspired by [pev](https://github.com/AlexTatiyants/pev)

![image](https://cloud.githubusercontent.com/assets/14410/15448062/9cd5df36-1f4d-11e6-83a0-6489b905d3b7.png)

## Usage

Prefix your query with a full explain:

`EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON)`

Then pipe it into `psql -qAt` to generate a query plan. Finally, pipe this into `gocmdpev`:

`pbpaste | psql -qAt | ./gocmdpev`
