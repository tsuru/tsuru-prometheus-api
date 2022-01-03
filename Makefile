example_request:
	curl -v -i -H 'Content-Type: application/x-yaml' -u admin:admin -XPUT http://localhost:8888/v1/pools/my-pool/rules/promgen --data-binary @ruleexample.yaml

build_tsuru_deploy:
	- rm -Rf _build
	- mkdir -p _build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o _build/app .
	echo 'web: ./app' > _build/Procfile
