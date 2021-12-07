module github.com/TykTechnologies/tyk-sync

go 1.13

require (
	github.com/TykTechnologies/tyk v1.9.2-0.20211118114729-bc0667a2871c
	github.com/jensneuse/graphql-go-tools v1.20.2
	github.com/kataras/go-errors v0.0.3
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/ongoingio/urljoin v0.0.0-20140909071054-8d88f7c81c3c
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.4 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v1.0.0
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
)

replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20210909091849-83a2f375a915
