package main

import (
	"net/http"
	"context"
	"io"
	"fmt"
	"time"
	"math/rand"
	
	"github.com/gorilla/websocket"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)           // broadcast channel
var item = make(chan [5]int)

var (
    upgrader = websocket.Upgrader {
        ReadBufferSize:1024,
        WriteBufferSize:1024,
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

    if wbsCon, err = upgrader.Upgrade(w, r, nil);err != nil {
        return
    }

    for {
        if _, data, err = wbsCon.ReadMessage();err != nil {
            goto ERR
        }
        if err = wbsCon.WriteMessage(websocket.TextMessage, data); err != nil {
            goto ERR
        }
    }

    ERR:
        wbsCon.Close()
}

func hello(res http.ResponseWriter, req *http.Request) {
    res.Header().Set(
        "Content-Type", 
        "text/html",
    )
			
	io.WriteString(
        res, 
        `<doctype html>
<html>
    <head>
        <title>Hello World</title>
    </head>
    <body>
        Hello World!
    </body>
</html>`,
    )
}

func server_error(res http.ResponseWriter, req *http.Request) {
    res.Header().Set(
        "Content-Type", 
        "text/html",
    )
			
	io.WriteString(
        res, 
        `<doctype html>
<html>
    <head>
        <title>server error</title>
    </head>
    <body>
        server error
    </body>
</html>`,
    )
}

func account_repeat(res http.ResponseWriter, req *http.Request) {
    res.Header().Set(
        "Content-Type", 
        "text/html",
    )
			
	io.WriteString(
        res, 
        `<doctype html>
<html>
    <head>
        <title>account repeat</title>
    </head>
    <body>
        account repeat
    </body>
</html>`,
    )
}

func login_fail(res http.ResponseWriter, req *http.Request) {
   // if req.Method != "POST" {
	res.Header().Set(
        "Content-Type", 
        "text/html",
	)
			
	io.WriteString(
        res, 
        `<doctype html>
<html>
    <head>
        <title>login fail</title>
    </head>
    <body>
        Please check ID and password again.
    </body>
</html>`,
    )
	//}
}

func signupPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "go_with_mongodb_sign_up.html")
		return
	}

	ID := req.FormValue("ID")
	password := req.FormValue("password")
	
	client, err := mongo.Connect(context.Background(), "mongodb+srv://USERNAME:PASSWORD@USERNAME-keqxy.gcp.mongodb.net/admin", nil)
	db := client.Database("GCP_test")
	coll := db.Collection("GCP_test")

	cursor, err := coll.Find(
			context.Background(),
			bson.NewDocument(bson.EC.String("name", ID)))
		
	if err != nil {
		fmt.Println(err);
		return
	}
	
	for cursor.Next(context.Background()) {
			elem := bson.NewDocument()
			
			if err = cursor.Decode(elem); err != nil {
 				fmt.Println(err)
 			}
			
			check,check_result:= elem.LookupElement("name").Value().StringValueOK()
 			if  check == ID && check_result {
				http.Redirect(res, req, "/account_repeat", 302)
				return
			}
	}
	
	
	_, err = coll.InsertOne(
			context.Background(),
			bson.NewDocument(bson.EC.String("name", ID),
			bson.EC.String("password", password),
			bson.EC.Int32("coin",10000)))
			if(err != nil){
			 fmt.Println(err)
			 return
			}
	
	http.Redirect(res, req, "/hello", 302)
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "go_with_mongodb_login.html")
		return
	}	
	
	client, err := mongo.Connect(context.Background(), "mongodb+srv://USERNAME:PASSWORD@USERNAME-keqxy.gcp.mongodb.net/admin", nil)
	db := client.Database("GCP_test")
	coll := db.Collection("GCP_test")

	ID := req.FormValue("ID")
	password := req.FormValue("password")
	which_button := req.FormValue("button")
	
	if which_button == "Sign Up" {
		http.Redirect(res, req, "/sign_up", 302)
	}

	cursor, err := coll.Find(
			context.Background(),
			bson.NewDocument(bson.EC.String("name", ID),
			bson.EC.String("password", password)),
		)
	
	if err != nil {
		fmt.Println(err);
		return
	}
	
	for cursor.Next(context.Background()) {
			elem := bson.NewDocument()
			
			if err = cursor.Decode(elem); err != nil {
 				fmt.Println(err)
 			}
			
			check,check_result:= elem.LookupElement("name").Value().StringValueOK()
 			if  check == ID && check_result {
				http.Redirect(res, req, "/hello", 302)
				return
			}
	}
	
	http.Redirect(res, req, "/login_fail", 302)
	return
}

type Player struct {
	Name  string `bson:"name"`
	Password []string `bson:"password"`
	coin int `bson:"coin"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	defer ws.Close()

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
		msg := <-broadcast

		for client := range clients {
		 client.WriteMessage(websocket.BinaryMessage,msg)
		}
	}
}

func UpdateStoreIn2Min() {
	rand.Seed(time.Now().UnixNano())
	
	var i,j int
	var temp int
	var flag bool = true
	
	for i=0; i<5; i++{
		temp=rand.Intn(30)
		item[i]=temp
		
		for flag{
			flag=true
			for j=0; j<i; j++{
				if temp==item[j]{
					flag=false
				}
			temp=rand.Intn(30)
			item[i]=temp
			}
		}
	}
	
	for i=0; i<5; i++{
		fmt.Println(item[i])
	}
	
	time.Sleep(time.Second * 10)
}

func main() {
	http.HandleFunc("/ws",handleConnections)
    http.HandleFunc("/hello", hello)
	http.HandleFunc("/server_error", server_error)
	http.HandleFunc("/account_repeat", account_repeat)
	http.HandleFunc("/login_fail", login_fail)
	http.HandleFunc("/sign_up", signupPage)
	http.HandleFunc("/login", loginPage)
	
	go handleMessages()
	go UpdateStoreIn2Min()
	
    http.ListenAndServe(":8080", nil)
}
