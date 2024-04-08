module netpoll

go 1.19

replace (
	client v0.0.0 => ../client/
	github.com/mzzsfy/go-async-adapter v0.0.0 => ../../
)

require (
	client v0.0.0
	github.com/cloudwego/netpoll v0.6.0
	github.com/mzzsfy/go-async-adapter v0.0.0
)

require (
	github.com/bytedance/gopkg v0.0.0-20220413063733-65bf48ffb3a7 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	golang.org/x/sys v0.19.0 // indirect
)
