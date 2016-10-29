package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
)

var (
	addr string
)

func init() {
	flag.StringVar(&addr, "l", ":80", "addr to listen on")
	flag.Parse()
}

type SvrErr struct {
	Err  string
	Code int
}

func Abort(err string, code int) *SvrErr {
	panic(&SvrErr{Err: err, Code: code})
}

func Recover(w http.ResponseWriter, r *http.Request) {
	if v := recover(); v != nil {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		fmt.Printf("backtrace:\n%s\n", string(buf))

		fmt.Println("paniced with:", v)

		if se, ok := v.(*SvrErr); ok {
			http.Error(w, se.Err+"\n"+string(buf), se.Code)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}

		// stack trace
	}
}

func DoBiz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world\n"))
}

func DemoError(w http.ResponseWriter, r *http.Request) {
	// try connect to db, and failed, just abort excution
	Abort("db connection is lost", http.StatusInternalServerError)
}

type Mux struct{}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("req = '%+v'\n", r)
	defer Recover(w, r) //

	switch r.URL.Path {
	case "/":
		DoBiz(w, r)

	case "/db":
		DemoError(w, r)
		// Abort
		// painc

	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

func main() {
	fmt.Println("listen on:", addr)

	var mux Mux
	panic(http.ListenAndServe(addr, &mux))
}
