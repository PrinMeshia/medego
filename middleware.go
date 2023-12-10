package medego

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)

func (c *Medego) SessionLoad(next http.Handler) http.Handler {
	return c.Session.LoadAndSave(next)
}

func (c *Medego) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(c.Config.Cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   c.Config.Cookie.domain,
	})
	return csrfHandler
}
