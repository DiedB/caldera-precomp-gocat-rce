package agent

import (
	"fmt"
	"sync"

	"github.com/mitre/gocat/output"
	"github.com/mitre/gocat/proxy"
)

func (a *Agent) ActivateP2pReceivers() {
	a.validP2pReceivers = make(map[string]proxy.P2pReceiver)
	a.p2pReceiverWaitGroup = &sync.WaitGroup{}
	for receiverName, p2pReceiver := range proxy.P2pReceiverChannels {
		if err := p2pReceiver.InitializeReceiver(a.server, a.beaconContact, a.p2pReceiverWaitGroup); err != nil {
			output.VerbosePrint(fmt.Sprintf("[-] Error when initializing p2p receiver %s: %s", receiverName, err.Error()))
		} else {
			output.VerbosePrint(fmt.Sprintf("[*] Initialized p2p receiver %s", receiverName))
			a.validP2pReceivers[receiverName] = p2pReceiver
			a.p2pReceiverWaitGroup.Add(1)
			go p2pReceiver.RunReceiver()
		}
	}
}

func (a *Agent) TerminateP2pReceivers() {
	for receiverName, p2pReceiver := range a.validP2pReceivers {
		output.VerbosePrint(fmt.Sprintf("[*] Terminating p2p receiver %s", receiverName))
		p2pReceiver.Terminate()
	}
	a.p2pReceiverWaitGroup.Wait()
}

// Returns list of 2-tuple lists of form ["proxy protocol", "proxy receiver address"]
func (a *Agent) getProxyReceiverList() [][]string {
	var returnList [][]string
	for receiverName, p2pReceiver := range a.validP2pReceivers {
		for _, address := range p2pReceiver.GetReceiverAddresses() {
			proxyInfo := make([]string, 2)
			proxyInfo[0] = receiverName
			proxyInfo[1] = address
			returnList = append(returnList, proxyInfo)
		}
	}
	return returnList
}