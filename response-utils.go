package medego

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
)

func (c *Medego) setHeaders(w http.ResponseWriter, status int, contentType string, headers ...http.Header) {
	// Définition des en-têtes supplémentaires
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	// Écriture des en-têtes standard
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
}

func (c *Medego) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(data); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only have a single json value")
	}
	return nil
}

func (c *Medego) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	c.setHeaders(w, status, "application/json", headers...)
	_, err = w.Write(out)
	return err
}

func (c *Medego) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	c.setHeaders(w, status, "application/xml", headers...)
	_, err = w.Write(out)
	return err
}

func (c *Medego) DownloadFile(w http.ResponseWriter, r *http.Request, pathTolFile, fileName string) error {
	fp := path.Join(pathTolFile, fileName)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (c *Medego) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (c *Medego) Error404(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusNotFound)
}
func (c *Medego) Error500(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusInternalServerError)
}
func (c *Medego) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusUnauthorized)
}
func (c *Medego) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	c.ErrorStatus(w, http.StatusForbidden)
}
