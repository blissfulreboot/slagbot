mock:
	go build -ldflags="-s -w" -o mock cmd/mock/main.go

slagbot:
	go build -ldflags="-s -w" -o slagbot cmd/slagbot/main.go

testplugin:
	go build -ldflags="-s -w" -buildmode=plugin -o testplugin.plugin ./examples/testplugin.go

all: mock slagbot testplugin

upx-mock:
	upx -9 -k mock
	rm mock.~

upx-slagbot:
	upx -9 -k slagbot
	rm slagbot.~

upx-all: upx-mock upx-slagbot

all-with-upx: all upx-all

clean:
	rm -f slagbot* mock* testplugin.*

