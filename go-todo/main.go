package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostname       string = "localhost:27017"
	dbname         string = "demo_todo"
	collectionname string = "todo"
	port           string = ":9000"
)

type (
	todomodel struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		Completed bool          `bson:"completed"`
		CreatedAt time.Time     `bson:"created"`
	}
	todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func into() {
	rnd = renderer.New()
	sess, err := mgo.Dial(hostname)
	checkError(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbname)
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"/static/home.tpl"}, nil)
	checkError(err)
}
func fetchTodo(w http.ResponseWriter, r *http.Request) {
	todos := []todomodel{}
	if err := db.C(collectionname).Find(bson.M{}).All(&todos); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "feild to fetch the error",
			"error":   err,
		})
		return
	}
	todoList := []todo{}
	for _, t := range todos {
		todoList = append(todoList, todo{
			ID:        t.ID.Hex(),
			Title:     t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}
func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}
	if t.Title == "" {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "the title is require",
		})
		return
	}
	tm := todomodel{
		ID:        bson.NewObjectId(),
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	if err := db.C(collectionname).Insert(&tm); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "failed to save todo",
			"error":   err,
		})
		return
	}
	rnd.JSON(w, http.StatusProcessing, renderer.M{
		"message": "todo create list",
		"todo_id": tm.ID.Hex(),
	})

}
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandler())
	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		log.Println("listen on the port", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Println("listen:%s\n", err)
		}
	}()
}
func todoHandler() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodo)
		r.Post("/", createTodo)
		r.Put("/", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
