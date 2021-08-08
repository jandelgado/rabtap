// a Logger implementation for tests
package testcommon

import "log"

type TestLogger struct{}

func (s TestLogger) Debugf(format string, a ...interface{}) { log.Printf("DEBUG "+format, a...) }
func (s TestLogger) Infof(format string, a ...interface{})  { log.Printf("INFO "+format, a...) }
func (s TestLogger) Errorf(format string, a ...interface{}) { log.Printf("ERROR "+format, a...) }

func NewTestLogger() *TestLogger {
	return &TestLogger{}
}
