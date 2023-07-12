APP_NAME=app

clean:
	rm -f ${APP_NAME}

build: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o ${APP_NAME} cmd/main.go

run: build
	./${APP_NAME} local

.PHONY: test
test:
	go test -v -count=1 ./...

test100:
	go test -v -count=100 ./...

lint:
	go fmt ./...

race:
	go test -v -race -count=1 ./...

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: gen
gen:
	mockgen yandex-team.ru/bstask/internal/order OrderRepository > ./internal/pkg/repository/order/mocks/mock_repo.go
	mockgen yandex-team.ru/bstask/internal/courier CourierRepository > ./internal/pkg/repository/courier/mocks/mock_repo.go

create_test_db:
	PGPASSWORD=password psql -h localhost -p 5432 -U postgres -tc "CREATE DATABASE lavka_test"