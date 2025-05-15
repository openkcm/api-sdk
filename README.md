# OpenKCM API SDK 

Have all service definitions and code generated stubs for OpenKCM APIs.

## Generate protobuf code

1. **Install `buf`:** If you haven't already, install `buf` by following the instructions on
   the [Buf website](https://buf.build/docs/installation/).
   You can install buf on macOS or Linux using Homebrew:

```sh
   brew install bufbuild/buf/buf
```

2. **Generate code:** Use the ```sh make generate``` command to run the entire process.

```sh 
   make generate
```

---

## Using a specific branch

Ensure that the GOPRIVATE environment variable is set up:

   ```sh
     go env -w GOPRIVATE="github.com/*"  
   ```

To use a specific branch of a repository in your Go project, you can use the `go get` command with the branch name.
Here's an example:

> go get github.com/openkcm/api-sdk/proto/kms/api/cmk/registry/system/v1@<branch-name>

To switch back to the default branch, follow these steps:

1. Update your dependencies to the latest version on the default branch:

> go get github.com/openkcm/api-sdk/proto/kms/api/cmk/registry/system/v1@latest

2. Clean up:

```sh
  go mod tidy
```

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/openkcm/api-sdk/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Security / Disclosure
If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/openkcm/api-sdk/security/policy) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/openkcm/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright (20xx-)20xx SAP SE or an SAP affiliate company and <your-project> contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/openkcm/api-sdk).
