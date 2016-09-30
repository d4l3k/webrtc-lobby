# webrtc-lobby [![GoDoc](https://godoc.org/github.com/d4l3k/webrtc-lobby?status.svg)](https://godoc.org/github.com/d4l3k/webrtc-lobby) [![Build Status](https://travis-ci.org/d4l3k/webrtc-lobby.svg?branch=master)](https://travis-ci.org/d4l3k/webrtc-lobby)

This provides a lobby service for webrtc clients to connect to others. It provides a JSON-RPC api over websockets such that hosts can announce themselves and have clients to connect to them. Some features include: names, authors, passwords, location and hidden services.

There's a set of webcomponents designed to easily integrate with the lobby service.
https://github.com/d4l3k/webrtc-lobby-elements

A hosted copy is provided at `wss://fn.lc/lobby`.

## Running with docker (port 8080)
```bash
docker run --restart=always -d -p 8080:5000 d4l3k/webrtc-lobby
```

## License
Licensed under the MIT license.
