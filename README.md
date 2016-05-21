# gocmdpev
A command-line GO Postgres query visualizer, heavily inspired by [pev](https://github.com/AlexTatiyants/pev)

![image](https://cloud.githubusercontent.com/assets/14410/15449790/d5531f22-1f7f-11e6-9020-7933b2a4c63b.png)

## Usage

Prefix your query with a full explain:

`EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON)`

Then pipe it into `psql -qAt` to generate a query plan. Finally, pipe this into `gocmdpev`:

`pbpaste | psql -qAt | ./gocmdpev`
