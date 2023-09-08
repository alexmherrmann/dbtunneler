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

type LogItem struct {
	Level string    `json:"level",omitempty`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
}

type WebMonitor struct {
	lock    *sync.Mutex
	Targets []monitor.TunnelTarget `json:"targets"`

	Logs map[string][]LogItem `json:"logs"`
	// Info  map[string][]LogItem `json:"info"`
	// Warn  map[string][]LogItem `json:"warn"`
	// Fatal map[string][]LogItem `json:"fatal"`

	GeneralInfo []LogItem `json:"general_info"`

	cliBackup *monitor.CliMonitor
}

// Claim the webmonitor for writing, returns a function to release the lock.
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
	newItem := LogItem{Msg: fmt.Sprintf(_fmt, args...), Time: time.Now(), Level: "warn"}
	if val, ok := w.Logs[targetName]; ok {
		w.Logs[targetName] = append(val, newItem)
	} else {
		// Just go make it
		w.Logs[targetName] = []LogItem{newItem}
	}
}

// ReportFatalError implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportFatalError(targetName string, _fmt string, args ...any) {
	w.cliBackup.ReportFatalError(targetName, _fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	newItem := LogItem{Msg: fmt.Sprintf(_fmt, args...), Time: time.Now(), Level: "fatal"}
	if val, ok := w.Logs[targetName]; ok {
		w.Logs[targetName] = append(val, newItem)
	} else {
		// Just go make it
		w.Logs[targetName] = []LogItem{newItem}
	}
}

// ReportGeneralMessage implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportGeneralMessage(_fmt string, args ...any) {
	w.cliBackup.ReportGeneralMessage(_fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	newItem := LogItem{Msg: fmt.Sprintf(_fmt, args...), Time: time.Now(), Level: "info"}

	w.GeneralInfo = append(w.GeneralInfo, newItem)
}

// ReportInfo implements monitor.MonitoringInteractor.
func (w *WebMonitor) ReportInfo(targetName string, _fmt string, args ...any) {
	w.cliBackup.ReportInfo(targetName, _fmt, args...)
	unclaim := w.claim()
	defer unclaim()
	newItem := LogItem{Msg: fmt.Sprintf(_fmt, args...), Time: time.Now(), Level: "info"}
	if val, ok := w.Logs[targetName]; ok {
		w.Logs[targetName] = append(val, newItem)
	} else {
		// Just go make it
		w.Logs[targetName] = []LogItem{newItem}
	}
}

func NewWebMonitor(ctx context.Context, targets []monitor.TunnelTarget) (*WebMonitor, error) {
	mon := &WebMonitor{
		lock:    &sync.Mutex{},
		Targets: targets,
		Logs:    make(map[string][]LogItem),

		GeneralInfo: []LogItem{},
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
			log.Printf("Error shutting down web server: %s", err)
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
