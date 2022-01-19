package execute

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitre/gocat/output"
)

const (
	SUCCESS_STATUS = "0"
	ERROR_STATUS   = "1"
	TIMEOUT_STATUS = "124"
	SUCCESS_PID    = "0"
	ERROR_PID      = "1"
)

type LocalExecutor interface {
	// Run takes a command string, timeout int, and instruction info.
	// Returns Raw Output, A String status code, and a String PID
	Run(command string, timeout int, info InstructionInfo) ([]byte, string, string, time.Time)
	String() string
	CheckIfAvailable() bool
	UpdateBinary(newBinary string)

	// Returns true if the executor wants the payload downloaded to memory, false if it wants the payload on disk.
	DownloadPayloadToMemory(payloadName string) bool
}

type RemoteExecutor interface {
	PrepareCommand(command string) string
}

type InstructionInfo struct {
	Profile          map[string]interface{}
	Instruction      map[string]interface{}
	RceCommand       string
	OnDiskPayloads   []string
	InMemoryPayloads map[string][]byte
}

func GetLocalExecutor() (executor string) {
	for _, e := range LocalExecutors {
		// Find one of supported remote executors, otherwise return null
		if e.String() == "sh" || e.String() == "psh" {
			executor = e.String()
			return
		}
	}
	return
}

var LocalExecutors = map[string]LocalExecutor{}

var RemoteExecutors = map[string]RemoteExecutor{}

//RunCommand runs the actual command
func RunCommand(info InstructionInfo) ([]byte, string, string, time.Time) {
	encodedCommand := info.Instruction["command"].(string)
	remoteExecutor := info.Instruction["executor"].(string)
	timeout := int(info.Instruction["timeout"].(float64))
	onDiskPayloads := info.OnDiskPayloads
	var status string
	var result []byte
	var pid string
	var executionTimestamp time.Time
	decoded, err := base64.StdEncoding.DecodeString(encodedCommand)
	var preparedCommand string
	if err != nil {
		result = []byte(fmt.Sprintf("Error when decoding command: %s", err.Error()))
		status = ERROR_STATUS
		pid = ERROR_STATUS
		executionTimestamp = time.Now().UTC()
	} else {
		// Ask remote executor for prepared remote command
		output.VerbosePrint(fmt.Sprintf("Remote executor is %s", remoteExecutor))

		preparedCommand = RemoteExecutors[remoteExecutor].PrepareCommand(string(decoded))

		// Substitute received instruction command into RceCommand
		command := strings.Replace(info.RceCommand, "COMMAND", preparedCommand, 1)

		output.VerbosePrint(fmt.Sprintf("Executing substituted command: %s", command))

		missingPaths := checkPayloadsAvailable(onDiskPayloads)
		if len(missingPaths) == 0 {
			result, status, pid, executionTimestamp =

				// Use local executor to run substituted command
				LocalExecutors[GetLocalExecutor()].Run(command, timeout, info)
		} else {
			result = []byte(fmt.Sprintf("Payload(s) not available: %s", strings.Join(missingPaths, ", ")))
			status = ERROR_STATUS
			pid = ERROR_STATUS
			executionTimestamp = time.Now().UTC()
		}
	}
	return result, status, pid, executionTimestamp
}

func RemoveExecutor(name string) {
	delete(LocalExecutors, name)
}

//checkPayloadsAvailable determines if any payloads are not on disk
func checkPayloadsAvailable(payloads []string) []string {
	var missing []string
	for i := range payloads {
		if fileExists(filepath.Join(payloads[i])) == false {
			missing = append(missing, payloads[i])
		}
	}
	return missing
}

// checks for a file
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
