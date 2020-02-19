package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/countries", GetAllCountries),
		rest.Post("/countries", PostCountry),
		rest.Get("/countries/:code", GetCountry),
		rest.Delete("/countries/:code", DeleteCountry),
		rest.Get("/100", Get101),
		rest.Get("/200", Get200),
		rest.Get("/300", Get302),
		rest.Get("/400", Get400),
		rest.Get("/500", Get500),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))

}

// Thing for Thing
type Thing struct {
	Name string
}

var b bool

// Get101 for to get 101
func Get101(w rest.ResponseWriter, r *rest.Request) {

	if b == false {
		b = true
		w.WriteHeader(http.StatusSwitchingProtocols)
		w.WriteJson(
			&Thing{
				Name: fmt.Sprintf("thing #%s", "suzuki1"),
			},
		)
	} else {
		b = false
		w.WriteHeader(http.StatusOK)
		w.WriteJson(
			&Thing{
				Name: fmt.Sprintf("thing #%s", "suzuki2"),
			},
		)
	}

	// cpt := 0
	// for {
	// 	cpt++
	// 	w.WriteJson(
	// 		&Thing{
	// 			Name: fmt.Sprintf("thing #%d", cpt),
	// 		},
	// 	)
	// 	w.WriteHeader(http.StatusSwitchingProtocols)
	// 	w.(http.ResponseWriter).Write([]byte("\n"))
	// 	// Flush the buffer to client
	// 	w.(http.Flusher).Flush()
	// 	// wait 1 seconds
	// 	time.Sleep(time.Duration(1) * time.Second)
	// 	if cpt == 10 {
	// 		return
	// 	}
	// }
}

// Get200 for to get 200
func Get200(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

// Get302 for to get 302
func Get302(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusFound)
	return
}

// Get400 for to get 400
func Get400(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusBadRequest)
	return
}

// Get500 for to get 500
func Get500(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	return
}

// Country model
type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

var lock = sync.RWMutex{}

// GetCountry for to get country
func GetCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")

	lock.RLock()
	var country *Country
	if store[code] != nil {
		country = &Country{}
		*country = *store[code]
	}
	lock.RUnlock()

	if country == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(country)
}

// GetAllCountries for to get all country
func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	countries := make([]Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = *country
		i++
	}
	lock.RUnlock()
	w.WriteJson(&countries)
}

// PostCountry for to register country
func PostCountry(w rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country.Code == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	lock.Lock()
	store[country.Code] = &country
	lock.Unlock()
	w.WriteJson(&country)
}

// DeleteCountry for to delete country
func DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	lock.Lock()
	delete(store, code)
	lock.Unlock()
	w.WriteHeader(http.StatusOK)
}
