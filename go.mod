module github.com/fezho/oidc-auth

go 1.12

require (
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	go.etcd.io/bbolt v1.3.4
	golang.org/x/crypto v0.0.0-20200423211502-4bdfaf469ed5 // indirect
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/square/go-jose.v2 v2.5.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
