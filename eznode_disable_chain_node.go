package eznode

import "time"

// DisableNode disables a node from a chain
func (e *EzNode) DisableNode(chainId string, nodeName string) {
	for _, chain := range e.chains {
		if chain.id == chainId {
			chain.disableNode(nodeName)
			break
		}
	}
}

// DisableNodeWithTime disables a node from a chain for a given time
func (e *EzNode) DisableNodeWithTime(chainId string, nodeName string, duration time.Duration) {
	for _, chain := range e.chains {
		if chain.id == chainId {
			chain.disableNodeWithTime(nodeName, duration)
			break
		}
	}
}

// EnableNode enables a node from a chain
func (e *EzNode) EnableNode(chainId string, nodeName string) {
	for _, chain := range e.chains {
		if chain.id == chainId {
			chain.enableNode(nodeName)
			break
		}
	}
}
