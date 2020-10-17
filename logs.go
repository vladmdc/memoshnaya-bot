package bot

import "github.com/rs/zerolog"

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, _ string) {
	if level != zerolog.NoLevel {
		e.Str("severity", severity(level))
	}
}

func severity(level zerolog.Level) string {
	switch level {
	case zerolog.DebugLevel:
		return "DEBUG"
	case zerolog.InfoLevel:
		return "INFO"
	case zerolog.WarnLevel:
		return "WARN"
	case zerolog.ErrorLevel:
		return "ERROR"
	case zerolog.PanicLevel:
		return "CRITICAL"
	default:
		return "DEFAULT"
	}
}
