module nbio

go 1.19

replace (
	client v0.0.0 => ../client/
	github.com/mzzsfy/go-async-adapter v0.0.0 => ../../../
)

require (
	client v0.0.0
	github.com/lesismal/nbio v1.5.3
)

require (
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/mzzsfy/go-async-adapter v0.0.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
)
