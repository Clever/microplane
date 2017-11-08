# microplane

A CLI tool to make git changes across many repos, especially useful with Microservices.

![microplane](https://cdn.pixabay.com/photo/2013/07/12/14/16/lemon-148119_640.png)

_"the lemon is Github"_

## Workflow

```
Init => Clone => Plan => Push => Merge
```

## Usage

See [per-command DOCTORcs](docs/)

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
