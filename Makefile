build:
	docker build -t local:go-site-benchmark .
run:
	docker run --rm -d -p 8090:8090 --name site_benchmark local:go-site-benchmark
stop:
	docker stop site_benchmark
