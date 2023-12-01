
# go params
GOCMD=go

# normal entry points
	
update:
	clear 
	@echo "updating dependencies..."
	@go get -u -t ./...
	@go mod tidy 

build:
	clear 
	@echo "building..."
	@$(GOCMD) build .
	
test-first:
	clear
	@echo "testing serviceworks primary auth functions..."
	@$(GOCMD) test -run TestFirst ./...

test-second:
	clear
	@echo "test serviceworks second level functions..."
	@$(GOCMD) test -run TestSecond ./...
