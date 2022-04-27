package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

var key = []byte("top-secret")

func Sign(key, msg []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)

	return hex.EncodeToString(append(msg, mac.Sum(nil)...))
}

func Verify(key []byte, hash string) bool {
	sig, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(sig[:4])

	return hmac.Equal(sig[4:], mac.Sum(nil))
}

func Check() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, err := c.Request().Cookie("url-auth")
			if err == http.ErrNoCookie || !Verify(key, auth.Value) {
				token := make([]byte, 4)
				if _, err := rand.Read(token); err != nil {
					c.Logger().Error(err)
				}

				cookie := new(http.Cookie)
				cookie.Name = "url-auth"
				cookie.Value = Sign(key, token)
				cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
				c.SetCookie(cookie)
				return next(c)
			}

			return next(c)
		}
	}
}
