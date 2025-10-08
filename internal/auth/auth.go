package auth 

import(
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"fmt"
	"time"
	"github.com/google/uuid"
	"strings"
	"net/http"
)
//hashing password while user creation func
func HashPassword(password string) (string,error){
	
	hash,err:=argon2id.CreateHash(password,argon2id.DefaultParams)
	if err!=nil{
		return "",err
	}

	return hash,nil
}
//verifying correct password func
func CheckPasswordHash(hash,password string)(bool, error){
	

	return argon2id.ComparePasswordAndHash(password,hash)
}
//jwt construction function
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration)(string,error){
	now:=time.Now().UTC()
	expiryTime:=now.Add(expiresIn)
	claims:=&jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiryTime),
		Subject: userID.String(),
	}
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	key:=[]byte(tokenSecret)
	signedString,err:=token.SignedString(key)
	if err!=nil{
		return "",err
	}
	return signedString,nil
}
//validating jwt function
func ValidateJWT(tokenString,tokenSecret string)(uuid.UUID, error){
	claims:=&jwt.RegisteredClaims{}
	token,err:=jwt.ParseWithClaims(tokenString,claims,func(t *jwt.Token)(any,error){if _,correctMethod:=t.Method.(*jwt.SigningMethodHMAC);!correctMethod{
		return nil,fmt.Errorf("unexpected signing method")}
	return []byte(tokenSecret),nil
	})
	if err!=nil{
		return uuid.Nil,err
	}
	if !token.Valid{
		return uuid.Nil,fmt.Errorf("invalid token")
	}
	subject:=claims.Subject
	id,err:=uuid.Parse(subject)
	if err!=nil{
		return uuid.Nil,err
	}	
	
	return id,nil


}

func GetBearerToken(headers http.Header)(string,error){
	stringValue:=headers.Get("Authorization")
	if stringValue==""{
		return "",fmt.Errorf("authorization header missing")
	}
	prefix:="Bearer "
	if !strings.HasPrefix(stringValue,prefix){
		return "",fmt.Errorf("invalid authorization scheme")
	}
	stringPrefixTrmmed:=strings.TrimPrefix(stringValue,prefix)
	tokenString:=strings.TrimSpace(stringPrefixTrmmed)
	if tokenString==""{
		return "",fmt.Errorf("empty token string")
	}
	return tokenString,nil

}

