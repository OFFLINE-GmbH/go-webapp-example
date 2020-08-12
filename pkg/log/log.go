package log

import (
	"io"
	"io/ioutil"
	"os"

	"go-webapp-example/pkg/fs"

	"github.com/sirupsen/logrus"
)

// Logger implementation is responsible for providing structured and levled
// logging functions.
type Logger interface {
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Info(msg string)
	Infoln(...interface{})
	Warn(msg string)
	Warnln(...interface{})
	Warnf(msg string, args ...interface{})
	Error(msg string)
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
	Print(args ...interface{})
	Printf(msg string, args ...interface{})
	Println(...interface{})
	Trace(args ...interface{})
	Tracef(msg string, args ...interface{})
	Traceln(...interface{})
	Verbose() bool

	// WithFields should return a logger which is annotated with the given
	// fields. These fields should be added to every logging call on the
	// returned logger.
	WithFields(m map[string]interface{}) Logger
	WithPrefix(prefix string) Logger
}

type Fields logrus.Fields

// New returns a logger implemented using the logrus package.
func New(wr io.Writer, level, dir string) Logger {
	if wr == nil {
		wr = os.Stderr
	}

	lr := logrus.New()
	lr.Out = wr

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.WarnLevel
		lr.Warnf("failed to parse log-level '%s', defaulting to 'warning'", level)
	}
	lr.SetLevel(lvl)
	lr.SetFormatter(getFormatter(false))

	if dir != "" {
		_ = fs.EnsureDir(dir)
		// nolint:gocritic
		fileHook, err := NewLogrusFileHook(dir+"/app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err == nil {
			lr.Hooks.Add(fileHook)
		} else {
			lr.Warnf("Failed to open logfile, using standard out: %v", err)
		}
	}

	return &logrusLogger{
		Entry: logrus.NewEntry(lr),
	}
}

func NewNullLogger() Logger {
	lr := logrus.New()
	lr.SetOutput(ioutil.Discard)
	return &logrusLogger{Entry: logrus.NewEntry(lr)}
}

func NewFromWriter(w io.Writer) Logger {
	lr := logrus.New()
	lr.SetOutput(w)
	lr.SetLevel(logrus.DebugLevel)
	return &logrusLogger{Entry: logrus.NewEntry(lr)}
}

// logrusLogger provides functions for structured logging.
type logrusLogger struct {
	*logrus.Entry
}

func (ll *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	annotatedEntry := ll.Entry.WithFields(fields)
	return &logrusLogger{
		Entry: annotatedEntry,
	}
}

func (ll *logrusLogger) Error(msg string) {
	ll.Errorf(msg)
}

func (ll *logrusLogger) Info(msg string) {
	ll.Infof(msg)
}

func (ll *logrusLogger) Print(args ...interface{}) {
	ll.Debug(args...)
}

func (ll *logrusLogger) Warn(msg string) {
	ll.Warnf(msg)
}

func (ll *logrusLogger) Verbose() bool {
	return ll.Entry.Logger.GetLevel().String() == "debug"
}

func (ll *logrusLogger) WithPrefix(prefix string) Logger {
	return ll.WithFields(Fields{"prefix": prefix})
}

// getFormatter returns the default log formatter.
func getFormatter(disableColors bool) *textFormatter {
	return &textFormatter{
		DisableColors:    disableColors,
		ForceFormatting:  true,
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    true,
		DisableSorting:   true,
		TimestampFormat:  "2006-01-02 15:04:05.000000",
		SpacePadding:     45,
	}
}
