build:
	docker build -t local:go-site-benchmark .
run:
	docker run --rm -d --env-file ./etc/.env -p 8090:8090 --name site_benchmark local:go-site-benchmark
logs:
	docker logs site_benchmark
stop:
	docker stop site_benchmark
