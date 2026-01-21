# Contributing and developing

This repository contains utility packages for the Go programming language. You are welcome to use them. 

Contributions following the minimalist spirit of this project are welcome. 
On the other hand, please be aware that any feature not stemming from a Pix4D need will probably not be implemented nor a PR accepted.

**Please, before opening a PR, open a ticket to discuss your use case**.

This allows to better understand the _why_ of a new feature and not to waste your time (and ours) developing a feature that for some reason doesn't fit well with the spirit of the project or could be implemented differently.
This is in the spirit of [Talk, then code](https://dave.cheney.net/2019/02/18/talk-then-code).

We care about code quality, readability and tests, so please follow the current style and provide adequate test coverage.
In case of doubts about how to tackle testing something, feel free to ask.

## Development Prerequisites

### Required

* Go, version >= 1.25

### Optional

* [Task], version >= 3.40
* [gopass] to securely store secrets for integration tests.

## Configuration and secrets

We require environment variables to pass test configuration. The reason is twofold:

* To enable any contributor to run their own tests without having to edit any file.
* To securely store secrets!

The list of environment variables to set up is:

* `TEST_REPO_NAME`: Name of the test repository
* `TEST_REPO_OWNER`: GitHub user or organization who own the test repository
* `TEST_COMMIT_SHA`: The full SHA (40 digits) from a commit of the test repository
* `TEST_OAUTH_TOKEN`: GitHub Personal Access Token (PAT) with `repo:status` scope
* `GITHUB_APP_PRIVATE_KEY`: GitHub App's private key
* `GITHUB_APP_CLIENT_ID`: GitHub App's client ID
* `GITHUB_APP_INSTALLATION_ID`: GitHub App's installation ID

For local workflow, we use the [gopass] tool, that stores secrets in the file system using GPG. We then make the secrets available as environment variables in the [Taskfile.yml](Taskfile.yml).

For GitHub Actions workflow, we require some Actions secrets and variables be created and available in GitHub. 

The following sections contain first instructions for the setup, then instructions on how to run the tests.

### TEST_REPO_NAME

1. In your GitHub account, create a test repository, say for example [`gokit-test-read-write`](https://github.com/Pix4D/gokit-test-read-write).
2. Export the name of the test repository as environment variable:
```console
$ export TEST_REPO_NAME=gokit-test-read-write
```
3. - Go to [Settings | Variables | Actions](https://github.com/Pix4D/go-kit/settings/variables/actions)
   - Create a `New repository variable`
   - Variable name: `TEST_REPO_NAME`.
   - Variable value: the name of the test repository

### TEST_REPO_OWNER

1. Export the owner of the test repository as environment variable:
```console
$ export TEST_REPO_OWNER=pix4d
```
2. - Go to [Settings | Variables | Actions](https://github.com/Pix4D/go-kit/settings/variables/actions)
   - Create a `New repository variable`
   - Variable name: `TEST_REPO_OWNER`.
   - Variable value: the owner of the test repository

### TEST_COMMIT_SHA

1. In the test repository, push at least one commit, on any branch you fancy. Take note of the 40 digits commit SHA (the API wants the full SHA).
2. Export the commit SHA as environment variable:
```console
$ export TEST_COMMIT_SHA=751affd155db7a00d936ee6e9f483deee69c5922
```
3. - Go to [Settings | Variables | Actions](https://github.com/Pix4D/go-kit/settings/variables/actions)
   - Create a `New repository variable`
   - Variable name: `TEST_COMMIT_SHA`.
   - Variable value: the commit SHA

### TEST_OAUTH_TOKEN

1. Go to [Settings | Tokens](https://github.com/settings/tokens) for your account / organization
2. Create or regenerate a token with name `test-gokit`, with scope `repo:status`. Set an expiration of 90 days.
3. Store the token using gopass:
```console
$ gopass insert gokit/test_oauth_token
```
4. - Go to [Settings | Secrets | Actions](https://github.com/Pix4D/go-kit/settings/secrets/actions)
   - Create a `New repository secret`
   - Secret name: `TEST_OAUTH_TOKEN`.
   - Secret value: the value of the token

### GITHUB_APP_PRIVATE_KEY

1. Create and register a GitHub App for your account / organization
2. Copy the private key of the App
3. Store the key using gopass:
```console
$ gopass insert --multiline gokit/github_app_private_key
```
4. - Go to [Settings | Secrets | Actions](https://github.com/Pix4D/go-kit/settings/secrets/actions)
   - Create a `New repository secret`
   - Secret name: `GITHUB_APP_PRIVATE_KEY`.
   - Secret value: the value of the private key

### GITHUB_APP_CLIENT_ID

1. Copy the client ID of the App
2. Export the client ID as environment variable:
```console
$ export GITHUB_APP_CLIENT_ID=Iv23lir9pyQlqmweDPbz
```
3. - Go to [Settings | Variables | Actions](https://github.com/Pix4D/go-kit/settings/variables/actions)
   - Create a `New repository variable`
   - Variable name: `GITHUB_APP_CLIENT_ID`.
   - Variable value: the value of the private key

### GITHUB_APP_INSTALLATION_ID

1. Copy the installation ID of the App
2. Export the installation ID as environment variable:
```console
$ export GITHUB_APP_INSTALLATION_ID=64650729
```
3. - Go to [Settings | Variables | Actions](https://github.com/Pix4D/go-kit/settings/variables/actions)
  - Create a `New repository variable`
  - Variable name: `GITHUB_APP_INSTALLATION_ID`.
  - Variable value: the value of the private key


## Running tests

### Install test dependencies

```console
$ task install:deps
```

###  Unit tests

```console
$ task test:unit
```

### Integration tests

```console
$ task test:all
```

The integration tests have the following logic:

* If none of the environment variables are set, we skip the test.
* If all the environment variables are set, we run the test.
* If some environment variables are set and some not, we fail the test. We do this on purpose to signal to the user that the environment variables are misconfigured.

### GitHub Actions

Trigger a manual run of the CI to verify that everything works fine.

## License

This code is licensed according to the MIT license (see file [LICENSE](./LICENSE)).

[Task]: https://taskfile.dev/
[gopass]: https://github.com/gopasspw/gopass