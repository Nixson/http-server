package http_server

import (
	"encoding/json"
	"fmt"
	annotation "github.com/Nixson/annotation/method"
	"github.com/Nixson/environment"
	"github.com/Nixson/http-server/session"
	"github.com/Nixson/logNx"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type ContextInterface interface {
	SetContext(Context)
}

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Session  session.Session
	Path     string
}
type Params struct {
	Annotation *annotation.Annotation
	Env        *environment.Env
}

var params *Params

func (c *Context) Access(access uint) bool {
	return access <= c.Session.Access
}
func getInfo(path string) info {
	nf, ok := method[path]
	if ok {
		return nf
	}
	pathList := strings.Split(path, "/")
	if len(pathList) < 1 {
		return info{
			access: "auth",
		}
	}
	pathList = pathList[:len(pathList)-1]
	return getInfo(strings.Join(pathList, "/"))
}
func (c *Context) IsGranted() bool {
	inf := getInfo(c.Path)
	if inf.access == "all" {
		return true
	}
	if params.Env.GetBool("security.enable") {
		tokenKey, err := JWTTokenOpenKey()
		if err != nil {
			c.Response.WriteHeader(500)
			c.Error(http.StatusBadRequest, fmt.Sprintf("Failed get jwt open key, '%s'", err))
			return false
		}
		if len(c.Request.Header["Authorization"]) < 1 {
			c.Error(http.StatusUnauthorized, "failed header Authorization")
			return false
		}
		authorizationHeader := c.Request.Header["Authorization"][0]
		token := strings.TrimPrefix(authorizationHeader, "Bearer ")
		ok, exception := verifyToken(token, *tokenKey)
		if ok {
			c.Session = session.Session{Access: 1}
			return true
		}
		c.Error(http.StatusBadRequest, exception.Marshal())
		return false
	}
	c.Session = session.Session{Access: 1}
	return true
}

type info struct {
	index  int
	access string
	handle *ContextInterface
}

var method = make(map[string]info)

func InitController(name string, controller *ContextInterface) {
	annotationList := params.Annotation.Get("Controller")
	var annotationMap map[string]annotation.Element
	for _, annotationMapEl := range annotationList {
		if annotationMapEl.StructName == name {
			for _, child := range annotationMapEl.Children {
				annotationMap[child.StructName] = child
			}
		}
	}
	_struct := reflect.TypeOf(controller)
	for index := 0; index < _struct.NumMethod(); index++ {
		_method := _struct.Method(index)
		annotationMapEl, ok := annotationMap[_method.Name]
		if !ok {
			continue
		}
		access, ok := annotationMapEl.Parameters["access"]
		if !ok {
			access = "auth"
		}
		inf := info{
			index:  _method.Index,
			access: access,
			handle: controller,
		}
		path, hasPath := annotationMapEl.Parameters["path"]
		if !hasPath {
			path = _method.Name
		}
		method[path] = inf
	}
}

func (c *Context) Call() {
	in := make([]reflect.Value, 0)
	info, ok := method[c.Path]
	if !ok {
		return
	}
	reflect.ValueOf(info.handle).Method(info.index).Call(in)
}
func (c *Context) Write(iface interface{}) {
	marshal, _ := json.Marshal(iface)
	_, err := c.Response.Write(marshal)
	if err != nil {
		logNx.GetLogger().Error(err.Error())
	}
}
func (c *Context) Error(status int, iface interface{}) {
	c.Response.WriteHeader(status)
	c.Write(iface)
}

type TokenException struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (t *TokenException) Marshal() string {
	marshal, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func verifyToken(tokenString, tokenOpenKey string) (bool, *TokenException) {
	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(tokenOpenKey))
	if err != nil {
		return false, &TokenException{
			Error:            "failed_public_key",
			ErrorDescription: "Failed parsing JWT public key (from security service)",
		}
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return false, &TokenException{
			Error:            "invalid_token",
			ErrorDescription: "Cannot convert access token to JSON",
		}
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false, &TokenException{
			Error:            "invalid_token",
			ErrorDescription: "Cannot convert access token to JSON",
		}
	}

	err = jwt.SigningMethodRS256.Verify(strings.Join(parts[0:2], "."), parts[2], key)
	if err != nil {
		return false, &TokenException{
			Error:            "invalid_token",
			ErrorDescription: "Failed verifying token",
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, &TokenException{
			Error:            "invalid_token",
			ErrorDescription: "Cannot convert access token to JSON",
		}
	}

	var tm time.Time
	switch iat := claims["exp"].(type) {
	case float64:
		tm = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		tm = time.Unix(v, 0)
	}

	if tm.Before(time.Now()) {
		return false, &TokenException{
			Error:            "invalid_token",
			ErrorDescription: "Access token expired: " + tokenString,
		}
	}

	return true, nil
}
