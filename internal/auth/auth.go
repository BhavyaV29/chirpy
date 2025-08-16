package auth 

import(
	"golang.org/x/crypto/bcrypt"

)

func HashPassword(password string) (string,error){
	passwordConverted:= []byte(password)
	hashBytes,err:=bcrypt.GenerateFromPassword(passwordConverted,bcrypt.DefaultCost)
	if err!=nil{
		return "",err
	}
	hash:=string(hashBytes)
	return hash,nil
}

func CheckPasswordHash(hash,password string) error{
	hashBytes:= []byte(hash)
	passwordBytes:=[]byte(password)
	return bcrypt.CompareHashAndPassword(hashBytes,passwordBytes)
}
