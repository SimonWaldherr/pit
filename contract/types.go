package contract

import (
	"fmt"
	"net/http"

	"github.com/zenazn/goji/web"
)

func UnexpectedResponseError(exp *http.Response, got *http.Response, err error) error {
	return TestingError(fmt.Sprintf("Implementation returned '%s', expected '%s': %s", got.Status, exp.Status, err.Error()))
}

func TestingError(msg string) error { return fmt.Errorf("Test ERROR: %s", msg) }
func MockingError(msg string) error { return fmt.Errorf("Mocking ERROR: %s", msg) }

type TestFunc func(host string, c *http.Client) error

type R interface {
	Actions() ([]A, error)
	Pattern() string
	Name() string
}

type A interface {
	AddPair(p *Pair)
	Method() string
	Handler() (web.Handler, error)
	Tests() []TestFunc
}

type C interface {
	Name() string
	Resources() ([]R, error)
}
