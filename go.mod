module awesomeProject/awesomeProject4

go 1.16

require (
	github.com/beego/beego/v2 v2.0.1
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/ethereum/go-ethereum v1.10.7
	github.com/ontio/ontology v1.14.1-alpha
	github.com/ontio/ontology-crypto v1.2.1
	github.com/ontio/ontology-go-sdk v1.12.4
	github.com/onto/ontogo-sdk v1.0.0
	github.com/polynetwork/bridge-common v0.0.26
	github.com/polynetwork/poly v1.3.1
	github.com/polynetwork/poly-go-sdk v0.0.0-20210114035303-84e1615f4ad4
)

replace (
	github.com/ethereum/go-ethereum => github.com/ethereum/go-ethereum v1.9.25
	github.com/ontio/ontology => github.com/ontio/ontology v1.14.0-beta.0.20210818114002-fedaf66010a7
	github.com/onto/ontogo-sdk v1.0.0 => /Users/rain/go/gopath/src/github.com/me/fork-onto-go-sdk/ontology-go-sdk
)
