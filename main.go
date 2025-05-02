package main

import (
	"net/http"
	"fmt"
)
func main(){
	//making router
	mux := http.NewServeMux()

	//defining handler
	mux.Handle("/",http.FileServer(http.Dir(".")))
	
	//setting up server
	server:= &http.Server{
		Handler:mux,
		Addr:":8080",
	}


	err:=server.ListenAndServe()
	if err!= nil{
		fmt.Println(err)
	}
	

}