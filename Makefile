

build:
	swag init
	go build

run: build
	./mockoauth --listen 127.0.0.1:8080 --host quant67.com

clean:
	rm -rf mockoauth
