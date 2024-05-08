prepare:
	docker pull consul:1.15
	docker run -d --rm --name=dev-consul -e CONSUL_BIND_INTERFACE=eth0 -p 8500:8500 consul:1.15

stop:
	docker stop dev-consul