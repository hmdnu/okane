package lib

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/hmdnu/okane/config"
)

var SessionStore = sessions.NewCookieStore([]byte(config.SESSION_SECRET))

const (
	SESSION_EXPIRES_24H        = 60 * 60 * 24
	SESSION_EXPIRES_REMEMBERED = SESSION_EXPIRES_24H * 30
)

func init() {
	gob.Register([]FormError{})
	gob.Register(FormError{})
	gob.Register(uint(0))
	gob.Register("")

	SessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   SESSION_EXPIRES_24H,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func SetSession(w http.ResponseWriter, r *http.Request, sessionName string, key string, value any) error {
	return SetSessionWithMaxAge(w, r, sessionName, key, value, SESSION_EXPIRES_24H)
}

func SetSessionWithMaxAge(w http.ResponseWriter, r *http.Request, sessionName string, key string, value any, maxAge int) error {
	session, err := SessionStore.Get(r, sessionName)

	if err != nil {
		return err
	}

	session.Options.MaxAge = maxAge
	session.Values[key] = value
	return SessionStore.Save(r, w, session)
}

func GetSession(r *http.Request, sessionName string, key string) (any, bool) {
	session, err := SessionStore.Get(r, sessionName)

	if err != nil {
		return nil, false
	}

	data, ok := session.Values[key]
	return data, ok
}

func ClearSession(w http.ResponseWriter, r *http.Request, sessionName string) error {
	session, err := SessionStore.Get(r, sessionName)

	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	return SessionStore.Save(r, w, session)
}
