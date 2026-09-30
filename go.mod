module github.com/Perruer/unclick

go 1.26.6

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/logging v1.20.0
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/Azure/go-autorest/autorest v0.11.30
	github.com/IBM-Cloud/bluemix-go v0.0.0-20220624043500-d538cb4fd9be
	github.com/IBM/go-sdk-core/v3 v3.3.1
	github.com/IBM/go-sdk-core/v4 v4.10.0
	github.com/IBM/go-sdk-core/v5 v5.23.4
	github.com/IBM/ibm-cos-sdk-go v1.15.1
	github.com/IBM/keyprotect-go-client v0.8.1
	github.com/IBM/networking-go-sdk v0.54.0
	github.com/IBM/platform-services-go-sdk v0.103.4
	github.com/IBM/vpc-go-sdk v0.91.0
	github.com/OctopusDeploy/go-octopusdeploy v1.8.6
	github.com/PaloAltoNetworks/pango v0.10.2
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.107
	github.com/aliyun/aliyun-tablestore-go-sdk v4.1.3+incompatible
	github.com/apache/openwhisk-client-go v0.0.0-20210106144548-17d556327cd3
	github.com/aws/aws-sdk-go-v2 v1.47.1
	github.com/aws/aws-sdk-go-v2/config v1.33.6
	github.com/aws/aws-sdk-go-v2/credentials v1.20.6
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.57.1
	github.com/aws/aws-sdk-go-v2/service/acm v1.50.1
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.50.0
	github.com/aws/aws-sdk-go-v2/service/appsync v1.62.1
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.78.1
	github.com/aws/aws-sdk-go-v2/service/batch v1.77.1
	github.com/aws/aws-sdk-go-v2/service/budgets v1.52.1
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.42.1
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.81.1
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.73.1
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.43.1
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.65.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.73.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchevents v1.41.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.88.1
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.78.1
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.43.1
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.45.1
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.55.1
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.42.1
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.74.1
	github.com/aws/aws-sdk-go-v2/service/configservice v1.74.1
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.40.0
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.50.0
	github.com/aws/aws-sdk-go-v2/service/docdb v1.57.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.69.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.336.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.66.1
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.47.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.99.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.49.1
	github.com/aws/aws-sdk-go-v2/service/eks v1.101.0
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.62.0
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.43.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.41.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.63.1
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.51.1
	github.com/aws/aws-sdk-go-v2/service/emr v1.70.1
	github.com/aws/aws-sdk-go-v2/service/firehose v1.52.1
	github.com/aws/aws-sdk-go-v2/service/glue v1.164.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.64.1
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.46.0
	github.com/aws/aws-sdk-go-v2/service/iot v1.84.1
	github.com/aws/aws-sdk-go-v2/service/kafka v1.65.1
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.56.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.61.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.110.0
	github.com/aws/aws-sdk-go-v2/service/medialive v1.111.1
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.48.1
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.38.1
	github.com/aws/aws-sdk-go-v2/service/mq v1.45.1
	github.com/aws/aws-sdk-go-v2/service/opsworks v1.31.0
	github.com/aws/aws-sdk-go-v2/service/organizations v1.60.1
	github.com/aws/aws-sdk-go-v2/service/qldb v1.32.2
	github.com/aws/aws-sdk-go-v2/service/rds v1.129.1
	github.com/aws/aws-sdk-go-v2/service/redshift v1.71.1
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.42.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.70.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.113.4
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.50.1
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.82.1
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.47.1
	github.com/aws/aws-sdk-go-v2/service/ses v1.42.1
	github.com/aws/aws-sdk-go-v2/service/sfn v1.51.1
	github.com/aws/aws-sdk-go-v2/service/sns v1.47.2
	github.com/aws/aws-sdk-go-v2/service/sqs v1.52.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.78.1
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.49.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.51.1
	github.com/aws/aws-sdk-go-v2/service/swf v1.43.1
	github.com/aws/aws-sdk-go-v2/service/waf v1.38.1
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.38.1
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.83.1
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.81.1
	github.com/aws/aws-sdk-go-v2/service/xray v1.45.1
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cloudflare/cloudflare-go v0.119.0
	github.com/cloudfoundry/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21 // indirect
	github.com/ddelnano/terraform-provider-mikrotik/client v0.0.0-20210401060029-7f652169b2c4
	github.com/ddelnano/terraform-provider-xenorchestra/client v0.18.0-alpha1
	github.com/denverdino/aliyungo v0.0.0-20200327235253-d59c209c7e93
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/digitalocean/godo v1.215.0
	github.com/fastly/go-fastly/v7 v7.5.5
	github.com/google/go-github/v35 v35.3.0
	github.com/gophercloud/gophercloud v1.14.1
	github.com/grafana/grafana-api-golang-client v0.27.0
	github.com/hashicorp/go-azure-helpers v0.82.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.8.0
	github.com/hashicorp/hcl v1.0.1-vault-7
	github.com/heimweh/go-pagerduty v0.0.0-20210930203304-530eff2acdc6
	github.com/heroku/heroku-go/v5 v5.5.0
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519 // indirect
	github.com/honeycombio/terraform-provider-honeycombio v0.55.0
	github.com/iancoleman/strcase v0.3.0
	github.com/ionos-cloud/sdk-go-dbaas-mongo v1.4.1
	github.com/ionos-cloud/sdk-go-dbaas-postgres v1.1.4
	github.com/ionos-cloud/sdk-go/v6 v6.3.11
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24
	github.com/jonboydell/logzio_client v1.2.0
	github.com/labd/commercetools-go-sdk v0.3.1
	github.com/linode/linodego v1.69.1
	github.com/microsoft/azure-devops-go-api/azuredevops v1.0.0-b5
	github.com/mrparkers/terraform-provider-keycloak v0.0.0-20221013232944-56f37a07590d
	github.com/nicksnyder/go-i18n v1.10.1 // indirect
	github.com/okta/okta-sdk-golang/v2 v2.20.0
	github.com/opsgenie/opsgenie-go-sdk-v2 v1.2.23
	github.com/packethost/packngo v0.31.0
	github.com/pkg/errors v0.9.1
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/tencentyun/cos-go-sdk-v5 v0.7.75
	github.com/vultr/govultr v1.1.1
	github.com/xanzy/go-gitlab v0.115.0
	github.com/yandex-cloud/go-genproto v0.123.0
	github.com/yandex-cloud/go-sdk v0.37.0
	github.com/zclconf/go-cty v1.18.1
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	golang.org/x/oauth2 v0.36.0
	golang.org/x/text v0.42.0
	gonum.org/v1/gonum v0.17.0
	google.golang.org/api v0.287.1
	google.golang.org/genproto v0.0.0-20260319201613-d00831a3d3e7
	k8s.io/apimachinery v0.34.1
	k8s.io/client-go v0.34.1
)

require (
	github.com/IBM-Cloud/container-services-go-sdk v0.0.0-20210705152127-41ca00fc9a62
	github.com/IBM/go-sdk-core v1.1.0
	github.com/mackerelio/mackerel-client-go v0.47.0
	github.com/okta/terraform-provider-okta v0.0.0-20210924173942-a5a664459d3b
)

require (
	github.com/antihax/optional v1.0.0 // indirect
	github.com/gofrs/uuid v4.3.0+incompatible // indirect
	github.com/launchdarkly/api-client-go v5.3.0+incompatible
)

require github.com/newrelic/newrelic-client-go v1.1.0

require (
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.1 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.2 // indirect
	github.com/Azure/go-autorest/tracing v0.6.1 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/appscode/go-querystring v0.0.0-20170504095604-0126cfb3f1dc // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.20 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.38.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.43.1 // indirect
	github.com/aws/smithy-go v1.28.1
	github.com/beevik/etree v1.1.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/crewjam/saml v0.4.13 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dghubble/sling v1.4.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/errors v0.22.8 // indirect
	github.com/go-openapi/strfmt v0.27.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-resty/resty/v2 v2.17.2 // indirect
	github.com/go-routeros/routeros v0.0.0-20210123142807-2a44d57c6730 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/jsonapi v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/gax-go/v2 v2.23.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.9.0
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mozillazg/go-httpheader v0.2.1 // indirect
	github.com/oklog/run v1.2.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/peterhellberg/link v1.1.0 // indirect
	github.com/russellhaering/goxmldsig v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/jsonrpc2 v0.0.0-20210201082850-366fbb520750 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.4 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
	gopkg.in/auth0.v5 v5.21.1
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.34.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

require github.com/PuerkitoBio/rehttp v1.0.0 // indirect

require (
	cloud.google.com/go/cloudbuild v1.25.0
	cloud.google.com/go/cloudtasks v1.13.7
	cloud.google.com/go/iam v1.11.0
	cloud.google.com/go/monitoring v1.24.3
	github.com/DataDog/datadog-api-client-go/v2 v2.66.0
	github.com/Myra-Security-GmbH/myrasec-go/v2 v2.60.0
	github.com/bradleyfalzon/ghinstallation/v2 v2.19.0
	github.com/manicminer/hamilton v0.44.0
	github.com/opalsecurity/opal-go v1.5.0
	gopkg.in/ns1/ns1-go.v2 v2.6.5
)

require (
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/longrunning v1.2.0 // indirect
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/Myra-Security-GmbH/signature v1.1.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.10.1 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dunglas/httpsfv v1.1.2 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-playground/validator/v10 v10.30.4 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-github/v88 v88.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.17 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/jsonapi v1.5.0 // indirect
	github.com/hashicorp/terraform-plugin-log v0.10.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx v1.2.29 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/ulid/v2 v2.1.2 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/valyala/fastjson v1.6.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.67.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260630182238-925bb5da69e7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260630182238-925bb5da69e7 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

require (
	github.com/IBM/continuous-delivery-go-sdk/v2 v2.0.20
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.53.0
	github.com/gofrs/uuid/v3 v3.1.2
	github.com/hashicorp/vault/api v1.22.0
	github.com/ionos-cloud/sdk-go-cert-manager v1.3.0
	github.com/ionos-cloud/sdk-go-container-registry v1.3.1
	github.com/ionos-cloud/sdk-go-dataplatform v1.1.1
	github.com/ionos-cloud/sdk-go-dns v1.4.0
	github.com/ionos-cloud/sdk-go-logging v1.3.0
	github.com/okta/okta-sdk-golang/v5 v5.0.2
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as v1.3.173
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs v1.3.179
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb v1.3.183
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn v1.3.173
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs v1.3.183
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb v1.3.186
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.186
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm v1.0.480
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.3.131
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es v1.3.176
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap v1.3.34
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb v1.3.183
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/pts v1.3.119
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis v1.3.175
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf v1.3.101
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses v1.3.136
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl v1.3.147
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tat v1.3.107
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcaplusdb v1.3.105
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc v1.3.185
)

replace gopkg.in/jarcoal/httpmock.v1 => github.com/jarcoal/httpmock v1.0.5

replace gopkg.in/ns1/ns1-go.v2 => github.com/ns1/ns1-go/v2 v2.6.5

replace github.com/tencentcloud/tencentcloud-sdk-go => github.com/tencentcloud/tencentcloud-sdk-go v1.0.392
