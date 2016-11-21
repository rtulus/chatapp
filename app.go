package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rtulus/chatapp/src/gorilla"
)

var homeTemplate = template.Must(template.ParseFiles("home.html"))

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("index")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(w, r.Host)
}

func main() {

	// init websocket
	gorilla.InitWebsocket()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/websocket/open", gorilla.ServeWebsocket)

	log.Println("starting chat app [localhost:8080]")
	log.Fatal(http.ListenAndServe(":8080", router))

}
