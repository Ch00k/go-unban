package f2bclient

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func mockExecCommand(command string, s ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, s...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, os.Getenv("GO_WANT_STDOUT"))
	fmt.Fprintf(os.Stderr, os.Getenv("GO_WANT_STDERR"))
	os.Exit(getReturnCode())
}

func getReturnCode() int {
	ret := os.Getenv("GO_WANT_RETURN_CODE")
	if ret == "" {
		ret = "0"
	}
	returnCode, err := strconv.Atoi(ret)
	if err != nil {
		panic(err)
	}
	return returnCode
}

func compileEnv(returnCode, stdout, stderr string) []string {
	return []string{
		fmt.Sprintf("GO_WANT_RETURN_CODE=%s", returnCode),
		fmt.Sprintf("GO_WANT_STDOUT=%s", stdout),
		fmt.Sprintf("GO_WANT_STDERR=%s", stderr),
	}
}

func checkResult(actualOut, expectedOut string, actualErr, expectedErr error, t *testing.T) {
	if actualOut != expectedOut {
		t.Errorf("Expected output '%s', got '%s' instead", expectedOut, actualOut)
	}
	if actualErr == nil && expectedErr == nil {
		return
	}
	// TODO: it will panic if one of them in nil, but the other one is not
	if actualErr.Error() != expectedErr.Error() {
		t.Errorf("Expected error '%s', got '%s' instead", expectedErr.Error(), actualErr.Error())
	}
}

func TestFail2banClient(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	t.Run("RC=0,O=foo,E=bar", func(t *testing.T) {
		rc := "0"
		o := "foo"
		e := "bar"
		env := compileEnv(rc, o, e)

		expectedOut := o + e
		var expectedErr error

		out, err := runFail2banClient(env)
		checkResult(out, expectedOut, err, expectedErr, t)
	})

	t.Run("RC=1,O=foo,E=bar", func(t *testing.T) {
		rc := "1"
		o := "foo"
		e := "bar"
		env := compileEnv(rc, o, e)

		expectedOut := o + e
		expectedErr := errors.New("exit status 1")

		out, err := runFail2banClient(env)
		checkResult(out, expectedOut, err, expectedErr, t)
	})
}
