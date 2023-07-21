package cli

import (
	"bytes"
	"io"
	"testing"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogsCmd(t *testing.T) {
	const (
		testOutput = `option-returner: INFO:     Started server process [1]
option-returner: INFO:     Waiting for application startup.
option-returner: INFO:     Application startup complete.
option-returner: INFO:     Uvicorn running on http://0.0.0.0:8080 (Press CTRL+C to quit)
option-returner: INFO:     172.29.0.3:38684 - "GET /metrics HTTP/1.1" 307 Temporary Redirect
option-returner: INFO:     172.29.0.3:38684 - "GET / HTTP/1.1" 200 OK`
		testOutputWithTimestamps = `option-returner: 2023-07-21T10:23:24.737869418Z INFO:     Started server process [1]
option-returner: 2023-07-21T10:23:24.737964084Z INFO:     Waiting for application startup.
option-returner: 2023-07-21T10:23:24.741941834Z INFO:     Application startup complete.
option-returner: 2023-07-21T10:23:24.749516459Z INFO:     Uvicorn running on http://0.0.0.0:8080 (Press CTRL+C to quit)
option-returner: 2023-07-21T10:23:35.542743798Z INFO:     172.29.0.3:38684 - "GET /metrics HTTP/1.1" 307 Temporary Redirect
option-returner: 2023-07-21T10:23:35.558762590Z INFO:     172.29.0.3:38684 - "GET / HTTP/1.1" 200 OK`
	)

	type testCase struct {
		name           string
		args           []string
		mocker         func(d *daemonMock.MockDaemon, out io.Writer)
		expectedOutput string
		expectedErr    error
	}
	tc := []testCase{
		{
			name: "success",
			args: []string{"id1"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Tail: "all",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "daemon error",
			args: []string{"id1"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", gomock.Any()).Return(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name: "with no flags",
			args: []string{"id1"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Tail: "all",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "follow flag is set properly",
			args: []string{"id1", "--follow"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Follow: true,
					Tail:   "all",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "follow is set with the shorthand letter",
			args: []string{"id1", "-f"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Follow: true,
					Tail:   "all",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "since flag is set properly",
			args: []string{"id1", "--since", "2013-01-02T13:23:37Z"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Since: "2013-01-02T13:23:37Z",
					Tail:  "all",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "until flag is set properly",
			args: []string{"id1", "--until", "2013-01-02T13:23:37Z"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Until: "2013-01-02T13:23:37Z",
					Tail:  "all",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "timestamps flag is set properly",
			args: []string{"id1", "--timestamps"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Timestamps: true,
					Tail:       "all",
				}).Return(nil)
				out.Write([]byte(testOutputWithTimestamps))
			},
			expectedOutput: testOutputWithTimestamps,
		},
		{
			name: "timestamps is set with the shorthand letter",
			args: []string{"id1", "-t"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Timestamps: true,
					Tail:       "all",
				}).Return(nil)
				out.Write([]byte(testOutputWithTimestamps))
			},
			expectedOutput: testOutputWithTimestamps,
		},
		{
			name: "tail flag is set properly",
			args: []string{"id1", "--tail", "6"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Tail: "6",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
		{
			name: "tail is set with the shorthand letter",
			args: []string{"id1", "-n", "6"},
			mocker: func(d *daemonMock.MockDaemon, out io.Writer) {
				d.EXPECT().NodeLogs(gomock.Any(), gomock.Any(), "id1", daemon.NodeLogsOptions{
					Tail: "6",
				}).Return(nil)
				out.Write([]byte(testOutput))
			},
			expectedOutput: testOutput,
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			cmdOut := new(bytes.Buffer)

			d := daemonMock.NewMockDaemon(gomock.NewController(t))
			if tt.mocker != nil {
				tt.mocker(d, cmdOut)
			}

			logsCmd := LogsCmd(d)
			logsCmd.SetOutput(cmdOut)
			logsCmd.SetArgs(tt.args)
			err := logsCmd.Execute()
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, cmdOut.String())
			}
		})
	}
}
