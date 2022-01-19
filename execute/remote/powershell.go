package remote

import (
	"encoding/base64"
	"fmt"

	"github.com/mitre/gocat/execute"
	"golang.org/x/text/encoding/unicode"
)

type RemotePowershell struct {
	shortName      string
	prependCommand string
}

func init() {
	remoteShell := &RemotePowershell{
		shortName:      "psh",
		prependCommand: "powershell.exe -exec bypass -enc",
	}
	execute.RemoteExecutors[remoteShell.shortName] = remoteShell

}

func (p *RemotePowershell) PrepareCommand(command string) string {
	uni := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	encodedCommand, _ := uni.NewEncoder().String(command)

	return fmt.Sprintf("%s %s", p.prependCommand, base64.StdEncoding.EncodeToString([]byte(encodedCommand)))
}
