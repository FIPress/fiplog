package fiplog

import (
	"fmt"
	"testing"
	"time"
)

var c chan bool

func TestInit(t *testing.T) {
	logger := GetLogger()

	//	logger.Log(Warning, "warning message")
	logger.Debug("debug message")
	logger.Error("error message")
	logger.Info("info message")
	logger.Warning("warning", "message")
	logger.Errorf("this is the %d message", 2)
	logger.Close()
	/*logger.Debug("debug message")
	logger.Info("info message")
	logger.Error("error")
	logger.Warn("warn")*/

	/*n := time.Now()
	t.Log(n.Format("2006-03-11 11:05:05"))
	t.Log(n.Format("2006-01-02 03:04:05PM"))
	t.Log(n.Format(time.RFC3339))*/
}

func sub(buf []byte) {
	fmt.Println("sub:enter")
	time.Sleep(time.Duration(300))
	fmt.Println("sub:", buf)
	c <- true
}
