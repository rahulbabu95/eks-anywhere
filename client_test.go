package main

import "testing"

func TestDefaultLogger(t *testing.T) {
	withDebug := defaultLogger(true)
	withDebug.Info("test")
	withDebug.V(1).Info("test")

	withoutDebug := defaultLogger(false)
	withoutDebug.Info("test")
	withoutDebug.V(1).Info("test")

	t.Fail()
}
