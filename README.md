# slagbot

# Compiling

More optimized build: 

- `go build -ldflags="-s -w" -buildmode=plugin -o testplugin.plugin ./examples/testplugin.go`
- `go build -ldflags="-s -w" -o slagbot cmd/slagbot/main.go`

Packing with upx:

- `upx -9 -k <target>`