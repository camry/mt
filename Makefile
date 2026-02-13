.PHONY: proto

proto:
	protoc-camry --go-camry_opt=paths=source_relative --go-camry_out=. cpb/*.proto
