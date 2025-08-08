package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"curltree/internal/auth"
	"curltree/internal/config"
	"curltree/internal/database"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewSQLiteDB(cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	authService := auth.NewAuthService(db)

	sshAddr := fmt.Sprintf("%s:%d", cfg.SSH.Host, cfg.SSH.Port)
	
	s, err := wish.NewServer(
		wish.WithAddress(sshAddr),
		wish.WithHostKeyPath(cfg.SSH.HostKeyPath),
		wish.WithMiddleware(
			authService.Middleware(),
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return newTUIModel(s, db), []tea.ProgramOption{tea.WithAltScreen()}
			}),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting SSH server on %s", sshAddr)
	log.Printf("Database: %s (%s)", cfg.GetDatabaseURL(), cfg.Database.Type)
	log.Printf("Host key: %s", cfg.SSH.HostKeyPath)
	
	go func() {
		if err = s.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && err != ssh.ErrServerClosed {
		log.Fatalf("Could not stop server: %v", err)
	}
}

