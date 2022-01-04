package eznode

import "time"

func (e *EzNode) DisableNode(chainId string, nodeName string) {
	for _, chain := range e.chains {
		if chain.id == chainId {
			chain.disableNode(nodeName)
			break
		}
	}
}

func (e *EzNode) DisableNodeWithTime(chainId string, nodeName string, duration time.Duration) {
	for _, chain := range e.chains {
		if chain.id == chainId {
			chain.disableNodeWithTime(nodeName, duration)
			break
		}
	}
}

func (e *EzNode) EnableNode(chainId string, nodeName string) {
	for _, chain := range e.chains {
		if chain.id == chainId {
			chain.enableNode(nodeName)
			break
		}
	}
}
