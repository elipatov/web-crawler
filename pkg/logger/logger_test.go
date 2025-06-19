package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	field1Name = "field1"
	field2Name = "field2"
)

type logMsg struct {
	Message    string `json:"msg"`
	Error      string `json:"error"`
	StackTrace string `json:"stack_trace"`
	ErrorCode  string `json:"error_code"`
	UserID     string `json:"user_id"`
	Field1     string `json:"field1"`
	Field2     string `json:"field2"`
}

func TestLogger(t *testing.T) {
	stderr := os.Stderr
	read, write, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		err := write.Close()
		assert.NoError(t, err)
	}()

	os.Stderr = write
	scanner := bufio.NewScanner(read)
	logger := New("DEBUG")

	getLog := func() logMsg {
		ok := scanner.Scan()
		require.True(t, ok)

		logStr := scanner.Text()
		log := logMsg{}
		err = json.Unmarshal([]byte(logStr), &log)
		require.NoError(t, err)

		return log
	}

	t.Run("Logs custom error fields", func(t *testing.T) {
		err := &testError{
			message:    "Error message",
			code:       "BadRequest",
			stackTrace: "logger.TestLogger",
		}

		logger.WithError(err).Error("Some message")

		log := getLog()
		assert.Equal(t, err.code, log.ErrorCode)
		assert.Equal(t, "Error message", log.Error)
		assert.Equal(t, "Some message", log.Message)
		assert.Contains(t, log.StackTrace, "logger.TestLogger")
	})

	t.Run("Logs error fields", func(t *testing.T) {
		err := &testError{
			message:    "Error message",
			code:       "E1024",
			stackTrace: "Logs error fields",
		}

		logger.WithError(err).Error("Some message")

		log := getLog()
		assert.Equal(t, err.message, log.Error)
		assert.Equal(t, "Some message", log.Message)
		assert.Equal(t, err.stackTrace, log.StackTrace)
		assert.Equal(t, err.code, log.ErrorCode)
	})

	t.Run("WithField keeps error fields", func(t *testing.T) {
		err := &testError{
			message:    "Error message",
			code:       "E064",
			stackTrace: "Logs error fields",
		}

		logger.WithField(field1Name, "WithField 1").
			WithField(field2Name, "WithField 2").
			WithError(err).Error("Some message")

		log := getLog()
		assert.Equal(t, err.message, log.Error)
		assert.Equal(t, "Some message", log.Message)
		assert.Equal(t, err.stackTrace, log.StackTrace)
		assert.Equal(t, err.code, log.ErrorCode)
		assert.Equal(t, "WithField 1", log.Field1)
		assert.Equal(t, "WithField 2", log.Field2)

		logger.WithField(field1Name, "WithFields 1").
			WithField(field2Name, "WithFields 2").
			WithError(err).Error("Some message")

		log = getLog()
		assert.Equal(t, err.message, log.Error)
		assert.Equal(t, "Some message", log.Message)
		assert.Equal(t, err.stackTrace, log.StackTrace)
		assert.Equal(t, err.code, log.ErrorCode)
		assert.Equal(t, "WithFields 1", log.Field1)
		assert.Equal(t, "WithFields 2", log.Field2)
	})

	os.Stderr = stderr
}

type testError struct {
	message    string
	code       string
	stackTrace string
}

// Error method returns error string representation.
func (err *testError) Error() string {
	return err.message
}

// Code returns ErrorCode as string.
func (err *testError) Code() string {
	return err.code
}

// StackTrace of the error.
func (err *testError) StackTrace() string {
	return err.stackTrace
}
