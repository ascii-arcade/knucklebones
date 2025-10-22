package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ascii-arcade/knucklebones/app"
	"github.com/ascii-arcade/knucklebones/config"
	"github.com/ascii-arcade/knucklebones/database"
	"github.com/ascii-arcade/knucklebones/players"
	"github.com/ascii-arcade/knucklebones/web"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	gossh "golang.org/x/crypto/ssh"
)

func init() {
	slogLevel := slog.LevelInfo
	if config.Debug {
		slogLevel = slog.LevelDebug
	}

	handlerOpts := &slog.HandlerOptions{
		Level: slogLevel,
	}
	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	if config.Debug {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	slog.SetDefault(slog.New(handler).With("app", "knucklebones", "version", config.Version))
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := database.Setup(ctx, config.DatabaseURI, config.Database); err != nil {
		slog.Error("could not connect to database", "error", err)
		os.Exit(1)
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(config.Host, config.SSHPort)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			decodedKey := string(bytes.TrimSuffix(gossh.MarshalAuthorizedKey(key), []byte{'\n'}))
			slog.Debug("ssh authentication attempt", "user", ctx.User())

			player, found := players.Get(decodedKey)
			if !found {
				var err error
				if player, err = players.NewPlayer(ctx, "default", decodedKey, "en"); err != nil {
					slog.Error("could not create player", "error", err)
					return false
				}
				slog.Debug("created new player", "user", ctx.User())

				if ctx.User() == "web-client" {
					slog.Info("web-client ssh authentication successful", "user", ctx.User(), "player_id", player.Id)
					player.Visitor = true
				}
			}
			player.WithContext(ctx).Connect()

			ctx.SetValue("PLAYER", player)
			ctx.SetValue("PUBKEY", decodedKey)

			return true
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(app.TeaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not create wish server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("starting ssh server", "host", config.Host, "port", config.SSHPort, "version", config.Version)
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start wish server", "error", err)
			done <- nil
		}
	}()

	go func() {
		log.Info("starting http server", "host", config.Host, "port", config.HTTPPort, "version", config.Version)
		if err := web.Run(); err != nil {
			log.Error("could not start web server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("shutting down servers...")
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("error shutting down servers", "error", err)
	}
}
