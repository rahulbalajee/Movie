.PHONY: run-metadata run-rating run-movie run-all consul-up consul-down consul-logs

run-metadata:
	go run metadata/cmd/main.go

run-rating:
	go run rating/cmd/main.go

run-movie:
	go run movie/cmd/main.go

run-all:
	$(MAKE) run-metadata & \
	$(MAKE) run-rating & \
	$(MAKE) run-movie & \
	wait

consul-up:
	@docker start consul 2>/dev/null || \
		docker run -d --name=consul \
			-p 8500:8500 -p 8600:8600/udp \
			hashicorp/consul:latest \
			agent -dev -client=0.0.0.0
	@echo "Consul UI: http://localhost:8500"

consul-down:
	@docker rm -f consul 2>/dev/null || true

consul-logs:
	docker logs -f consul

run-rating-producer:
	go run cmd/ratingproducer/main.go

rating-producer-up:
	docker-compose -f cmd/ratingproducer/docker-compose.yaml up -d

rating-producer-down:
	docker-compose -f cmd/ratingproducer/docker-compose.yaml down

rating-producer-logs:
	docker-compose -f cmd/ratingproducer/docker-compose.yaml logs -f
