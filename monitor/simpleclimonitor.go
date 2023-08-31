package monitor

import "fmt"

type CliMonitor struct {
}

func (m *CliMonitor) ReportInfo(targetName string, _fmt string, args ...interface{}) {
	passedFmted := fmt.Sprintf(_fmt, args...)
	fmt.Printf("[I %s]: %s\n", targetName, passedFmted)
}

func (m *CliMonitor) ReportError(targetName string, _fmt string, args ...interface{}) {
	passedFmted := fmt.Sprintf(_fmt, args...)
	fmt.Printf("[E %s]: %s\n", targetName, passedFmted)
}

func (m *CliMonitor) ReportFatalError(targetName string, _fmt string, args ...interface{}) {
	passedFmted := fmt.Sprintf(_fmt, args...)
	fmt.Printf("[F %s]: %s\n", targetName, passedFmted)
}

func (m *CliMonitor) ReportGeneralMessage(_fmt string, args ...interface{}) {
	passedFmted := fmt.Sprintf(_fmt, args...)
	fmt.Printf("[I]: %s\n", passedFmted)
}
