<img src="docs/logo.svg" width="96" align="right" alt="Логотип Unclick">

# Unclick

**Анкликни своё облако: преврати ClickOps в код OpenTofu / Terraform.**

[English version](README.md)

Unclick находит ресурсы, которые уже есть в ваших облачных аккаунтах и SaaS-сервисах, и записывает их как инфраструктуру-как-код вместе с блоками `import`, чтобы один `tofu plan` взял их под управление:

```
$ unclick scan aws --config region=eu-west-1
$ cd generated/aws && tofu init && tofu plan
...
Plan: 12 to import, 0 to add, 0 to change, 0 to destroy.
```

Unclick продолжает [Terraformer](https://github.com/GoogleCloudPlatform/terraformer), который перевели в архив 16 марта 2026 года. Все 44 его импортёра на месте, но работают на новом ядре: с актуальными провайдерами, с OpenTofu и с Terraform.

## Чем Unclick лучше

- **Работает с актуальными провайдерами.** Внутри Terraformer был Terraform 0.12, который понимал только протокол плагинов версии 5. Провайдеры на новом фреймворке (Cloudflare v5 и другие) падали с ошибкой «Incompatible API version». Unclick понимает протоколы 5 и 6.
- **Работает с OpenTofu.** Находит провайдеры, установленные OpenTofu, а недостающие ставит через `tofu init`. С Terraform тоже работает.
- **Конфиг проходит `plan`.** Каждый сгенерированный ресурс перед записью проверяет сам провайдер, и аргументы, которые он отвергает, в конфиг не попадают. Сквозные тесты требуют, чтобы `tofu plan` показывал только импорт без изменений.
- **Блоки `import` вместо файлов состояния.** Terraformer писал state версии 3. При его обновлении OpenTofu и Terraform подставляют неверный адрес провайдера для большинства провайдеров не от HashiCorp. Unclick пишет `imports.tf`, и ваш state не трогается, пока вы сами не сделаете `apply`.
- **`scan` не требует кода под каждый ресурс.** Он просит провайдер перечислить то, что есть. Это тот же механизм, на котором построена команда `query` в Terraform 1.14; в OpenTofu его пока нет.

## Установка

Скачайте бинарник для своей платформы из [Releases](https://github.com/Perruer/unclick/releases) или соберите его на Go 1.24+:

```
go install github.com/Perruer/unclick@latest
```

Unclick нужен `tofu` (или `terraform`) в `PATH`. Через него скачиваются провайдеры, ровно как при `tofu init`, с вашими зеркалами и учётными данными.

## Два способа найти ресурсы

| | `unclick scan ПРОВАЙДЕР` | `unclick import ПРОВАЙДЕР` |
|---|---|---|
| Как ищет ресурсы | Спрашивает провайдер (list-ресурсы) | Импортёры Terraformer обращаются к API облака |
| Провайдеры | Любой провайдер с list-ресурсами: AWS (241 тип в версии 6.66), Google, AzureRM | 44 провайдера, список ниже |
| Новые типы ресурсов | Появляются с новыми версиями провайдера | Нужен новый импортёр |
| Ссылки между ресурсами | По ID (`vpc_id = aws_vpc.main.id`) | Из таблиц связей импортёров |

### scan

```
unclick scan aws --config region=eu-west-1
unclick scan aws --config region=eu-west-1 --types aws_vpc,aws_subnet,aws_security_group
unclick scan hashicorp/google --config project=my-project --config region=europe-west1
```

`--config ключ=значение` задаёт аргументы провайдера. Учётные данные берутся из окружения, как у OpenTofu. Без `--types` сканируются все типы, которые провайдер умеет перечислять.

### import

```
unclick import aws --regions=eu-west-1 --resources=vpc,subnet,sg
unclick import github --owner=my-org --resources=repositories
unclick import kubernetes --resources=deployments,services
unclick import aws list        # какие ресурсы поддерживает импортёр
```

Опции и ресурсы каждого провайдера описаны в [docs/](docs). Фильтры работают как в Terraformer, см. [docs/upstream-README.md](docs/upstream-README.md#filtering).

### Результат

По умолчанию файлы кладутся в `generated/<провайдер>/`:

- `provider.tf`: `required_providers` с правильным адресом и использованной версией провайдера, плюс блок провайдера;
- по одному `.tf` на каждый тип ресурса;
- `imports.tf`: по блоку `import` на каждый ресурс.

Запустите там `tofu init && tofu plan`. Если в плане только импорт, `tofu apply` возьмёт ресурсы под управление, ничего в них не меняя.

## Провайдеры

`import` поддерживает: AliCloud, Auth0, AWS, Azure, Azure AD, Azure DevOps, Cloudflare, commercetools, Datadog, DigitalOcean, Equinix Metal, Fastly, GitHub, GitLab, фильтры Gmail, Google Cloud, Grafana, Heroku, Honeycomb, IBM Cloud, IONOS Cloud, Keycloak, Kubernetes, LaunchDarkly, Linode, Logz.io, Mackerel, MikroTik, Myra Security, New Relic, NS1, Octopus Deploy, Okta, Opal, OpenStack, Opsgenie, PagerDuty, PAN-OS, RabbitMQ, Tencent Cloud, Vault, Vultr, Xen Orchestra, Yandex Cloud.

Проверено сквозными тестами: AWS на [moto](https://github.com/getmoto/moto) (и `import`, и `scan`) и GitHub на настоящем аккаунте. В CI также есть тест Kubernetes на [kind](https://kind.sigs.k8s.io). Остальные импортёры компилируются и работают на новом ядре, но заново ещё не проверялись, так что сообщения о проблемах очень приветствуются.

Провайдера Yandex Cloud нет в реестре OpenTofu. Unclick ставит его из registry.terraform.io или с зеркала Yandex (terraform-mirror.yandexcloud.net), если оно указано в вашей конфигурации OpenTofu.

## Переход с Terraformer

| Terraformer | Unclick |
|---|---|
| `terraformer import aws ...` | `unclick import aws ...` (те же импортёры и флаги) |
| `terraform.tfstate` рядом с кодом | `imports.tf`; запустите `tofu plan` / `apply` |
| `--state=bucket`, `--bucket` | Удалены: загружать больше нечего |
| Папка на каждый сервис | Папка на провайдер; старую раскладку вернёт `--path-pattern "{output}/{provider}/{service}/"` |
| Провайдеры в `~/.terraform.d/plugins` | Ищутся в кэшах OpenTofu и Terraform или ставятся через `tofu init` |
| `--profile` у AWS по умолчанию `default` | Стандартная цепочка SDK: окружение, `AWS_PROFILE`, роли инстанса |

## Сборка из исходников

```
go build .                           # все провайдеры
go build -tags slim,aws -o unclick . # только импортёр AWS, гораздо меньше
go test ./...
UNCLICK_ACC=1 go test ./internal/... # против настоящих плагинов
```

## Поддержать проект

Unclick бесплатный и с открытым кодом. Если он экономит вам время, можно поддержать разработку:

- [Boosty](https://boosty.to/mikio_kuroki/donate)
- USDT или TRX (TRC-20): `TXUBW4e88SDTfrnJRKfbhYfFcggufbonc1`
- USDT, USDC или ETH (ERC-20): `0x1378491169064702786b2E5b58c6375776177E8A`
- TON или USDT в сети TON: `UQAhI7EKzoa-JuKOfv0ULMzA3FrmpxsDkXj8Qevwj2z1cMRN`

## Авторы и лицензия

Unclick основан на Terraformer, который создали Waze SRE и [авторы Terraformer](AUTHORS). Лицензия — [Apache License 2.0](LICENSE). Несколько файлов из проектов HashiCorp остаются под Mozilla Public License 2.0, подробности в [NOTICE](NOTICE).

Unclick — независимый проект, не связанный с Google, HashiCorp или проектом OpenTofu и не одобренный ими. Terraform — торговая марка HashiCorp. OpenTofu — торговая марка The Linux Foundation.
