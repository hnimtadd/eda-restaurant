up:
	docker-compose up -d
order:
	cd services/order && go run main.go
cook:
	cd services/cook && go run main.go
storage:
	cd services/storage && go run main.go
