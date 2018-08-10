autotranslate CircleCI 1.0 -> 2.0

### Jira

[INFRA-3160](https://clever.atlassian.net/browse/INFRA-3160)

### Overview

**This is an automated PR, created with `circle-v2-migrate` v0.1.0.**

If you have any questions, please ask in #circleci-1-sunset in slack.

CircleCI 1.0 is sunsetting August 31st, meaning CircleCI 1.0 builds will no longer work on September 1st.

This PR uses the output of the `circle-v2-migrate` automigration script, with `microplane`, to translate the build config to CircleCI 2.0.

This PR should not make any changes to application code.

### Reviewing

Please check the following:

- [ ] build works
- [ ] build process includes all steps from the CircleCI 1.0 config


If the build has failed, please link to the failing build in this repo's row in the [CircleCI 1.0 -> 2.0 migration tracking spreadsheet](https://docs.google.com/spreadsheets/d/1Uv6i2TXxZGBUCdjidp2xbqn3gMrgnikJnLgZBXicDBQ/edit?usp=sharing).

If the build is missing steps, please leave a note in this repo's row in the [CircleCI 1.0 -> 2.0 migration tracking spreadsheet](https://docs.google.com/spreadsheets/d/1Uv6i2TXxZGBUCdjidp2xbqn3gMrgnikJnLgZBXicDBQ/edit?usp=sharing).

### Roll Out

All test, build, publish, and deploy steps should have been preserved in the translation.

Use normal rollout practices for this repository.