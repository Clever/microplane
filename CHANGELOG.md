# changelog

2025-04-16 - v0.0.36

- Chore - Updates dependency libraries
- Fixes - Git clone errors are more detailed
- Adds - support for HTTPS cloning

2021-08-16 - v0.0.34

- Changes - use the default branch as the Gitlab PR target
- Changes - make the Gitlab PR body consistent with Github

2021-08-11 - v0.0.33

- Add option to set merge method (merge, squash or rebase). (Example: `mp merge --merge-method squash`)

2021-05-22 - v0.0.32

- Adds - `--diff` flag added to `mp plan` to show the diff of the changes made in each repo
- Changes - `mp plan` no longer shows the diff for a single repo by default
- Changes - diffs are now shown in color

2021-05-15 - v0.0.31

- Adds - `--draft` flag added to `mp push`, to create a draft Pull Request
- Adds - `mp version` command prints Microplane's version

2021-05-14 - v0.0.30

- Fixes - `mp sync` bug in Gitlab, due to ID vs IID

2021-05-14 - v0.0.29

- Adds - Command `mp sync` to sync local status with remote status [#81](https://github.com/Clever/microplane/pull/81)
- Changes - Command `mp status` has flag `--sync` to sync local status with remote status [#81](https://github.com/Clever/microplane/pull/81)

2021-03-22 - v0.0.28

- Adds MVP support for Github enterprise (testing pending). [#74](https://github.com/Clever/microplane/pull/74)
- Choose Github vs Gitlab by passing `--provider` flag to `mp init`. [#73](https://github.com/Clever/microplane/pull/73)

2020-08-04 - v0.0.22

- Fixes --body-file flag usage [#60](https://github.com/Clever/microplane/pull/60)

2019-01-13 - v0.0.21

- Adding an --all-repos flag to init all repos under a specific org

2019-01-10 - v0.0.20

- Adding a --repo-search flag to init using github repo search

2019-08-10 - v0.0.16

- Increase default throttle time for merge/push from 1ms => 30s. Before, it was effectively *not* throttled.

2019-08-10 - v0.0.15

- Add support for init from file

2019-08-02 - v0.0.14

- Add support for pagination in GitLab search.
