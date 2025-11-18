package auth_test

import(
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/BhavyaV29/chirpy/internal/auth"
	"net/http"
)
//tests for auth creation and validation funcs
func TestMakeAndValidateJWT_HappyPath(t *testing.T){
	userID:=uuid.New()
	secret:="hahahha"
	expiresIn:=time.Minute
	token,err:=auth.MakeJWT(userID,secret,expiresIn)
	if err!=nil{
		t.Fatalf("Make JWT error: %v",err)
	}

	gotUserID,err:=auth.ValidateJWT(token,secret)
	if err!=nil{
		t.Fatalf("Validate JWT error: %v",err)
	}

	if userID!=gotUserID{
		t.Errorf("expected %s, got %s", userID,gotUserID)
	}

}
func TestMakeAndValidateJWT_ExpiredTokenPath(t *testing.T){
	userID:=uuid.New()
	secret:="hahahaha"
	expiresIn:=-time.Minute
	token,err:=auth.MakeJWT(userID,secret,expiresIn)
	if err!=nil{
		t.Fatalf("Make JWT returned error with negative expiresIn(should still have created the token): %v",err)
	}

	_,err=auth.ValidateJWT(token,secret)
	if err==nil{
		t.Fatalf("expected error for expired token, got nil")
	}
}

func TestMakeAndValidateJWT_WrongSecret(t *testing.T){
	userID:=uuid.New()
	goodSecret:="alpha"
	badSecret:="beta"
	expiresIn:=time.Minute
	token,err:=auth.MakeJWT(userID,goodSecret,expiresIn)
	if err!=nil{
		t.Fatalf("Make JWT error: %v",err)
	}

	_,err=auth.ValidateJWT(token,badSecret)
	if err==nil{
		t.Fatalf("Validation should have failed due to different HMAC keys but it did not")
	}
}


//tests for getbearertoken func

func TestGetBearerToken_Valid(t *testing.T){
	h:=http.Header{}
	h.Set("Authorization","Bearer token")
	token,err:=auth.GetBearerToken(h)
	if err!=nil{
		t.Fatalf("valid token string and header still got an error: %v",err)
	}
	if token!="token"{
		t.Errorf("expected %s got %s", "token", token)
	}
}
func TestGetBearerToken_MissingHeader(t *testing.T){
	h:=http.Header{}
	_,err:=auth.GetBearerToken(h)
	if err==nil{
		t.Fatalf("missing header should have returned error")
	}
}
func TestGetBearerToken_EmptyValue(t *testing.T){
	h:=http.Header{}
	h.Set("Authorization","")
	_,err:=auth.GetBearerToken(h)
	if err==nil{
		t.Fatalf("empty string value to authorization header should have returned error")
	}
}
func TestGetBearerToken_WrongScheme(t *testing.T){
	h:=http.Header{}
	h.Set("Authorization","Basic aabsb")
	_,err:=auth.GetBearerToken(h)
	if err==nil{
		t.Fatalf("bearer prefix not present should have returned error")
	}
}
func TestGetBearerToken_EmptyToken(t *testing.T){
	h:=http.Header{}
	h.Set("Authorization","Bearer ")
	_,err:=auth.GetBearerToken(h)
	if err==nil{
		t.Fatalf("empty token string should have returned an error")
	}
}