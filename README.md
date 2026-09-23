<img src="docs/logo.svg" width="96" align="right" alt="Unclick logo">

# Unclick

**Unclick your cloud: turn ClickOps into OpenTofu / Terraform code.**

[Русская версия](README.ru.md)

Unclick finds the resources that already exist in your cloud accounts and SaaS tools and writes them out as infrastructure as code, with `import` blocks, so that one `tofu plan` adopts them:

```
$ unclick scan aws --config region=eu-west-1
$ cd generated/aws && tofu init && tofu plan
...
Plan: 12 to import, 0 to add, 0 to change, 0 to destroy.
```

Unclick continues [Terraformer](https://github.com/GoogleCloudPlatform/terraformer), which was archived on March 16, 2026. Its 44 importers are still here, on a new core that works with current providers, with OpenTofu and with Terraform.

## Why Unclick

- **Works with current providers.** Terraformer embedded Terraform 0.12 and only spoke plugin protocol 5, so providers built on the plugin framework (Cloudflare v5 and others) failed with "Incompatible API version". Unclick speaks protocols 5 and 6.
- **Works with OpenTofu.** Providers installed by OpenTofu are found, and missing ones are installed with `tofu init`. Terraform works too.
- **Configuration you can plan.** Every generated resource is validated by the provider itself before it is written, and arguments the provider rejects are left out. The end-to-end tests expect `tofu plan` to report imports only, with no changes.
- **Import blocks instead of state files.** Terraformer wrote a version 3 state file, which OpenTofu and Terraform upgrade with the wrong provider address for most providers not published by HashiCorp. Unclick writes `imports.tf`; nothing touches your state until you apply.
- **`scan` needs no per-resource code.** It asks the provider to list what exists, the mechanism behind Terraform 1.14's `query` command, which OpenTofu does not have yet.

## Install

Download a binary for your platform from [Releases](https://github.com/Perruer/unclick/releases), or build it with Go 1.24 or newer:

```
go install github.com/Perruer/unclick@latest
```

Unclick needs `tofu` (or `terraform`) in `PATH`: it uses it to download providers, exactly as `tofu init` would, with your mirrors and credentials.

## Two ways to find resources

| | `unclick scan PROVIDER` | `unclick import PROVIDER` |
|---|---|---|
| How it finds resources | Asks the provider (list resources) | Terraformer's importers call the cloud APIs |
| Providers | Any provider with list resources: AWS (241 types in 6.66), Google, AzureRM | 44 providers, listed below |
| New resource types | Arrive with provider releases | Need a new importer |
| References between resources | By ID (`vpc_id = aws_vpc.main.id`) | From the importers' connection tables |

### scan

```
unclick scan aws --config region=eu-west-1
unclick scan aws --config region=eu-west-1 --types aws_vpc,aws_subnet,aws_security_group
unclick scan hashicorp/google --config project=my-project --config region=europe-west1
```

`--config key=value` sets provider arguments; credentials come from the environment, as for OpenTofu. Without `--types`, every type the provider can list is scanned.

### import

```
unclick import aws --regions=eu-west-1 --resources=vpc,subnet,sg
unclick import github --owner=my-org --resources=repositories
unclick import kubernetes --resources=deployments,services
unclick import aws list        # resources an importer supports
```

See [docs/](docs) for every provider's options and resources. Filtering works as in Terraformer; see [docs/upstream-README.md](docs/upstream-README.md#filtering).

### Output

Files go to `generated/<provider>/` by default:

- `provider.tf`: `required_providers` with the right source and the provider release that was used, and the provider block.
- One `.tf` file per resource type.
- `imports.tf`: one `import` block per resource.

Run `tofu init && tofu plan` there. When the plan shows only imports, `tofu apply` adopts the resources without changing them.

## Providers

`import` supports: AliCloud, Auth0, AWS, Azure, Azure AD, Azure DevOps, Cloudflare, commercetools, Datadog, DigitalOcean, Equinix Metal, Fastly, GitHub, GitLab, Gmail filters, Google Cloud, Grafana, Heroku, Honeycomb, IBM Cloud, IONOS Cloud, Keycloak, Kubernetes, LaunchDarkly, Linode, Logz.io, Mackerel, MikroTik, Myra Security, New Relic, NS1, Octopus Deploy, Okta, Opal, OpenStack, Opsgenie, PagerDuty, PAN-OS, RabbitMQ, Tencent Cloud, Vault, Vultr, Xen Orchestra, Yandex Cloud.

Tested end to end: AWS against [moto](https://github.com/getmoto/moto) (both `import` and `scan`) and GitHub against a real account; CI also runs a Kubernetes test against [kind](https://kind.sigs.k8s.io). The other importers compile and run on the new core but have not been re-tested yet; reports are welcome.

## Coming from Terraformer

| Terraformer | Unclick |
|---|---|
| `terraformer import aws ...` | `unclick import aws ...` (same importers, same flags) |
| `terraform.tfstate` next to the code | `imports.tf`; run `tofu plan` / `apply` |
| `--state=bucket`, `--bucket` | Removed: there is no state to upload |
| One directory per service | One directory per provider; `--path-pattern "{output}/{provider}/{service}/"` restores the old layout |
| Providers under `~/.terraform.d/plugins` | Found in OpenTofu and Terraform caches, or installed with `tofu init` |
| `--profile` defaulted to `default` for AWS | The SDK credential chain (environment, `AWS_PROFILE`, instance roles) |

## Building from source

```
go build .                           # every provider
go build -tags slim,aws -o unclick . # only the AWS importer, much smaller
go test ./...
UNCLICK_ACC=1 go test ./internal/... # against real provider plugins
```

## Support the project

Unclick is free and open source. If it saves you work, you can support it:

- [Boosty](https://boosty.to/mikio_kuroki/donate)
- USDT or TRX (TRC-20): `TXUBW4e88SDTfrnJRKfbhYfFcggufbonc1`
- USDT, USDC or ETH (ERC-20): `0x1378491169064702786b2E5b58c6375776177E8A`
- TON or USDT on TON: `UQAhI7EKzoa-JuKOfv0ULMzA3FrmpxsDkXj8Qevwj2z1cMRN`

## Credits and license

Unclick is based on Terraformer, created by Waze SRE and [The Terraformer Authors](AUTHORS). It is licensed under the [Apache License 2.0](LICENSE); a few files from HashiCorp projects keep the Mozilla Public License 2.0, see [NOTICE](NOTICE).

Unclick is an independent project, not affiliated with or endorsed by Google, HashiCorp or the OpenTofu project. Terraform is a trademark of HashiCorp. OpenTofu is a trademark of The Linux Foundation.
