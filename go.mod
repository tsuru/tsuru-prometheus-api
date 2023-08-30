module github.com/tsuru/tsuru-prometheus-api

go 1.16

require (
	github.com/labstack/echo/v4 v4.6.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.50.0
	github.com/prometheus/prometheus v1.8.2-0.20210914090109-37468d88dce8
	github.com/stretchr/testify v1.7.0
	github.com/tsuru/go-tsuruclient v0.0.0-20211213213525-0d2868229cfd
	golang.org/x/oauth2 v0.0.0-20220608161450-d0670ef3b1eb // indirect
	gopkg.in/yaml.v3 v3.0.0
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.1
)
