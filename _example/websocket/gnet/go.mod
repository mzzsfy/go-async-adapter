module gnet

go 1.18

replace (
	client v0.0.0 => ../client/
	github.com/mzzsfy/go-async-adapter v0.0.0 => ../../../
)

require (
	client v0.0.0
	github.com/mzzsfy/go-async-adapter v0.0.0
	github.com/panjf2000/gnet/v2 v2.3.5
)

require (
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
