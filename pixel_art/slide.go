package main

import (
    "net/http"
    "log"
    "github.com/gorilla/websocket"
	//"bytes"
	"fmt"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)           // broadcast channel


var (
    upgrader = websocket.Upgrader {
        // 读取存储空间大小
        ReadBufferSize:1024,
        // 写入存储空间大小
        WriteBufferSize:1024,
        // 允许跨域
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
    var (
        wbsCon *websocket.Conn
        err error
        data []byte
    )
    // 完成http应答，在httpheader中放下如下参数
    if wbsCon, err = upgrader.Upgrade(w, r, nil);err != nil {
        return // 获取连接失败直接返回
    }

    for {
        // 只能发送Text, Binary 类型的数据,下划线意思是忽略这个变量.
        if _, data, err = wbsCon.ReadMessage();err != nil {
            goto ERR // 跳转到关闭连接
        }
        if err = wbsCon.WriteMessage(websocket.TextMessage, data); err != nil {
            goto ERR // 发送消息失败，关闭连接
        }
    }

    ERR:
        // 关闭连接
        wbsCon.Close()
}

func main()  {
    http.HandleFunc("/ws",handleConnections)
	http.Handle(
    "/slide/", 
    http.StripPrefix(
        "/slide/", 
        http.FileServer(http.Dir("slide")),
    ),)
	
	go handleMessages()
	
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe", err.Error())
    }
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg []byte
		_,msg,err = ws.ReadMessage()
		fmt.Println(msg)
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
		 client.WriteMessage(websocket.BinaryMessage,msg)
		}
	}
}