package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"github.com/wesleywxie/gogetit/internal/config"
	"github.com/wesleywxie/gogetit/internal/model"
	"github.com/wesleywxie/gogetit/internal/task"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	zap.S().Debug("Initialization main module...")
}

type application struct {
	auth struct {
		username string
		password string
	}
}

func main() {
	model.InitDB()
	defer model.Disconnect()
	task.StartTasks()

	go handleSignal()
	startWebServer()
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-c

	task.StopTasks()
	model.Disconnect()
	os.Exit(0)
}

func startWebServer() {
	app := new(application)

	app.auth.username = config.Username
	app.auth.password = config.Password

	http.HandleFunc("/torrents", app.basicAuth(torrentsHandler)) // Update this line of code

	fmt.Printf("Starting server at port 31065\n")
	if err := http.ListenAndServe("127.0.0.1:31065", nil); err != nil {
		zap.S().Fatal(err)
	}
}

func torrentsHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./static/tmpl/torrents.html"))
	torrents := model.FindAndUpdateSelectedTorrents()
	t.Execute(w, torrents)
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
