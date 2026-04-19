.PHONY: run-metadata run-rating run-movie run-all

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
