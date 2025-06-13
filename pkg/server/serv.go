package server

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var ErrAlreadyStopped = errors.New("server already stopped")

type Server interface {
	Run() error
	Kill() error
}

func Multi(servers ...Server) Server {
	return multiserv{servers: servers}
}

type multiserv struct {
	servers []Server
	killed  bool
}

func (m multiserv) Run() error {
	if m.killed {
		return ErrAlreadyStopped
	}
	slog.Info("Starting servers", "n", len(m.servers))
	m.registerSignals()
	errCh := make(chan error)
	var wg sync.WaitGroup
	for _, server := range m.servers {
		wg.Add(1)
		go func() {
			err := server.Run()
			errCh <- err
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()
	errs := make([]error, 0)
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
			if err = m.Kill(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (m multiserv) Kill() error {
	if m.killed {
		return nil
	}
	slog.Info("Shutting down servers", "n", len(m.servers))
	m.killed = true
	errs := make([]error, 0)
	for _, server := range m.servers {
		if err := server.Kill(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m multiserv) registerSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		slog.Info("Received signal", "signal", sig)
		if err := m.Kill(); err != nil {
			slog.Error("Shutdown error", "error", err)
		}
	}()
}
