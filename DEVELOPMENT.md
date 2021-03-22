## Development

### Build Locally

Microplane is a Golang project.

First, clone the repo.

To install dependencies run
```
make install_deps
```

To build, run
```
make build
```

You should now have a working build of Microplane in `./bin/mp`.

### Design

Microplane parallelizes various git commands and API calls.

At each step in the Microplane workflow, a repo only moves forward if the previous step for that repo was successful.

We persist the progress of a Microplane run in the following local file structure.

```
mp/
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

### Releasing

Before releasing:

- Ensure you've tested microplane flow end to end. (future: improve integration tests)
- PR your changes and get them reviewed.

To publish a release:

- Merge your approved pull request to `master`.
- Push another commit to `master`, updating both `VERSION` with the new version and `CHANGELOG.md` with a description of the changes.
- CircleCI will publish a release to GitHub.
