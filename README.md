# go-websocket-chat-server

This is the code that I use I use to serve a chat app on my website.  It keeps the client updated with the names of all connected users by using websockets, go routines and channels to provide low latency while allowing for huge scaling.


Steps to run the server:
1) Create a .env file in the same folder and add the following:
  PORT=3000
2) Install the latest version of Golang
3) Start the server from bash with the following command:
  cat .env | xargs go run main.go
 
