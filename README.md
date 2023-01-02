# tswap
Package tswap automatically updates an html/template.Template when files in the directory where the template definitions are stored are updated. This can be useful if you want to work on changes to a website's UI without recompiling your application every time minor updates to a template definition file are made. tswap is dependent on github.com/fsnotify/fsnotify, which works for most, but not all, commonly used OS's.
# Example
```go
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/cdillond/tswap"
)

type App struct {
	Mux *http.ServeMux
    	Rwm sync.RWMutex
	T   *template.Template
}

func main() {
	a := App{
		Mux: http.NewServeMux(),
        	Rwm: sync.RWMutex{},
	}
	dir := `templates/`
	t, err := template.ParseGlob(dir + `*`)
	if err != nil {
		panic(err)
	}
	a.T = t

	errChan := tswap.AutoUpdate(a.T, dir, &a.Rwm)
	go func() {
		for {
			err = <-errChan
			fmt.Println(err)
		}
	}()

	a.Mux.HandleFunc("/", a.Index)

	http.ListenAndServe(":1234", a.Mux)
}

func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	a.Rwm.RLock()
	t, err := a.T.Lookup(`index.html`).Clone()
    	a.Rwm.RUnlock()
	if err != nil {
		return
	}
	t.Execute(w, struct{}{})
}
```
