package web

import (
	"html/template"
	"net/http"
	_ "net/http/pprof"

	"github.com/ascii-arcade/knucklebones/config"
	"github.com/ascii-arcade/knucklebones/games"
	"github.com/ascii-arcade/knucklebones/players"
)

func Run() error {
	mux := http.NewServeMux()

	if config.Debug {
		mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/goroutine", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/threadcreate", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/block", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/all", http.DefaultServeMux.ServeHTTP)
	}

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/favicon.ico")
	})

	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web"+r.URL.Path)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Eventually implement admin only information
		// params := r.URL.Query()
		// if params.Get("admin_key") == config.GetWebAdminKey() {}
		totalGames := len(games.GetAll())
		totalStartedGames := 0
		for _, game := range games.GetAll() {
			if game.InProgress {
				totalStartedGames++
			}
		}

		t, err := template.ParseFiles("web/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = t.Execute(w, struct {
			TotalGames            int
			TotalStartedGames     int
			TotalUniquePlayers    int
			TotalConnectedPlayers int
		}{
			TotalGames:            totalGames,
			TotalStartedGames:     totalStartedGames,
			TotalUniquePlayers:    players.GetUniquePlayerCount(),
			TotalConnectedPlayers: players.GetConnectedPlayerCount(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return http.ListenAndServe(":"+config.HTTPPort, mux)
}
