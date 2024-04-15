module benchmark

go 1.19

replace (
	client v0.0.0 => ../client/
	evio v0.0.0 => ../evio/
	gaio v0.0.0 => ../gaio/
	github.com/mzzsfy/go-async-adapter v0.0.0 => ../../../
	gnet v0.0.0 => ../gnet/
	nbio v0.0.0 => ../nbio/
	netpoll v0.0.0 => ../netpoll/
)

require (
	client v0.0.0
	evio v0.0.0
	github.com/gobwas/ws v1.3.2
	github.com/mzzsfy/go-async-adapter v0.0.0
	gnet v0.0.0
	nbio v0.0.0
	netpoll v0.0.0
)

require (
	github.com/bytedance/gopkg v0.0.0-20220413063733-65bf48ffb3a7 // indirect
	github.com/cloudwego/netpoll v0.6.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/kavu/go_reuseport v1.5.0 // indirect
	github.com/lesismal/nbio v1.5.3 // indirect
	github.com/panjf2000/gnet/v2 v2.3.5 // indirect
	github.com/tidwall/evio v1.0.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
