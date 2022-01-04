package eznode

import "time"

func (c *Chain) disableNode(nodeName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, node := range c.nodes {
		if node.name == nodeName {
			node.disabled = true
			break
		}
	}
}

func (c *Chain) disableNodeWithTime(nodeName string, duration time.Duration) {
	c.disableNode(nodeName)

	time.AfterFunc(duration, func() {
		c.enableNode(nodeName)
	})
}

func (c *Chain) enableNode(nodeName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, node := range c.nodes {
		if node.name == nodeName {
			node.disabled = false
		}
	}
}
