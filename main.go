package main

import (
	"net/http"
	"fmt"
	"sync/atomic"
	"encoding/json"
	"strings"
)

type apiConfig struct{
	fileserverHits atomic.Int32
}
//helper functions
func respondWithJSON(w http.ResponseWriter, code int, payload interface{})error{
	data,err:=json.Marshal(payload)
	if err!=nil{
		return err 
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(data)
	return nil
}
func respondWithError(w http.ResponseWriter, code int, msg string)error{
	return respondWithJSON(w,code,map[string]string{"error":msg})
}
//handlers
func readinessHandler(w http.ResponseWriter,r *http.Request){
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) hitCountHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/html; charset=utf-8")
	body:= fmt.Sprintf(`
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>`,cfg.fileserverHits.Load())
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
//bad word replacement
func badWordReplace(ourString string) string{
	badWords:=[]string {"kerfuffle","sharbert","fornax"}
	newString:=[]string {}
	replacedWord:="****"
	ourSlice:=strings.Split(ourString," ")
	for _,word:= range ourSlice{
		wordLower:=strings.ToLower(word)
		replacementNeeded:=false
		for _,badWord:= range badWords{
			if wordLower==strings.ToLower(badWord){
				replacementNeeded=true
			}
		} 
		if replacementNeeded{
			newString=append(newString,replacedWord)
		}else{
			newString=append(newString,word)
		}
	}
	finalString:=strings.Join(newString," ")
	return finalString
}
func validateChirpHandler(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Body string `json:"body"`
	} 
	type errorResponse struct{
		Error string `json:"error"`
	}
	type validResponse struct{
		Cleaned_body string `json:"cleaned_body"`
	}
	decoder:=json.NewDecoder(r.Body)
	params:=parameters{}
	err:= decoder.Decode(&params)
	if err!=nil{
		ourError:=errorResponse{
			Error:"Something went wrong",
		}
		data,_:=json.Marshal(ourError)
		w.Header().Set("Content-Type","application/json")	
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	//chirp too long
	if len(params.Body)>140{
		ourError:=errorResponse{
			Error:"chirp is too long",
		}
		data,_:=json.Marshal(ourError)
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}
	//chirp is valid
	ourResponse:=validResponse{
		Cleaned_body: badWordReplace(params.Body),
	}
	data,_:=json.Marshal(ourResponse)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return

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
	//non website handlers
	mux.Handle("GET /api/healthz",http.HandlerFunc(readinessHandler))
	mux.Handle("POST /api/validate_chirp",http.HandlerFunc(validateChirpHandler))
	mux.Handle("GET /admin/metrics",http.HandlerFunc(apiCfg.hitCountHandler))
	mux.Handle("POST /admin/reset",http.HandlerFunc(apiCfg.resetHitHandler))
	

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