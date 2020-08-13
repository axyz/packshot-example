build:
	docker build -t axyz/packshot-example .

run:
	docker run --rm -p 9090:9090 axyz/packshot-example
