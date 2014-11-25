package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

type router struct {
	*httprouter.Router
}

func newRouter() *router {
	return &router{httprouter.New()}
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

func (r *router) Put(path string, handler http.Handler) {
	r.PUT(path, wrapHandler(handler))
}

func (r *router) Patch(path string, handler http.Handler) {
	r.PATCH(path, wrapHandler(handler))
}

func (r *router) Delete(path string, handler http.Handler) {
	r.DELETE(path, wrapHandler(handler))
}

type appError struct {
	err     error
	code    int
	message string
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		if err.err != nil {
			log.Println(err.err)
		}
		http.Error(w, err.message, err.code)
	}
}

var (
	errParamsTypeAssertion = errors.New("Can't get URL parameters")
	errEmptyParameter      = errors.New("Empty URL parameter")
)

func serveHello(w http.ResponseWriter, r *http.Request) *appError {
	params, ok := context.Get(r, "params").(httprouter.Params)
	if !ok {
		return &appError{
			err:     errParamsTypeAssertion,
			code:    http.StatusInternalServerError,
			message: http.StatusText(http.StatusInternalServerError),
		}
	}
	name := params.ByName("name")
	if name == "" {
		return &appError{
			err:     errEmptyParameter,
			code:    http.StatusBadRequest,
			message: http.StatusText(http.StatusBadRequest),
		}
	}
	_, err := w.Write([]byte("Hello " + name + "\n"))
	if err != nil {
		return &appError{
			err:     err,
			code:    http.StatusInternalServerError,
			message: http.StatusText(http.StatusInternalServerError),
		}
	}
	return nil
}

func main() {
	router := newRouter()
	router.Get("/hello/:name/", appHandler(serveHello))
	log.Fatalln(http.ListenAndServe(":8080", router))
}
