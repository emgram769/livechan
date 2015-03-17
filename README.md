# livechan
A chat server written in Go.

### Installation

- Install Go.

- Install go imagick
```
sudo apt-get --no-install-recommends install libmagickwand-dev
go get github.com/gographics/imagick/imacick
```

- Install gorilla/websocket.
```
go get github.com/gorilla/websocket
```
- Install captcha lib
```
go get github.com/dchest/captcha
```
- Install sqlite3 drivers.
```
go get github.com/mattn/go-sqlite3
```
- Get the source code.
```
git clone https://github.com/majestrate/livechan.git
```
- Run the server.
```
go run *.go
```
- Open a browser and go to `localhost:18080`.
