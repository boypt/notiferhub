module github.com/boypt/notiferhub

go 1.17

require (
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/anacrolix/torrent v1.39.1
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.4.0
	github.com/mdlayher/genetlink v1.2.0
	github.com/mdlayher/netlink v1.6.0
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/viper v1.10.1
	github.com/ti-mo/conntrack v0.4.0
	github.com/ti-mo/netfilter v0.3.1
	golang.org/x/text v0.3.7
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20220208144051-fde48d68ee68
)

require (
	github.com/anacrolix/missinggo v1.3.0 // indirect
	github.com/anacrolix/missinggo/v2 v2.5.2 // indirect
	github.com/andybalholm/cascadia v1.0.0 // indirect
	github.com/bradfitz/iter v0.0.0-20191230175014-e8f45d346db8 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/josharian/native v1.0.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mdlayher/socket v0.1.1 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20220208050332-20e1d8d225ab // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220207234003-57398862261d // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20220202223031-3b95c81cc178 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace golang.zx2c4.com/wireguard/wgctrl => github.com/boypt/wgctrl-go v0.0.0-20220214075737-dacb5d2aebb9
