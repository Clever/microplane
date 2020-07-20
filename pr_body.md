## Link to JIRA
[SEC-870](https://clever.atlassian.net/browse/SEC-870)

## Overview
Running express server locally without specifying host will run it on `0.0.0.0`. This means locally running services are accessible by anyone on the network. `0.00.0` is expect during producation, however, the server should be bound to  `127.0.0.1` during local run.

## Testing
- make test
- make build

## Rollout
