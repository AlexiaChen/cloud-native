package goservice

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry-community/go-cfenv"
	cftools "github.com/cloudnativego/cf-tools"
	"github.com/cloudnativego/cfmgo"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// NewServer configures and return new server
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON:    true,
		IsDevelopment: true,
	})

	n := negroni.Classic()
	mx := mux.NewRouter()

	repo := initRepository(nil)
	initRoutes(mx, formatter, repo)
	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render, repo matchRepository) {
	//  curl http://localhost:3000/test
	mx.HandleFunc("/test", testHandler(formatter)).Methods("GET")
	mx.HandleFunc("/matches", createMatchHandler(formatter, repo)).Methods("POST")
	mx.HandleFunc("/matches", getMatchListHandler(formatter, repo)).Methods("GET")
	mx.HandleFunc("/matches/{id}", getMatchDetailsHandler(formatter, repo)).Methods("GET")
	mx.HandleFunc("/matches/{id}/moves", addMoveHandler(formatter, repo)).Methods("POST")
}

func testHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"This is a test"})
	}
}

func initRepository(appEnv *cfenv.App) (repo matchRepository) {
	dbServiceURI, err := cftools.GetVCAPServiceProperty(dbServiceName, "url", appEnv)
	if err != nil || dbServiceURI == "" {
		if err != nil {
			fmt.Printf("Error getting database configuration: %v\n", err)
		}

		fmt.Println("MongoDB was not detected ; configuring in-memory repo")
		repo = newInMemoryRepository()
		return
	}

	matchCollection := cfmgo.Connect(cfmgo.NewCollectionDialer, dbServiceURI, MatchesCollectionName)
	fmt.Printf("Connecting to MongoDB service: %s ... \n", dbServiceName)
	repo = newMongoMatchRepository(matchCollection)
	return
}
