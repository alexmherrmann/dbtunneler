package webmonitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"tunny/monitor"
)

type WebMonitor struct {
	lock    *sync.Mutex
	Targets []monitor.TunnelTarget `json:"targets"`

	Info  map[string][]string `json:"info"`
	Warn  map[string][]string `json:"warn"`
	Fatal map[string][]string `json:"fatal"`

	GeneralInfo []string `json:"general_info"`

	cliBackup *monitor.CliMonitor
}

func (w *WebMonitor) claim() func() {
	w.lock.Lock()

	return func() {
		w.lock.Unlock()
	}
}

// ReportError implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportError(targetName string, _fmt string, args ...any) {
	w.cliBackup.ReportError(targetName, _fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	if val, ok := w.Warn[targetName]; ok {
		w.Warn[targetName] = append(val, fmt.Sprintf(_fmt, args...))
	} else {
		// Just go make it
		w.Warn[targetName] = []string{fmt.Sprintf(_fmt, args...)}
	}
}

// ReportFatalError implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportFatalError(targetName string, _fmt string, args ...any) {
	w.cliBackup.ReportFatalError(targetName, _fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	if val, ok := w.Fatal[targetName]; ok {
		w.Fatal[targetName] = append(val, fmt.Sprintf(_fmt, args...))
	} else {
		// Just go make it
		w.Fatal[targetName] = []string{fmt.Sprintf(_fmt, args...)}
	}
}

// ReportGeneralMessage implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportGeneralMessage(_fmt string, args ...any) {
	w.cliBackup.ReportGeneralMessage(_fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	w.GeneralInfo = append(w.GeneralInfo, fmt.Sprintf(_fmt, args...))
}

// ReportInfo implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportInfo(targetName string, _fmt string, args ...any) {
	w.cliBackup.ReportInfo(targetName, _fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	if val, ok := w.Info[targetName]; ok {
		w.Info[targetName] = append(val, fmt.Sprintf(_fmt, args...))
	} else {
		// Just go make it
		w.Info[targetName] = []string{fmt.Sprintf(_fmt, args...)}
	}
}

func NewWebMonitor(ctx context.Context, targets []monitor.TunnelTarget) (*WebMonitor, error) {
	mon := &WebMonitor{
		lock:        &sync.Mutex{},
		Targets:     targets,
		Info:        make(map[string][]string),
		Warn:        make(map[string][]string),
		Fatal:       make(map[string][]string),
		GeneralInfo: make([]string, 0),
	}

	mux := mon.GetMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func(ctx context.Context) {
		<-ctx.Done()
		log.Printf("Received shutdown signal, shutting down web server")
		newCtx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFunc()
		err := server.Shutdown(newCtx)
		if err != nil {
			fmt.Printf("Error shutting down web server: %s", err)
		}
	}(ctx)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("%s", err)
		}
	}()

	return mon, nil
}
