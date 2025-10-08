package main

import (
	_ "github.com/lib/pq"
	"net/http"
	"fmt"
	"sync/atomic"
	"encoding/json"
	"strings"
	"github.com/BhavyaV29/chirpy/internal/database"
	"github.com/google/uuid"
	"time"
	"github.com/joho/godotenv"
	"os"
	"database/sql"
	"github.com/BhavyaV29/chirpy/internal/auth"
	
)

type apiConfig struct{
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	jwtSecret string
}

//our user struct
type User struct{
	ID 		  uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email 	  string	`json:"email"`
}
type Chirp struct{
	ID 			uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
	Body 		string 	  `json:"body"`
	UserID 		uuid.UUID `json:"user_id"`

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
	if cfg.platform=="dev"{
		err:=cfg.db.DeleteUsers(r.Context())
		if err!=nil{
			respondWithError(w,500,"failed to delete db")
			return
		}
	}else{
		respondWithError(w,403,"403 Forbidden")
		return
	}
	w.Write([]byte("reset\n"))
	return
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


//user creation handler
func (cfg *apiConfig) CreateUserHandler( w http.ResponseWriter,r *http.Request){
	type Body struct{
		Password string `json:"password"`
		Email string `json:"email"`

	}
	//decoding email and handling error
	decoder:=json.NewDecoder(r.Body)
	ourBody:= Body{}
	err:= decoder.Decode(&ourBody)
	if err !=nil{
		respondWithError(w,400,"Body format wrong")
		return
	}

	//hashing password
	hashedPassword,err:= auth.HashPassword(ourBody.Password)
	if err!=nil{
		respondWithError(w,500,"hashing failed")
		return 
	}
	//creating user using email
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: ourBody.Email,
		HashedPassword: hashedPassword,
	})
	if err!= nil{
		respondWithError(w,500,"User not created")
		return
	}

	ourUser:=User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJSON(w,201,ourUser)
	return

}
//user login handler
func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request){
	type Body struct{
		Password string `json:"password"`
		Email string `json:"email"`
	}
	//decoding our login request
	decoder:=json.NewDecoder(r.Body)
	ourBody:=Body{}
	err:=decoder.Decode(&ourBody)
	if err!=nil{
		respondWithError(w,400,"Request body wrong")
		return
	}
	user,err:= cfg.db.GetUserByEmail(r.Context(),ourBody.Email)
	if err!=nil{
		respondWithError(w,401,"Incorrect email or password")
		return
	}
	correctPass,err:= auth.CheckPasswordHash(user.HashedPassword,ourBody.Password)
	if err!= nil || correctPass!=true{
		respondWithError(w,401,"Incorrect email password")
		return
	}
	ourUser:=User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJSON(w,200,ourUser)
	

}

//chirp creation handler

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request){
	type inputChirp struct{
		Body string `json:"body"`
		UserID string `json:"user_id"`
	}
	decoder:=json.NewDecoder(r.Body)
	ourChirp:=inputChirp{}
	err:=decoder.Decode(&ourChirp)
	if err!=nil{
		respondWithError(w,400,"error decoding JSON")
		return
	}

	//chirp too long
	if len(ourChirp.Body)>140{
		respondWithError(w,400,"chirp is too long")
		return
	}
	//chirp is valid
	ourChirp.Body=badWordReplace(ourChirp.Body)

	//parse user ID string into uuid.UUID
	uid,err:= uuid.Parse(ourChirp.UserID)
	if err!=nil{
		respondWithError(w,400,"invalid user id")
		return
	}
	/*nullUID:=uuid.NullUUID{
		UUID:uid,
		Valid:true,
	}*/

	//inserting in chirps db
	params:= database.CreateChirpParams{
		Body: ourChirp.Body,
		UserID: uid,
	}

	createdChirp,err:= cfg.db.CreateChirp(r.Context(),params)
	if err!=nil{
		
		respondWithError(w,500,"database error creating chirp")
		return
	}
	resp := Chirp{
		ID:        createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body:      createdChirp.Body,
		UserID:    createdChirp.UserID, // or just createdChirp.UserID if it's not a NullUUID
	}
	respondWithJSON(w, 201, resp)
	
	return 

	
	
}

func (cfg *apiConfig) getChirpsHandler (w http.ResponseWriter, r *http.Request){
	dbChirps,err:=cfg.db.GetChirps(r.Context())
	if err!=nil{
		respondWithError(w,500,"Failed to retrive chirps from db")
		return
	}
	listChirps:=[]Chirp{}
	for _,dbChirp :=range dbChirps{
		singleChirp:=Chirp{
			ID: dbChirp.ID,
			Body: dbChirp.Body,
			UserID: dbChirp.UserID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
		}
		listChirps=append(listChirps,singleChirp)

	}
	respondWithJSON(w,200,listChirps)
	return
}

func (cfg *apiConfig) getSingleChirpHandler (w http.ResponseWriter, r *http.Request){
	chirpID:= r.PathValue("chirpID")
	if len(chirpID)==0{
		respondWithError(w,404,"path value doesn't match")
	}
	uid,err:= uuid.Parse(chirpID)
	if err!=nil{
		respondWithError(w,400,"invalid user id")
		return
	}
	dbChirp,err:=cfg.db.GetSingleChirp(r.Context(),uid)
	if err!=nil{
		if err==sql.ErrNoRows{
			respondWithError(w,404,"ID not found")
			return
		}else{
			respondWithError(w,500,"Database error")
		}
	}
	responseChirp:=Chirp{
		ID : dbChirp.ID,
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,

	}
	respondWithJSON(w,200,responseChirp)
}

func main(){
	
	//loading db,platform,jwt envs
	godotenv.Load()
	dbURL:=os.Getenv("DB_URL")
	if dbURL==""{
		log.Fatal("DB_URL must be set")
	}

	platformVal:=os.Getenv("PLATFORM")
	if platformVal==""{
		log.Fatal("PLATFORM must be set" )
	}

	jwtSecret:=os.Getenv("JWT_SECRET")
	if jwtSecret==""{
		log.Fatal("JWT_SECRET must be set")
	}
	//open db
	db,err := sql.Open("postgres",dbURL)
	if err!=nil{
		log.Fatal("DB opening failed")
	}
	queries:=database.New(db)

	//making our server apiConfig instance
	apiCfg:=&apiConfig{
		db : queries,
		platform : platformVal,
		jwtSecret: jwtSecret

	}
	
	//making router
	mux := http.NewServeMux()

	//defining handlers
	fileserver:=http.FileServer(http.Dir("."))
	appHandler:=http.StripPrefix("/app",fileserver) //stripping path
	mux.Handle("/app/",apiCfg.middlewareMetricsInc(appHandler))
	//non website handlers
	mux.Handle("GET /api/healthz",http.HandlerFunc(readinessHandler))
	mux.Handle("POST /api/users",http.HandlerFunc(apiCfg.CreateUserHandler))
	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.chirpsHandler))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.getChirpsHandler))
	mux.Handle("GET /api/chirps/{chirpID}",http.HandlerFunc(apiCfg.getSingleChirpHandler))
	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.loginHandler))
	mux.Handle("GET /admin/metrics",http.HandlerFunc(apiCfg.hitCountHandler))
	mux.Handle("POST /admin/reset",http.HandlerFunc(apiCfg.resetHitHandler))
	
	

	//setting up server
	server:= &http.Server{
		Handler:mux,
		Addr:":8080",
	}

	//running the server
	err=server.ListenAndServe()
	if err!= nil{
		fmt.Println(err)
	}
	

}