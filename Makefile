gen-cert:
	openssl req -new -nodes -x509 -out certs/client.pem -keyout certs/client.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=noname@gmail.com"
build:
	docker build -t local:go-site-benchmark .
run:
	docker run --rm -d -p 8090:8090 --name site_benchmark local:go-site-benchmark
logs:
	docker logs site_benchmark
stop:
	docker stop site_benchmark
