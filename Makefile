init:
	docker run --name phmis_redis -d -p 6379:6379  redis:7-alpine 
	docker run --name phmis_postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=ch@mpi*ns -d -p 5432:5432  postgres:16-alpine
	until docker exec -it phmis_postgres pg_isready; do sleep 1; done
	sleep 1
	docker exec -it phmis_postgres createdb -U root dev_phcmis
	sleep 1
	make run


start:
	docker start phmis_postgres
	docker start phmis_redis
	make run
	
clean:
	make stop
	make prune
	make init	

stop:
	docker stop phmis_postgres
	docker stop phmis_redis

prune:
	docker rm phmis_postgres
	docker rm phmis_redis

createdb:
	docker exec -it phmis_postgres createdb -U root dev_phcmis
dropdb:
	docker exec -it phmis_postgres dropdb -U root dev_phcmis
sqlc:
	sqlc generate

mock:
	mockgen -package mock -destination test/mock/store.go phcmis/databases/persist/db Store
	mockgen -package mock -destination test/mock/distributor.go phcmis/databases/redis/daemon TaskDistributor
run:
	go run cmd/main.go 
test:
	go test -v -cover -short ./...
migration:
	migrate create -ext sql -dir databases/persist/migrations -seq $(name)

burst_requests:
	for i in $$(seq 1 10000); do \
    echo -n "RQST $$i "; \
    curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"first_name\": \"Stevary\", \
        \"last_name\": \"Bosuben\", \
        \"username\": \"bstevary$$i\", \
        \"pf_number\": \"PF_NO$$i\",\
		\"pf_number\": \"PHC_ID$$i\",\
        \"gender\": \"male\", \
        \"email\": \"Stevarybosuben$$i@gmail.com\", \
        \"password\": \"Champions123\", \
        \"confirm_password\": \"Champions123\"}" \
    http://localhost:8080/internal/rigister; \
    echo ""; \
    done


ratelimit:
	for i in $$(seq 1 9999999); do \
	echo -n "RQST $$i "; \
	curl http://localhost:8080/health; \
		echo ""; \
    done





.PHONY: init start clean stop prune createdb dropdb sqlc mock run test migration burst_requests ratelimit