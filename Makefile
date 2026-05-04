.PHONY: run-metadata run-rating run-movie run-all \
	build-metadata build-rating build-movie build-all clean \
	consul-up consul-down consul-logs \
	run-rating-producer rating-producer-up rating-producer-down rating-producer-logs \
	k8s-apply k8s-delete k8s-status

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

build-metadata:
	docker build -f metadata/Dockerfile -t movie/metadata:latest .

build-rating:
	docker build -f rating/Dockerfile -t movie/rating:latest .

build-movie:
	docker build -f movie/Dockerfile -t movie/movie:latest .

build-all:
	$(MAKE) build-metadata & \
	$(MAKE) build-rating & \
	$(MAKE) build-movie & \
	wait

clean:
	docker rmi -f movie/metadata:latest movie/rating:latest movie/movie:latest 2>/dev/null || true

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

# ---------------------------------------------------------------------------
# Kubernetes (services only)
# ---------------------------------------------------------------------------
# Infrastructure (Consul, MySQL, Kafka) is installed manually via Helm.
# See k8s/README.md (or the values files in k8s/) for the install commands.
# These targets only manage the three application Deployments + Services.

k8s-apply:
	kubectl apply -f metadata/kubernetes-deployment.yaml
	kubectl apply -f rating/kubernetes-deployment.yaml
	kubectl apply -f movie/kubernetes-deployment.yaml

k8s-delete:
	kubectl delete -f metadata/kubernetes-deployment.yaml --ignore-not-found
	kubectl delete -f rating/kubernetes-deployment.yaml --ignore-not-found
	kubectl delete -f movie/kubernetes-deployment.yaml --ignore-not-found

k8s-status:
	@echo "=== Pods (default) ==="
	@kubectl get pods
	@echo ""
	@echo "=== Infrastructure ==="
	@kubectl get pods -n consul 2>/dev/null || echo "consul namespace not found"
	@kubectl get pods -n mysql 2>/dev/null  || echo "mysql namespace not found"
	@kubectl get pods -n kafka 2>/dev/null  || echo "kafka namespace not found"
