### Use with Yandex Cloud

Example:

```
export YC_TOKEN=[YANDEX_CLOUD_OAUTH_OR_IAM_TOKEN]
./unclick import yandex -r subnet --folder_ids <comma-separated folder IDs>
```

List of supported Yandex resources:

*   `disk`
    * `yandex_compute_disk`
*   `instance`
    * `yandex_compute_instance`
*   `network`
    * `yandex_vpc_network`
*   `subnet`
    * `yandex_vpc_subnet`

The configuration and its `imports.tf` are written by default to
`generated/yandex/`. Run `tofu plan` there to adopt the resources.

The Yandex Cloud provider is not in the OpenTofu registry: Unclick installs it
from registry.terraform.io, or from the Yandex mirror
(terraform-mirror.yandexcloud.net) if your OpenTofu CLI configuration uses it.