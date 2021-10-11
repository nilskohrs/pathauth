module github.com/nilskohrs/pathauth

go 1.17

require (
    github.com/traefik/traefik/v2 v2.5.3
	github.com/containous/alice v0.0.0-20181107144136-d83ebdd94cbd // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.7.3
    github.com/miekg/dns v1.1.43 // indirect
    github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
    github.com/sirupsen/logrus v1.7.0 // indirect
    github.com/traefik/paerser v0.1.4 // indirect
    github.com/vulcand/predicate v1.1.0 // indirect
    golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2 // indirect
	github.com/gravitational/trace v0.0.0-20190726142706-a535a178675f // indirect
    github.com/jonboulle/clockwork v0.1.0 // indirect
    golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
    golang.org/x/net v0.0.0-20210428140749-89ef3d95e781 // indirect
    golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
)

replace github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898

exclude (
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000
	github.com/go-check/check v0.0.0-00010101000000-000000000000
)
