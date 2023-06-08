# BuildPulse Test Reporter [![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/buildpulse/test-reporter/master/LICENSE)

The BuildPulse test reporter is a binary that connects your continuous integration (CI) to [buildpulse.io][] to help you detect, track, and eliminate flaky tests.

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
