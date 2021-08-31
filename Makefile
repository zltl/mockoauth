

build:
	swag init
	go build

run: build
	./mockoauth --listen :8080 --host quant67.com

clean:
	rm -rf mockoauth
