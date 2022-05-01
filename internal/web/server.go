package web

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/wesleywxie/gogetit/internal/config"
	"github.com/wesleywxie/gogetit/internal/model"
	"go.uber.org/zap"
	"net/http"
)

type application struct {
	auth struct {
		username string
		password string
	}
}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (app *application) torrentsHandler(w http.ResponseWriter, r *http.Request) {

	update := false
	param, ok := r.URL.Query()["complete"]
	if ok && len(param[0]) > 0 {
		update = true
	}

	w.Header().Set("Content-Type", "application/json")
	torrents := model.FindAndUpdateSelectedTorrents(update)
	var response []string
	for _, t := range torrents {
		response = append(response, t.MagnetLink)
	}
	json.NewEncoder(w).Encode(response)
}

func Start(port int) {
	app := new(application)

	app.auth.username = config.Username
	app.auth.password = config.Password

	http.HandleFunc("/torrents", app.basicAuth(app.torrentsHandler)) // Update this line of code

	fmt.Printf("Starting server at port %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil); err != nil {
		zap.S().Fatal(err)
	}
}
