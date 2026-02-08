package api

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/getAlby/hub/tests/mocks"
)

func TestMain(m *testing.M) {
	logger.Init("")
	os.Exit(m.Run())
}

func TestGetCustomNodeCommandDefinitions(t *testing.T) {
	lnClient := mocks.NewMockLNClient(t)
	svc := mocks.NewMockService(t)

	mockLNCommandDefs := []lnclient.CustomNodeCommandDef{
		{
			Name:        "no_args",
			Description: "command without args",
			Args:        nil,
		},
		{
			Name:        "with_args",
			Description: "command with args",
			Args: []lnclient.CustomNodeCommandArgDef{
				{Name: "arg1", Description: "first argument"},
				{Name: "arg2", Description: "second argument"},
			},
		},
	}

	expectedCommands := []CustomNodeCommandDef{
		{
			Name:        "no_args",
			Description: "command without args",
			Args:        []CustomNodeCommandArgDef{},
		},
		{
			Name:        "with_args",
			Description: "command with args",
			Args: []CustomNodeCommandArgDef{
				{Name: "arg1", Description: "first argument"},
				{Name: "arg2", Description: "second argument"},
			},
		},
	}

	lnClient.On("GetCustomNodeCommandDefinitions").Return(mockLNCommandDefs)
	svc.On("GetLNClient").Return(lnClient)

	theAPI := instantiateAPIWithService(svc)

	commands, err := theAPI.GetCustomNodeCommands()
	require.NoError(t, err)
	require.NotNil(t, commands)
	require.ElementsMatch(t, expectedCommands, commands.Commands)
}

func TestExecuteCustomNodeCommand(t *testing.T) {
	type testCase struct {
		name                 string
		apiCommandLine       string
		lnSupportedCommands  []lnclient.CustomNodeCommandDef
		lnExpectedCommandReq *lnclient.CustomNodeCommandRequest
		lnResponse           *lnclient.CustomNodeCommandResponse
		lnError              error
		apiExpectedResponse  interface{}
		apiExpectedErr       string
	}

	// Successful execution of a command without args.
	testCaseOkNoArgs := testCase{
		name:                 "command without args",
		apiCommandLine:       "test_command",
		lnSupportedCommands:  []lnclient.CustomNodeCommandDef{{Name: "test_command"}},
		lnExpectedCommandReq: &lnclient.CustomNodeCommandRequest{Name: "test_command", Args: []lnclient.CustomNodeCommandArg{}},
		lnResponse:           &lnclient.CustomNodeCommandResponse{Response: "ok"},
		lnError:              nil,
		apiExpectedResponse:  "ok",
		apiExpectedErr:       "",
	}

	// Successful execution of a command with args. The command line contains
	// different arg value styles: with '=' and with space.
	testCaseOkWithArgs := testCase{
		name:           "command with args",
		apiCommandLine: "test_command --arg1=foo --arg2 bar",
		lnSupportedCommands: []lnclient.CustomNodeCommandDef{
			{
				Name: "test_command",
				Args: []lnclient.CustomNodeCommandArgDef{
					{Name: "arg1", Description: "argument one"},
					{Name: "arg2", Description: "argument two"},
				},
			},
		},
		lnExpectedCommandReq: &lnclient.CustomNodeCommandRequest{Name: "test_command", Args: []lnclient.CustomNodeCommandArg{
			{Name: "arg1", Value: "foo"},
			{Name: "arg2", Value: "bar"},
		}},
		lnResponse:          &lnclient.CustomNodeCommandResponse{Response: "ok"},
		lnError:             nil,
		apiExpectedResponse: "ok",
		apiExpectedErr:      "",
	}

	// Successful execution of a command with a possible but unset arg.
	testCaseOkWithUnsetArg := testCase{
		name:           "command with unset arg",
		apiCommandLine: "test_command",
		lnSupportedCommands: []lnclient.CustomNodeCommandDef{
			{Name: "test_command", Args: []lnclient.CustomNodeCommandArgDef{{Name: "arg1", Description: "argument one"}}},
		},
		lnExpectedCommandReq: &lnclient.CustomNodeCommandRequest{Name: "test_command", Args: []lnclient.CustomNodeCommandArg{}},
		lnResponse:           &lnclient.CustomNodeCommandResponse{Response: "ok"},
		lnError:              nil,
		apiExpectedResponse:  "ok",
		apiExpectedErr:       "",
	}

	// Error: command line is empty.
	testCaseErrEmptyCommand := testCase{
		name:                 "empty command",
		apiCommandLine:       "",
		lnSupportedCommands:  nil,
		lnExpectedCommandReq: nil,
		lnResponse:           nil,
		lnError:              nil,
		apiExpectedResponse:  nil,
		apiExpectedErr:       "no command provided",
	}

	// Error: command line is malformed, i.e. non-parseable.
	testCaseErrMalformedCommand := testCase{
		name:                 "command with unclosed quote",
		apiCommandLine:       "test_command\"",
		lnSupportedCommands:  nil,
		lnExpectedCommandReq: nil,
		lnResponse:           nil,
		lnError:              nil,
		apiExpectedResponse:  nil,
		apiExpectedErr:       "failed to parse node command",
	}

	// Error: node does not support this command.
	testCaseErrUnknownCommand := testCase{
		name:                 "unknown command",
		apiCommandLine:       "test_command_unknown",
		lnSupportedCommands:  []lnclient.CustomNodeCommandDef{{Name: "test_command"}},
		lnExpectedCommandReq: nil,
		lnResponse:           nil,
		lnError:              nil,
		apiExpectedResponse:  nil,
		apiExpectedErr:       "unknown command",
	}

	// Error: unsupported command argument.
	testCaseErrUnknownArg := testCase{
		name:                 "unknown argument",
		apiCommandLine:       "test_command --unknown=fail",
		lnSupportedCommands:  []lnclient.CustomNodeCommandDef{{Name: "test_command"}},
		lnExpectedCommandReq: nil,
		lnResponse:           nil,
		lnError:              nil,
		apiExpectedResponse:  nil,
		apiExpectedErr:       "flag provided but not defined: -unknown",
	}

	// Error: the command is valid but the node fails to execute it.
	testCaseErrNodeFailed := testCase{
		name:                 "node failed to execute command",
		apiCommandLine:       "test_command",
		lnSupportedCommands:  []lnclient.CustomNodeCommandDef{{Name: "test_command"}},
		lnExpectedCommandReq: &lnclient.CustomNodeCommandRequest{Name: "test_command", Args: []lnclient.CustomNodeCommandArg{}},
		lnResponse:           nil,
		lnError:              fmt.Errorf("utter failure"),
		apiExpectedResponse:  nil,
		apiExpectedErr:       "utter failure",
	}

	testCases := []testCase{
		testCaseOkNoArgs,
		testCaseOkWithArgs,
		testCaseOkWithUnsetArg,
		testCaseErrEmptyCommand,
		testCaseErrMalformedCommand,
		testCaseErrUnknownCommand,
		testCaseErrUnknownArg,
		testCaseErrNodeFailed,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lnClient := mocks.NewMockLNClient(t)
			svc := mocks.NewMockService(t)

			if tc.lnSupportedCommands != nil {
				lnClient.On("GetCustomNodeCommandDefinitions").Return(tc.lnSupportedCommands)
			}

			if tc.lnExpectedCommandReq != nil {
				lnClient.On("ExecuteCustomNodeCommand", mock.Anything, tc.lnExpectedCommandReq).Return(tc.lnResponse, tc.lnError)
			}

			svc.On("GetLNClient").Return(lnClient)

			theAPI := instantiateAPIWithService(svc)

			response, err := theAPI.ExecuteCustomNodeCommand(context.TODO(), tc.apiCommandLine)
			require.Equal(t, tc.apiExpectedResponse, response)
			if tc.apiExpectedErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.apiExpectedErr)
			}
		})
	}
}

func TestSetNodeAlias(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := mocks.NewMockConfig(t)
		cfg.On("SetUpdate", "NodeAlias", "SatoshiSquirrel", "").Return(nil)

		theAPI := &api{cfg: cfg}

		err := theAPI.SetNodeAlias("SatoshiSquirrel")
		require.NoError(t, err)
		cfg.AssertExpectations(t)
	})

	t.Run("empty alias", func(t *testing.T) {
		cfg := mocks.NewMockConfig(t)
		cfg.On("SetUpdate", "NodeAlias", "", "").Return(nil)

		theAPI := &api{cfg: cfg}

		err := theAPI.SetNodeAlias("")
		require.NoError(t, err)
		cfg.AssertExpectations(t)
	})

	t.Run("config error", func(t *testing.T) {
		cfg := mocks.NewMockConfig(t)
		cfg.On("SetUpdate", "NodeAlias", "test", "").Return(fmt.Errorf("database error"))

		theAPI := &api{cfg: cfg}

		err := theAPI.SetNodeAlias("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "database error")
		cfg.AssertExpectations(t)
	})
}

// instantiateAPIWithService is a helper function that returns a partially
// constructed API instance. It is only suitable for the simplest of test cases.
func instantiateAPIWithService(s service.Service) *api {
	return &api{svc: s}
}
