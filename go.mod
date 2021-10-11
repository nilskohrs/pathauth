module github.com/nilskohrs/pathauth

go 1.17

require (
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.7.3
)

exclude (
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000
	github.com/go-check/check v0.0.0-00010101000000-000000000000
)

replace github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
