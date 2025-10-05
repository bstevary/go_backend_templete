init:
	docker run --name dev_redis -d -p 6379:6379  redis:7-alpine 
	docker run --name dev_postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=ch@mpi*ns -d -p 5432:5432  postgres:16-alpine
	until docker exec -it dev_postgres pg_isready; do sleep 1; done
	sleep 1
	docker exec -it dev_postgres createdb -U root dev_phcmis
	sleep 1
	make run


start:
	docker start dev_postgres
	docker stdev
	make run
	
clean:
	make stop
	make prune
	make init	

stop:
	docker stop dev_postgres
	docker stop phmis_redis

prune:
	docker rm dev_postgres
	docker rm phmis_redis

createdb:
	docker exec -it dev_postgres createdb -U root dev_phcmis
dropdb:
	docker exec -it dev_postgres dropdb -U root dev_phcmis
sqlc:
	sqlc generate

mock:
	mockgen -package mock -destination test/mock/database.go github.com/bstevary/hexagonal/database/db Database
	mockgen -package mock -destination test/mock/distributor.go github.com/bstevary/hexagonal/jobs TaskDistributor
run:
	go run cmd/main.go 
test:
	go test -v -cover -short ./...
migration:
	migrate create -ext sql -dir database/migrations -seq $(name)


.PHONY: init start clean stop prune createdb dropdb sqlc mock run test migration burst_requests ratelimit