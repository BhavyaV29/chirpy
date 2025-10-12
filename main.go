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
	"log"
	"errors"
	
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
	Token 	  string    `json:"token"`
	RefreshToken string `json:"refresh_token"`
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
//user update handler
func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request){
	type Body struct{
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type UserResponse struct{
		ID uuid.UUID `json:"id"`
		Email string `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	//authenticating jwt
	accessTokenReceived,err:=auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w,401,"malformed or missing access token")
		return
	}
	userID,err:=auth.ValidateJWT(accessTokenReceived,cfg.jwtSecret)
	if err!=nil{
		respondWithError(w,401,"Invalid token")
		return
	}
	//decoding request body
	decoder:=json.NewDecoder(r.Body)
	requestBody:=Body{}
	err=decoder.Decode(&requestBody)
	if err!=nil{
		respondWithError(w,400,"Bad JSON/body missing fields")
		return
	}
	if requestBody.Email=="" || requestBody.Password==""{
		respondWithError(w,400,"email and password required")
		return
	}
	//hashing new password and setting both pass and  email
	newPassword,err:=auth.HashPassword(requestBody.Password)
	if err!=nil{
		respondWithError(w,500,"hashing/db failed")
		return
	}
	//putting the new data in the database
	userNewParams:=database.NewPasswordEmailParams{
		Email: requestBody.Email,
		HashedPassword: newPassword,
		ID: userID,
	}
	updatedUser,err:=cfg.db.NewPasswordEmail(r.Context(),userNewParams)
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			respondWithError(w,404,"user not found")
			return
		}
		respondWithError(w,500,"database error")
		return
	}
	responseUser:=UserResponse{
		ID:updatedUser.ID,
		Email:updatedUser.Email,
		CreatedAt:updatedUser.CreatedAt,
		UpdatedAt:updatedUser.UpdatedAt,
	}
	respondWithJSON(w,200,responseUser)
}

//user login handler
func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request){
	
	type Body struct{
		Password string `json:"password"`
		Email string `json:"email"`
		//ExpiresInSeconds *int64 `json:"expires_in_seconds"` 
	}
	
	//decoding our login request
	decoder:=json.NewDecoder(r.Body)
	ourBody:=Body{}
	err:=decoder.Decode(&ourBody)
	if err!=nil{
		respondWithError(w,400,"Request body wrong")
		return
	}
	//setting expiresinseconds
	max:=int64(time.Hour.Seconds())
	/*var secs int64
	if ourBody.ExpiresInSeconds==nil || *ourBody.ExpiresInSeconds<=0{
		secs=max
	} else if *ourBody.ExpiresInSeconds>max{
		secs=max
	} else{
		secs=*ourBody.ExpiresInSeconds
	}*/
	

	user,err:= cfg.db.GetUserByEmail(r.Context(),ourBody.Email)
	if err!=nil{
		respondWithError(w,401,"Incorrect email or password")
		return
	}
	correctPass,err:= auth.CheckPasswordHash(user.HashedPassword,ourBody.Password)
	if err!= nil || correctPass!=true{
		respondWithError(w,401,"Incorrect email or password")
		return
	}

	refresh_token,err:=auth.MakeRefreshToken()
	if err!=nil{
		respondWithError(w,401,"failed to create refresh token")
	}
	refreshTokenParams:=database.CreateRefreshTokenParams{
		Token: refresh_token,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Hour*24 * 60),
		RevokedAt: sql.NullTime{Valid:false},
	}

	refreshToken,err:=cfg.db.CreateRefreshToken(r.Context(),refreshTokenParams)
	if err!=nil{
		respondWithError(w,500,"database error creating refresh token")
	}
	exp:=time.Duration(max)*time.Second
	token,err:=auth.MakeJWT(user.ID,cfg.jwtSecret,exp)
	if err!=nil{
		respondWithError(w,401,"JWT token generation failed")
		return
	}


	ourUser:=User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token, 
		RefreshToken: refreshToken.Token,
	}
	respondWithJSON(w,200,ourUser)
	

}

//refresh handler to make new access token

func(cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request){
	refreshToken,err:=auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w,401,"invalid or missing refresh token")
		return
	}
	userID,err:=cfg.db.GetUserFromRefreshToken(r.Context(),refreshToken)
	
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			respondWithError(w,401,"invalid,expired or revoked refresh token")
			return
		}
		respondWithError(w,500,"database error")
		return
	}
	//1 hour expiry
	
	expiresIn:=1*time.Hour
	newAccessToken,err:=auth.MakeJWT(userID,cfg.jwtSecret,expiresIn)
	if err!=nil{
		respondWithError(w,500,"new access token creation failed")
		return
	}
	respondWithJSON(w,200,map[string]string{"token":newAccessToken})
}
//revoke Handler to revoke refresh token
func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request){
	refreshToken,err:=auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w,401,"invalid or missing refresh token in header")
	}

	_,err=cfg.db.RevokeRefreshToken(r.Context(),refreshToken)
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			respondWithError(w,401,"invalid token")
			return
		}
		respondWithError(w,500,"database error")
		return
	}
	
	w.WriteHeader(http.StatusNoContent) //204, no body

}
//chirp creation handler

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request){
	type inputChirp struct{
		Body string `json:"body"`
	}
	decoder:=json.NewDecoder(r.Body)
	ourChirp:=inputChirp{}
	err:=decoder.Decode(&ourChirp)
	if err!=nil{
		respondWithError(w,400,"error decoding JSON")
		return
	}

	tokenStr,err:=auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w,401,"invalid or missing token")
		return
	}
	userID,err:=auth.ValidateJWT(tokenStr,cfg.jwtSecret)
	if err!=nil{
		respondWithError(w,401,"Unauthorized")
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
	/*uid,err:= uuid.Parse(ourChirp.UserID)
	if err!=nil{
		respondWithError(w,400,"invalid user id")
		return
	}*/
	/*nullUID:=uuid.NullUUID{
		UUID:uid,
		Valid:true,
	}*/

	//inserting in chirps db
	params:= database.CreateChirpParams{
		Body: ourChirp.Body,
		UserID: userID,
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
//delete chirp handler
func(cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request){
	//authentication
	token,err:=auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w,401,"malformed or missing token")
		return
	}
	userID,err:=auth.ValidateJWT(token,cfg.jwtSecret)
	if err!=nil{
		respondWithError(w,401,"invalid token")
		return
	}
	//extract chirp id from path
	receivedChirpID:=r.PathValue("chirpID")
	if len(receivedChirpID)==0{
		respondWithError(w,400,"path value doesn't match")
		return
	}
	chirpID,err:=uuid.Parse(receivedChirpID)
	if err!=nil{
		respondWithError(w,400,"invalid chirp ID")
		return
	}
	//see if chirp exists
	dbChirp,err:=cfg.db.GetSingleChirp(r.Context(),chirpID)
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			respondWithError(w,404,"chirp not found")
			return
		}
		respondWithError(w,500,"db error")
		return
	}
	//see if authenticated user is owner of the chirp
	if userID!=dbChirp.UserID{
		respondWithError(w,403,"user not author of the chirp")
		return
	}
	//delete Chirp
	err=cfg.db.DeleteChirp(r.Context(),chirpID)
	if err!=nil{
		respondWithError(w,500,"db error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
		log.Fatal("opening DB: %v", err)
	}
	queries:=database.New(db)

	//making our server apiConfig instance
	apiCfg:=&apiConfig{
		db : queries,
		platform : platformVal,
		jwtSecret: jwtSecret,

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
	mux.Handle("PUT /api/users",http.HandlerFunc(apiCfg.updateUserHandler))
	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.chirpsHandler))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.getChirpsHandler))
	mux.Handle("GET /api/chirps/{chirpID}",http.HandlerFunc(apiCfg.getSingleChirpHandler))
	mux.Handle("DELETE /api/chirps/{chirpID}",http.HandlerFunc(apiCfg.deleteChirpHandler))
	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.loginHandler))
	mux.Handle("POST /api/refresh",http.HandlerFunc(apiCfg.refreshHandler))
	mux.Handle("POST /api/revoke", http.HandlerFunc(apiCfg.revokeHandler))
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