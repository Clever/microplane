# microplane

A CLI tool to make git changes across many repos.

Learn more about microplane in this [introductory blogpost](https://medium.com/always-a-student/mo-repos-mo-problems-how-we-make-changes-across-many-git-repositories-293ad7d418f0).

![microplane](https://cdn.pixabay.com/photo/2013/07/12/14/16/lemon-148119_640.png)

_"the lemon is Git{Hub,Lab}"_

## Setup

Here are several ways to install microplane:

- *Pre-built release* - You can download a pre-built version of Microplane from the [Github releases](https://github.com/Clever/microplane/releases).
- *Compile it yourself*  - Run `go get github.com/Clever/microplane/cmd`. In this case the binary will be installed to `$GOPATH/bin/microplane`. Alternately, you can follow the steps under "Development", below.
- *Homebrew* - `brew install microplane`. The latest homebrew formula is [here](https://github.com/Homebrew/homebrew-core/blob/master/Formula/microplane.rb)

## Usage

### GitHub setup

The `GITHUB_API_TOKEN` environment variable must be set for Github. This should be a [GitHub Token](https://github.com/settings/tokens) with `repo` scope.

Optional: If you use self-hosted Github, you can specify its URL by passing `--provider_url=<your URL>` when running `mp init`.
This URL should look like: `https://[hostname]`. Don't include path parameters like `/api/v3` or `/api/uploads`.

_Self-hosted Github setup with different URLs for the main API and uploads API are not yet supported. If this is a blocker for you, please file an issue or make a PR._

### GitLab setup

The `GITLAB_API_TOKEN` environment variable must be set for Gitlab. This should be a [GitLab access token](https://gitlab.com/profile/personal_access_tokens)

To use Gitlab, you must specifically pass `--provider=gitlab` when running `mp init`.

Optional: If you use a self-hosted Gitlab, you can specify its URL by passing `--provider_url=<your URL>` when running `mp init`.

### Using Microplane

Microplane has an opinionated workflow for how you should manage git changes across many repos.
To make a change, use the following series of commands.

1. [Init](docs/mp_init.md) - target the repos you want to change
2. [Clone](docs/mp_clone.md) - clone the repos you just targeted
3. [Plan](docs/mp_plan.md) - run a script against each of the repos and preview the diff
4. [Push](docs/mp_push.md) - commit, push, and open a Pull Request
5. [Merge](docs/mp_merge.md) - merge the PRs

For an in-depth example, check out the [introductory blogpost](https://medium.com/always-a-student/mo-repos-mo-problems-how-we-make-changes-across-many-git-repositories-293ad7d418f0).

## Related projects

- https://github.com/Skyscanner/turbolift
- https://github.com/octoherd/cli

## Development

See [`Development.md`](./DEVELOPMENT.md).
