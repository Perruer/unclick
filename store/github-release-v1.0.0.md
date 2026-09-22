## Unclick 1.0.0

Unclick turns resources that already exist in your cloud accounts into OpenTofu / Terraform code with `import` blocks. It continues [Terraformer](https://github.com/GoogleCloudPlatform/terraformer), archived in March 2026, on a new core.

### Highlights

- **Current providers work.** A new plugin client speaks protocol 5 and 6, so providers built on the plugin framework (Cloudflare v5 and others) no longer fail with "Incompatible API version".
- **OpenTofu and Terraform.** Providers are found in both tools' caches or installed with `tofu init`.
- **`unclick scan`.** Discovery through the providers' own list resources: 241 AWS resource types, no Unclick code per type. OpenTofu cannot use list resources yet; Unclick can.
- **Configuration that plans.** Each resource is validated by the provider before it is written; `imports.tf` replaces the old state file. The end-to-end tests expect `tofu plan` to show imports only.

```
unclick scan aws --config region=eu-west-1
cd generated/aws && tofu init && tofu plan
```

### Coming from Terraformer

Same importers and flags under `unclick import`. The state file and `--state` / `--bucket` are gone (import blocks replace them), the default layout is one directory per provider, and AWS uses the SDK credential chain. See the [README](https://github.com/Perruer/unclick#coming-from-terraformer) and the [changelog](https://github.com/Perruer/unclick/blob/main/CHANGELOG.md).

### Downloads

One binary per platform with every provider. Unpack it and put `unclick` in your `PATH`; `tofu` or `terraform` must be installed too. `checksums.txt` has the SHA-256 of every archive.
