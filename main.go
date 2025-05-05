package main

import (
	"net/http"
	"fmt"
	"sync/atomic"
)
type apiConfig struct{
	fileserverHits atomic.Int32
}

func readinessHandler(w http.ResponseWriter,r *http.Request){
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) hitCountHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	body:= fmt.Sprintf("Hits: %d",cfg.fileserverHits.Load())
	w.Write([]byte(body))
}

func (cfg *apiConfig) resetHitHandler(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	w.Write([]byte("reset\n"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
	newNext:= http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w,r)
	})
	return newNext
}


func main(){
	//making our server hit counter
	
	apiCfg:=&apiConfig{}
	
	//making router
	mux := http.NewServeMux()

	//defining handlers
	fileserver:=http.FileServer(http.Dir("."))
	appHandler:=http.StripPrefix("/app",fileserver) //stripping path
	mux.Handle("/app/",apiCfg.middlewareMetricsInc(appHandler))
	mux.HandleFunc("/healthz",readinessHandler)
	mux.Handle("/metrics",http.HandlerFunc(apiCfg.hitCountHandler))
	mux.Handle("/reset",http.HandlerFunc(apiCfg.resetHitHandler))

	//setting up server
	server:= &http.Server{
		Handler:mux,
		Addr:":8080",
	}

	//running the server
	err:=server.ListenAndServe()
	if err!= nil{
		fmt.Println(err)
	}
	

}