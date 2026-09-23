# Changelog

## 1.0.0 (2026-09-23)

First release of Unclick, the continuation of Terraformer 0.8.30.

### Added

- `unclick scan PROVIDER`: finds existing resources through the provider's list resources (Terraform 1.14's discovery mechanism, which OpenTofu cannot use yet) and needs no code per resource type. References between scanned resources are made by ID.
- Every generated resource is validated by the provider; optional arguments the provider rejects are left out, so the configuration passes `tofu plan`.
- `imports.tf` with one `import` block per resource.
- `--tf-binary` and `--provider-version` to choose the tool that installs providers and the provider release.
- Build tags for smaller binaries: `go build -tags slim,aws`.
- End-to-end tests against moto (AWS) and kind (Kubernetes).

### Changed

- New provider plugin client: protocol 5 and 6, so providers built on the plugin framework work. Terraform 0.12 is no longer embedded.
- Providers are found in OpenTofu and Terraform caches, or installed with `tofu init` / `terraform init`.
- `required_providers` always carries the provider's registry address and the release used.
- Default layout is one directory per provider (`generated/<provider>/`) with direct references between resources.
- Deprecated arguments and, for SDKv2 resources, optional numbers equal to 0 are left out of the configuration.
- AWS: the credential chain of the SDK is used unless `--profile` is given.
- GitHub: `GITHUB_TOKEN` works (it was ignored), and personal accounts can be imported, not only organizations.

### Removed

- The `terraform.tfstate` output and the `--state` / `--bucket` options: `import` blocks replace them.
- The Snap package definition.
