all:
	go build -o ./vmdk ./bin

generate:
	cd parser/ && binparsegen conversion.spec.yaml > vmdk_gen.go
