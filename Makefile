all: bonniego 

.PHONY: bonniego 

bonniego:
	go build -a -o bonniego gobonniego/gobonniego.go
