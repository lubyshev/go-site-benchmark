# go-site-benchmark

## Notes

Все настроено на 8090 порты. 

Смотреть в браузере по адресу: 

[http://localhost:8090/sites?search=](http://localhost:8090/sites?search=)

## Increase open files limit (current session only)

Возможно не хватит лимита по открытым файлам.

```bash
ulimit -n 65535
```

## Generate client certificate

Для работы по протоколу HTTPS необходимо сгенерировать клиентский сертификат.

```bash
make gen-cert
```

## Build docker image

```bash
make build
```

## Run docker container

```bash
make run
```

## Show docker container logs

```bash
make logs
```

## Stop docker container

```bash
make stop
```
