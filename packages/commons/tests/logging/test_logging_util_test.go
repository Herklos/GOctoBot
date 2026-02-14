package tests

import (
    "reflect"
    "testing"

    logging "octobot_commons/octobot_commons/logging"
)

func TestRegisterErrorCallback(t *testing.T) {
    other := func(_ error, _ string) {}
    logging.RegisterErrorCallback(logging.DefaultCallback())
    if reflect.ValueOf(logging.ErrorCallback()).Pointer() != reflect.ValueOf(logging.DefaultCallback()).Pointer() {
        t.Fatalf("expected default callback")
    }
    logging.RegisterErrorCallback(other)
    if reflect.ValueOf(logging.ErrorCallback()).Pointer() == reflect.ValueOf(logging.DefaultCallback()).Pointer() {
        t.Fatalf("expected other callback")
    }
}

func TestError(t *testing.T) {
    called := 0
    callback := func(_ error, message string) {
        if message != "err" {
            t.Fatalf("unexpected message")
        }
        called++
    }
    logging.RegisterErrorCallback(callback)
    logger := logging.GetLogger("test")
    logger.Error("err")
    if called != 1 {
        t.Fatalf("expected callback called once")
    }
    called = 0
    logger.ErrorWithSkip("err", true)
    if called != 0 {
        t.Fatalf("expected callback not called")
    }
}

func TestException(t *testing.T) {
    called := 0
    callback := func(err error, message string) {
        if message != "error" {
            t.Fatalf("unexpected message")
        }
        if err == nil {
            t.Fatalf("expected error")
        }
        called++
    }
    logging.RegisterErrorCallback(callback)
    logger := logging.GetLogger("test")
    err := errTest{}
    logger.Exception(err, true, "error", true, false)
    if called != 1 {
        t.Fatalf("expected callback called once")
    }
    called = 0
    logger.Exception(err, true, "error", true, true)
    if called != 0 {
        t.Fatalf("expected callback not called")
    }
}

type errTest struct{}

func (e errTest) Error() string { return "err" }
