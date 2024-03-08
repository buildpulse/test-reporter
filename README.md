# BuildPulse Test Reporter [![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/buildpulse/test-reporter/master/LICENSE)

The BuildPulse test reporter is a binary that connects your continuous integration (CI) to [buildpulse.io][] to help you find and [fix flaky tests](https://buildpulse.io/products/flaky-tests).

Get started at [buildpulse.io][].

## Setup

Install goreleaser to build
```
brew install goreleaser/tap/goreleaser upx
```

## Build
The following will build the binary.
```
./script/build-snapshot
```

The binary can be found in `./dist`. The following platforms + architectures are supported:

- Darwin / Mac OS (amd64, arm64)
- Windows (amd64)
- Linux (amd64, arm64)
  - Ubuntu
  - Debian
  - Fedora
  - CentOS
  - RedHat
  - Alpine

## Natively Supported CI Providers
We are able to infer the required environment variables from the following CI providers:

  - Github Actions
  - BuildKit
  - CircleCI
  - Github Actions
  - Jenkins
  - Semaphore
  - Travis CI
  - Webapp.io
  - AWS CodeBuild
  - BitBucket Pipelines
  - Azure DevOps Pipelines

## Other CI Providers / Standalone Usage
To use `test-reporter` with another CI provider, the following environment variables must be set:

| Environment Variable | Description                                                        |
|----------------------|--------------------------------------------------------------------|
| `GIT_COMMIT`         | Git commit SHA                                                     |
| `GIT_BRANCH`         | Git branch of the build, or PR number                              |
| `BUILD_URL`          | URL of the build. If running locally, set as `https://example.com` |
| `ORGANIZATION_NAME`  | Name of the Github organization                                    |
| `REPOSITORY_NAME`    | Name of the repository                                             |

The following are flags that can be set. Make sure to **set flags after CLI args**.
| Flag                 | Required                          | Description                                     |
|----------------------|-----------------------------------|-------------------------------------------------|
| `account-id`         |   ✓                               | BuildPulse account ID (see dashboard)           |
| `repository-id`      |   ✓                               | BuildPulse repository ID (see dashboard)        |
| `repository-dir`     | Only if `tree` not set            | Path to repository directory                    |
| `tree`               | Only if `repository-dir` not set  | Git tree SHA                                    |
| `coverage-files`     | Only if using BuildPulse Coverage | **Space-separated** paths to coverage files.    |
| `tags`               |                                   | **Space-separated** tags to apply to the build. |
| `quota-id`           |                                   | ID of the quota to apply upload to. Quotas can be set from the BuildPulse Dashboard. |

Example:
```
BUILDPULSE_ACCESS_KEY_ID=$INPUT_KEY \
BUILDPULSE_SECRET_ACCESS_KEY=$INPUT_SECRET \
GIT_COMMIT=$GIT_COMMIT \
GIT_BRANCH=$GIT_BRANCH \
BUILD_URL=$BUILD_URL \
ORGANIZATION_NAME=$ORGANIZATION_NAME \
REPOSITORY_NAME=$REPOSITORY_NAME \
./buildpulse-test-reporter submit $REPORT_PATH --account-id $ACCOUNT_ID --repository-id $REPOSITORY_ID --repository-dir $REPOSITORY_PATH
```

[buildpulse.io]: https://buildpulse.io?utm_source=github.com&utm_campaign=tool-repositories&utm_content=test-reporter-text-link
