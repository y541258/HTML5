package main

import (
	"net/http"
	"context"
	"io"
	"fmt"
	"time"
	"sync"
	"math/rand"
	"strconv"
	"strings"
	
	"github.com/gorilla/websocket"
	"github.com/gorilla/securecookie"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)           // broadcast channel
var wg sync.WaitGroup

var a []int
//var temp_a []byte = []byte{48,44,49,44,50,44,51,44,52}
var temp_a []byte = []byte("store,1,2,3,4,5")
var temp_user []byte
var user_n [5]string

var cookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
//請看http://www.gorillatoolkit.org/pkg/securecookie
//第一個參數是必要的,建議是32或64 bytes長度
//第二個參數是用來加密的,長度允許16,24,32 bytes,對應不同的AES加密法

var (
    upgrader = websocket.Upgrader {
        ReadBufferSize:1024,
        WriteBufferSize:1024,
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
)

func iinit() {
	temp_user = []byte("user,,,,,")
	
	for i:=0;i<5;i++{
		user_n[i]=""
	}
}

func SetCookieHandler(user_name string,res http.ResponseWriter) {
    value := map[string]string{
        "ID": user_name,
    }
    if encoded, err := cookieHandler.Encode("cookie-name", value); err == nil {
        cookie := &http.Cookie{
            Name:  "cookie-name",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(res, cookie)
    }
}

func ReadCookieHandler(res http.ResponseWriter, req *http.Request) (user_name string){
    if cookie, err := req.Cookie("cookie-name"); err == nil {
        value := make(map[string]string)
        if err = cookieHandler.Decode("cookie-name", cookie.Value, &value); err == nil {
            user_name = value["ID"]
			return user_name
        }
    }
	
	return
}

func clearCookie(res http.ResponseWriter) {
     cookie := &http.Cookie{
         Name:   "cookie-name",
         Value:  "",
         Path:   "/",
         MaxAge: -1,
		 Expires:time.Unix(1, 0),
     }
     http.SetCookie(res, cookie)
 }

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
	
	SetCookieHandler(ID,res)
	http.Redirect(res, req, "/store", 302)
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	clearCookie(res)
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
				SetCookieHandler(ID,res)
				http.Redirect(res, req, "/store", 302)
				return
			}
	}
	
	http.Redirect(res, req, "/login_fail", 302)
	return
}

func storePage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		if ReadCookieHandler(res,req) != ""{
			http.ServeFile(res, req, "store.html")
			return
		} else {
			http.Redirect(res, req, "/login", 302)
			return
		}
	}	
}

func jsPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "jquery-3.3.1.js")
		return
	}
}

func jsonPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "store_item.json")
		return
	}	
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
	//broadcast <- temp_a
	ws.WriteMessage(websocket.BinaryMessage,temp_a)
	ws.WriteMessage(websocket.BinaryMessage,temp_user)
	
	for {
		var msg []byte
		
		_,msg,err = ws.ReadMessage()
		s := strings.Split(string(msg),",")
		if s[0] == "buy"{
			client, err := mongo.Connect(context.Background(), "mongodb+srv://USERNAME:PASSWORD@USERNAME-keqxy.gcp.mongodb.net/admin", nil)
			db := client.Database("GCP_test")
			coll := db.Collection("GCP_test")
			
			var name string = ReadCookieHandler(w,r)

			_, _ = coll.UpdateOne(
				context.Background(),
				bson.NewDocument(
					bson.EC.String("name", name),
				),
				bson.NewDocument(
				    bson.EC.SubDocumentFromElements("$push",
					bson.EC.String("item",s[1]),
					),
				),
			)
			
			var temp_value,_ = strconv.Atoi(s[2])
			
			user_n[temp_value]=name
			
			temp_user = []byte("user")
			
			for i:=0;i<5;i++{
				temp_user = append(temp_user,[]byte(",")...)
				temp_user = append(temp_user,[]byte(user_n[i])...)
			}
			
			for client := range clients {
				client.WriteMessage(websocket.BinaryMessage,temp_user)
			}
		
			broadcast <- msg
		
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
		 client.WriteMessage(websocket.BinaryMessage,msg)
		}
		
		fmt.Println(msg)
	}
}

func UpdateStoreIn2Min() {
	defer wg.Done()
	
	for i:=0;i<30;i++{
	 a=append(a,i)
	}
	
	t := time.Now()
	rounded := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+1, 0, 0, t.Location())
	 
	time.AfterFunc(rounded.Sub(t),func() {
	 
	ticker := time.NewTicker(time.Second * 20)
	
	rand.Shuffle(len(a), func(i, j int) {
			a[i], a[j] = a[j], a[i]
			})
			
			temp_a=temp_a[:0]
			iinit();
			
			temp_a=[]byte("store,")
			
			for i:=0;i<5;i++{
				temp_a=append(temp_a,[]byte(strconv.Itoa(a[i]))...)
				temp_a=append(temp_a,[]byte(",")...)
			}
			temp_a=temp_a[:len(temp_a)-1]
			broadcast<-temp_a
	 
	/*defer ticker.Stop()
	done := make(chan bool)*/
	 
	go func() {
		/*time.Sleep(60 * time.Second)
		done <- true*/
	}()
	
	for {
		select {
		/*case <-done:
			fmt.Println("Done!")
			return*/
		case ti := <-ticker.C:
			rand.Shuffle(len(a), func(i, j int) {
			a[i], a[j] = a[j], a[i]
			})
			fmt.Println(a,ti)
			
			temp_a=temp_a[:0]
			iinit()
			
			temp_a=[]byte("store,")
			
			for i:=0;i<5;i++{
				temp_a=append(temp_a,[]byte(strconv.Itoa(a[i]))...)
				temp_a=append(temp_a,[]byte(",")...)
			}
			temp_a=temp_a[:len(temp_a)-1]
			broadcast<-temp_a
		}
	}
})
}

func main() {
	iinit()
	rand.Seed(time.Now().UnixNano())
	wg.Add(1)
	
	http.HandleFunc("/ws",handleConnections)
	http.HandleFunc("/server_error", server_error)
	http.HandleFunc("/account_repeat", account_repeat)
	http.HandleFunc("/login_fail", login_fail)
	http.HandleFunc("/sign_up", signupPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/store", storePage)
	http.HandleFunc("/jquery-3.3.1.js", jsPage)
	http.HandleFunc("/store_item.json", jsonPage)
	
	go handleMessages()
	go UpdateStoreIn2Min()
	
	wg.Wait()
	
    http.ListenAndServe(":8080", nil)
}
