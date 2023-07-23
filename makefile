up:
	docker-compose up -d
order:
	cd services/order && go run main.go
cook:
	cd services/cook && go run main.go
storage:
	cd services/storage && go run main.go
table:
	cd services/tableRunner && go run main.go
payment:
	cd services/payment && go run main.go
error:
	cd services/errorHandler && go run main.go
all: up order cook storage table payment error
