module github.com/Luzifer/lounge-control

go 1.14

replace github.com/Luzifer/lounge-control/sioclient => ./sioclient

require (
	github.com/Luzifer/go_helpers/v2 v2.10.0
	github.com/Luzifer/lounge-control/sioclient v0.0.0-00010101000000-000000000000
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sacOO7/go-logger v0.0.0-20180719173527-9ac9add5a50d // indirect
	github.com/sacOO7/gowebsocket v0.0.0-20180719182212-1436bb906a4e
	github.com/sirupsen/logrus v1.6.0
)
