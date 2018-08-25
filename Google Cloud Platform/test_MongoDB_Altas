package main

import (
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"context"
	"io"

	"github.com/mongodb/mongo-go-driver/mongo"
)

func hello(res http.ResponseWriter, req *http.Request) {
    res.Header().Set(
        "Content-Type", 
        "text/html",
    )
    
	client, err := mongo.Connect(context.Background(), "mongodb+srv://USERNAME:PASSWORD@USERNAME-keqxy.gcp.mongodb.net/admin", nil)
	db := client.Database("DATABASE_NAME")
	coll := db.Collection("COLLECTION_NAME")
  
  //這邊可能我需要我找個時間錄或實況MongoDB_Altas的影片
	
	_, err = coll.InsertOne(
			context.Background(),
			bson.M{"name": "Tommy"})
			if(err != nil){
			 return
			}
			
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

func main() {
    http.HandleFunc("/hello", hello)
    http.ListenAndServe(":8080", nil)
}
