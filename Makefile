example_request:
	curl -v -i -H 'Content-Type: application/x-yaml' -u admin:admin -XPUT http://localhost:8888/v1/pools/mypool/rules/promgen --data-binary @ruleexample.yaml