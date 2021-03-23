# changelog

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
