.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker rmi -f zhanqi/webook:v0.0.3
	@docker build -t zhanqi/webook:v0.0.3 .
