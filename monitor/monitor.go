package monitor

type MonitoringInteractor interface {
	ReportInfo(targetName string, fmt string, args ...any)
	ReportError(targetName string, fmt string, args ...any)
	ReportFatalError(targetName string, fmt string, args ...any)

	ReportGeneralMessage(fmt string, args ...any)

	// TODO Ask for input
}

type (
	MonitoredTargets struct {
		// The tunnel targets
		Monitored []TunnelTarget

		// How to start/kill
		// Targeter map[string]IndividualTarget

		// What are we outputting to? The terminal
		Output MonitoringInteractor
	}
)
