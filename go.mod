module github.com/commitdev/zero

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12
	github.com/buger/goterm v1.0.0
	github.com/coreos/go-semver v0.2.0
	github.com/gabriel-vasile/mimetype v1.1.1
	github.com/google/go-cmp v0.3.1
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-getter v1.4.2-0.20200106182914-9813cbd4eb02
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/terraform v0.12.26
	github.com/iancoleman/strcase v0.1.2
	github.com/juju/ansiterm v0.0.0-20210706145210-9283cdf370b5 // indirect
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/kyokomi/emoji v2.1.0+incompatible
	github.com/logrusorgru/aurora v0.0.0-20191017060258-dc85c304c434
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/machinebox/graphql v0.2.2
	github.com/manifoldco/promptui v0.8.0
	github.com/matryer/is v1.3.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	github.com/termie/go-shutil v0.0.0-20140729215957-bcacb06fecae
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.2.2

)

// Tencent cloud unpublished their version v3.0.82 and became v1.0.191
// https://github.com/hashicorp/terraform/issues/29021
replace github.com/tencentcloud/tencentcloud-sdk-go v3.0.82+incompatible => github.com/tencentcloud/tencentcloud-sdk-go v1.0.191
