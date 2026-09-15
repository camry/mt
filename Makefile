.PHONY: proto

proto:
	protoc --go_opt=paths=source_relative --go_out=. cpb/*.proto
