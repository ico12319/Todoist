COVERAGE_FILE := coverage.out
TEST_PKGS := ./...

MIGRATE         := migrate
MIGRATION_DIR   := internProject/migrations


.PHONY: test coverage create-migration clean
run:
	go run main.go

test:
	@go test $(TEST_PKGS) -v

coverage: clean
	@go test $(TEST_PKGS) -covermode=count -coverprofile=$(COVERAGE_FILE)
	@go tool cover -func=$(COVERAGE_FILE)


create-migration:
ifndef name
	$(error pass name: make create-migration name=<some name>)
endif
	@$(MIGRATE) create -ext sql -dir $(MIGRATION_DIR) -seq $(name)

clean:
	@rm -f $(COVERAGE_FILE)

generate_gql:
	@go run github.com/99designs/gqlgen@v0.17.73