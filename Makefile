NPROC=2
CURL=curl -H 'content-type:application/json;' -v
MAXIMUM=100
STEP=2
REDIS_TEST=localhost:6379

.PHONY: integration_test test redis_test build mock

mock:
	${GOPATH}/bin/mockgen -package=db -destination db/db_mock.go github.com/dsociative/incrementer/db DB

proto:
	protoc --proto_path=${GOPATH}/src:. --twirp_out=. --go_out=. ./api/api.proto

lint:
	docker run --rm -v /dev/log:/dev/log -v $(shell pwd)/build:/build -v $(shell pwd):/go/src/github.com/dsociative/incrementer dsociative/gobuilder gometalinter --deadline 15m --vendored-linters --vendor -j $(NPROC) --disable-all -E unused -E gosimple -E deadcode src/github.com/dsociative/incrementer/...

test:
	go test ./... --count 1 -v

redis_start:
	docker run --rm -d --name redis -p 6379:6379 redis

redis_stop:
	docker rm -f redis

redis_test:
	REDIS_TEST=${REDIS_TEST} go test ./db --count 1 -v

integration_test:
	go test ./integration_test --count 1 -v -tags=integration

build:
	docker build -t incrementer .

compose-up:
	docker-compose up

incr:
	${CURL} "http://localhost:8080/twirp/incrementer.Incrementer/IncrementNumber" -d {}

get_number:
	${CURL} "http://localhost:8080/twirp/incrementer.Incrementer/GetNumber" -d {}

set_settings:
	${CURL} "http://localhost:8080/twirp/incrementer.Incrementer/SetSettings" -d '{"maximum": ${MAXIMUM}, "step": ${STEP}}'
