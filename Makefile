all: docker-arm64
main.amd64: main.go
	CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w" -o main.amd64
main.arm64: main.go
	CGO_ENABLED=0 GOOS="linux" GOARCH="arm64" go build -ldflags="-s -w" -o main.arm64
docker-arm64: main.arm64
	podman build -f Dockerfile.scratch -t 192.168.12.50:5000/jeanribes/my-ip:main .
	podman push --tls-verify=false 192.168.12.50:5000/jeanribes/my-ip:main
