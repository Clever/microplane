# microplane

A CLI tool to make git changes across many repos, especially useful with Microservices.

## Workflow

```
Init => Clone => Plan => Push => Merge
```

## Usage

### Init

```
$ mp init <GH Search URL String>
target-foo-bar
```

### Clone

```
$ mp clone <target>
```

### Plan

Previews a change. Does not push commits to GitHub.

```
$ mp plan <target> -c <command_to_run> -m <message>
```

Should output the repos that errored.

### Push

Pushes the change you planned and opens pull requests.

```
$ mp push <target> [--delay <duration>]
```

### Merge

Merges the PRs that you pushed.

```
$ mp merge <target> [--delay <duration>]
```


### Status

View the status of a change.

```
$ mp status <target>
```

## Data Model

We persist the progress of a Microplane run in the following local file structure.

```
./mp/<target>/
  init.json
  repo1/
    clone/
      clone.json
      <git-repo>
    plan/
      plan.json
      <git-repo-with-commit>
    push/
      push.json
    merge/
      merge.json
  repo2/
    ...
```
