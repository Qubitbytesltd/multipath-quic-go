package utils

import (
	"bytes"
	"log"
	"os"
<<<<<<< HEAD
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogLevelNothing(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	DefaultLogger.SetLogLevel(LogLevelNothing)
	DefaultLogger.Debugf("debug")
	DefaultLogger.Infof("info")
	DefaultLogger.Errorf("err")
	require.Empty(t, b.String())
}

func TestLogLevelError(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	DefaultLogger.SetLogLevel(LogLevelError)
	DefaultLogger.Debugf("debug")
	DefaultLogger.Infof("info")
	DefaultLogger.Errorf("err")
	require.Contains(t, b.String(), "err\n")
	require.NotContains(t, b.String(), "info")
	require.NotContains(t, b.String(), "debug")
}

func TestLogLevelInfo(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	DefaultLogger.SetLogLevel(LogLevelInfo)
	DefaultLogger.Debugf("debug")
	DefaultLogger.Infof("info")
	DefaultLogger.Errorf("err")
	require.Contains(t, b.String(), "err\n")
	require.Contains(t, b.String(), "info\n")
	require.NotContains(t, b.String(), "debug")
}

func TestLogLevelDebug(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	require.False(t, DefaultLogger.Debug())
	DefaultLogger.SetLogLevel(LogLevelDebug)
	require.True(t, DefaultLogger.Debug())
	DefaultLogger.Debugf("debug")
	DefaultLogger.Infof("info")
	DefaultLogger.Errorf("err")
	require.Contains(t, b.String(), "err\n")
	require.Contains(t, b.String(), "info\n")
	require.Contains(t, b.String(), "debug\n")
}

func TestNoTimestampWithEmptyFormat(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	DefaultLogger.SetLogLevel(LogLevelDebug)
	DefaultLogger.SetLogTimeFormat("")
	DefaultLogger.Debugf("debug")
	require.Equal(t, "debug\n", b.String())
}

func TestAddTimestamp(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	format := "Jan 2, 2006"
	DefaultLogger.SetLogTimeFormat(format)
	DefaultLogger.SetLogLevel(LogLevelInfo)
	DefaultLogger.Infof("info")
	timestamp := b.String()[:b.Len()-6]
	parsedTime, err := time.Parse(format, timestamp)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), parsedTime, 25*time.Hour)
}

func TestLogAddPrefixes(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	defer log.SetOutput(os.Stdout)
	defer DefaultLogger.SetLogLevel(LogLevelNothing)

	DefaultLogger.SetLogLevel(LogLevelDebug)

	// single prefix
	prefixLogger := DefaultLogger.WithPrefix("prefix")
	prefixLogger.Debugf("debug1")
	require.Contains(t, b.String(), "prefix")
	require.Contains(t, b.String(), "debug1")

	// multiple prefixes
	b.Reset()
	prefixLogger1 := DefaultLogger.WithPrefix("prefix1")
	prefixLogger2 := prefixLogger1.WithPrefix("prefix2")
	prefixLogger2.Debugf("debug2")
	require.Contains(t, b.String(), "prefix1")
	require.Contains(t, b.String(), "prefix2")
	require.Contains(t, b.String(), "debug2")
}

func TestLogLevelFromEnv(t *testing.T) {
	defer os.Unsetenv(logEnv)

	testCases := []struct {
		envValue string
		expected LogLevel
	}{
		{"DEBUG", LogLevelDebug},
		{"debug", LogLevelDebug},
		{"INFO", LogLevelInfo},
		{"ERROR", LogLevelError},
	}

	for _, tc := range testCases {
		os.Setenv(logEnv, tc.envValue)
		require.Equal(t, tc.expected, readLoggingEnv())
	}

	// invalid values
	os.Setenv(logEnv, "")
	require.Equal(t, LogLevelNothing, readLoggingEnv())
	os.Setenv(logEnv, "asdf")
	require.Equal(t, LogLevelNothing, readLoggingEnv())
}
=======
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Log", func() {
	var (
		b *bytes.Buffer

		initialTimeFormat string
	)

	BeforeEach(func() {
		b = bytes.NewBuffer([]byte{})
		log.SetOutput(b)
		initialTimeFormat = timeFormat
	})

	AfterEach(func() {
		log.SetOutput(os.Stdout)
		SetLogLevel(LogLevelNothing)
		timeFormat = initialTimeFormat
	})

	It("the log level has the correct numeric value", func() {
		Expect(LogLevelNothing).To(BeEquivalentTo(0))
		Expect(LogLevelError).To(BeEquivalentTo(1))
		Expect(LogLevelInfo).To(BeEquivalentTo(2))
		Expect(LogLevelDebug).To(BeEquivalentTo(3))
	})

	It("log level nothing", func() {
		SetLogLevel(LogLevelNothing)
		Debugf("debug")
		Infof("info")
		Errorf("err")
		Expect(b.Bytes()).To(Equal([]byte("")))
	})

	It("log level err", func() {
		SetLogLevel(LogLevelError)
		Debugf("debug")
		Infof("info")
		Errorf("err")
		Expect(b.Bytes()).To(ContainSubstring("err\n"))
		Expect(b.Bytes()).ToNot(ContainSubstring("info"))
		Expect(b.Bytes()).ToNot(ContainSubstring("debug"))
	})

	It("log level info", func() {
		SetLogLevel(LogLevelInfo)
		Debugf("debug")
		Infof("info")
		Errorf("err")
		Expect(b.Bytes()).To(ContainSubstring("err\n"))
		Expect(b.Bytes()).To(ContainSubstring("info\n"))
		Expect(b.Bytes()).ToNot(ContainSubstring("debug"))
	})

	It("log level debug", func() {
		SetLogLevel(LogLevelDebug)
		Debugf("debug")
		Infof("info")
		Errorf("err")
		Expect(b.Bytes()).To(ContainSubstring("err\n"))
		Expect(b.Bytes()).To(ContainSubstring("info\n"))
		Expect(b.Bytes()).To(ContainSubstring("debug\n"))
	})

	It("doesn't add a timestamp if the time format is empty", func() {
		SetLogLevel(LogLevelDebug)
		SetLogTimeFormat("")
		Debugf("debug")
		Expect(b.Bytes()).To(Equal([]byte("debug\n")))
	})

	It("adds a timestamp", func() {
		format := "Jan 2, 2006"
		SetLogTimeFormat(format)
		SetLogLevel(LogLevelInfo)
		Infof("info")
		t, err := time.Parse(format, string(b.Bytes()[:b.Len()-6]))
		Expect(err).ToNot(HaveOccurred())
		Expect(t).To(BeTemporally("~", time.Now(), 25*time.Hour))
	})

	It("says whether debug is enabled", func() {
		Expect(Debug()).To(BeFalse())
		SetLogLevel(LogLevelDebug)
		Expect(Debug()).To(BeTrue())
	})

	Context("reading from env", func() {
		BeforeEach(func() {
			Expect(logLevel).To(Equal(LogLevelNothing))
		})

		It("reads DEBUG", func() {
			os.Setenv(logEnv, "DEBUG")
			readLoggingEnv()
			Expect(logLevel).To(Equal(LogLevelDebug))
		})

		It("reads debug", func() {
			os.Setenv(logEnv, "debug")
			readLoggingEnv()
			Expect(logLevel).To(Equal(LogLevelDebug))
		})

		It("reads INFO", func() {
			os.Setenv(logEnv, "INFO")
			readLoggingEnv()
			Expect(logLevel).To(Equal(LogLevelInfo))
		})

		It("reads ERROR", func() {
			os.Setenv(logEnv, "ERROR")
			readLoggingEnv()
			Expect(logLevel).To(Equal(LogLevelError))
		})

		It("does not error reading invalid log levels from env", func() {
			Expect(logLevel).To(Equal(LogLevelNothing))
			os.Setenv(logEnv, "")
			readLoggingEnv()
			Expect(logLevel).To(Equal(LogLevelNothing))
			os.Setenv(logEnv, "asdf")
			readLoggingEnv()
			Expect(logLevel).To(Equal(LogLevelNothing))
		})
	})
})
>>>>>>> project-faster/main
