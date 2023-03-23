proto:
	@protoc --go_out=. events.proto

run:
	@go run .
