package command

import (
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/codegangsta/cli"
)

var tmpl_test_success = `Tested successful!`
var tmpl_test_failed = `test failed!`

type Test struct {
	*cmd

	mock   *Mock   //the mock command
	unmock *Unmock //the unmock command
	build  *Build  // the build command
}

func NewTest(out io.Writer, mock *Mock, unmock *Unmock, build *Build) *Test {
	return &Test{
		cmd:    newCmd(out),
		mock:   mock,
		unmock: unmock,
		build:  build,
	}
}

func (c *Test) Name() string {
	return "test"
}

func (c *Test) Description() string {
	return fmt.Sprintf("...")
}

func (c *Test) Usage() string {
	return "the the implementation against current examples"
}

func (c *Test) Flags() []cli.Flag {
	fs := []cli.Flag{}
	fs = append(fs, c.build.Flags()...) //this also covers the flags for running the mock command

	return fs
}

func (c *Test) Action() func(ctx *cli.Context) {
	return c.templated(c.Run)
}

func (c *Test) Run(ctx *cli.Context) (*template.Template, interface{}, error) {

	host := ctx.Args().First()
	if host == "" {
		return nil, nil, fmt.Errorf("Please provide the address (e.g http://localhost:8000) to test as a first argument")
	}

	//run mock command
	_, _, err := c.mock.Run(ctx)
	if err != nil {
		return nil, nil, err
	}

	//run build command
	_, _, err = c.build.Run(ctx)
	if err != nil {
		return nil, nil, err
	}

	//get contract
	contract, err := c.ParseExamples(ctx)
	if err != nil {
		return nil, nil, err
	}

	//get the state manager
	sm, err := c.StateManager(ctx)
	if err != nil {
		return nil, nil, err
	}

	//run all tests, for all resources
	res, err := contract.Resources()
	if err != nil {
		return nil, nil, err
	}

	errs := []error{}
	for _, r := range res {
		acs, err := r.Actions()
		if err != nil {
			return nil, nil, err
		}

		for _, a := range acs {
			for _, tt := range a.Tests() {

				err := tt(host, http.DefaultClient, sm)
				if err != nil {
					//@todo handle failed tests better

					//gather testing errors
					errs = append(errs, err)

					fmt.Fprintf(c.out, "%s\n", err.Error())
				}
			}
		}
	}

	//unmock afterwards
	_, _, err = c.unmock.Run(ctx)
	if err != nil {
		return nil, nil, err
	}

	//we got some testing errs
	if len(errs) > 0 {
		return template.Must(template.New("test.failed").Parse(tmpl_test_failed)), errs, fmt.Errorf("Test failed")
	}

	return template.Must(template.New("test.success").Parse(tmpl_test_success)), nil, nil
}