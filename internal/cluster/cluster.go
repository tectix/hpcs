package cluster

import (
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/tectix/hpcs/internal/config"
	"github.com/tectix/hpcs/internal/hash"
)

type NodeStatus int

const (
	NodeStatusUnknown NodeStatus = iota
	NodeStatusAlive
	NodeStatusDead
)

type Node struct {
	ID       string
	Address  string
	Status   NodeStatus
	LastSeen time.Time
}

type Cluster struct {
	mu       sync.RWMutex
	selfID   string
	selfAddr string
	nodes    map[string]*Node
	ring     *hash.Ring
	cfg      *config.ClusterConfig
	logger   *zap.Logger
}

func New(selfID, selfAddr string, cfg *config.ClusterConfig, logger *zap.Logger) *Cluster {
	ring := hash.NewRing(cfg.VirtualNodes)
	
	cluster := &Cluster{
		selfID:   selfID,
		selfAddr: selfAddr,
		nodes:    make(map[string]*Node),
		ring:     ring,
		cfg:      cfg,
		logger:   logger,
	}
	
	cluster.addNode(selfID, selfAddr, NodeStatusAlive)
	
	return cluster
}

func (c *Cluster) Start() error {
	if !c.cfg.Enabled {
		c.logger.Info("Cluster mode disabled")
		return nil
	}
	
	c.logger.Info("Starting cluster", zap.String("self_id", c.selfID))
	
	for _, nodeAddr := range c.cfg.Nodes {
		if nodeAddr != c.selfAddr {
			nodeID := fmt.Sprintf("node_%s", nodeAddr)
			c.addNode(nodeID, nodeAddr, NodeStatusUnknown)
		}
	}
	
	go c.healthCheckLoop()
	
	return nil
}

func (c *Cluster) GetNode(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.ring.GetNode(key)
}

func (c *Cluster) GetNodes(key string, count int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.ring.GetNodes(key, count)
}

func (c *Cluster) GetReplicaNodes(key string) []string {
	return c.GetNodes(key, c.cfg.ReplicaCount)
}

func (c *Cluster) IsLocalKey(key string) bool {
	node := c.GetNode(key)
	return node == c.selfID
}

func (c *Cluster) GetAliveNodes() []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	var alive []*Node
	for _, node := range c.nodes {
		if node.Status == NodeStatusAlive {
			alive = append(alive, node)
		}
	}
	return alive
}

func (c *Cluster) GetNodeByID(id string) (*Node, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	node, exists := c.nodes[id]
	return node, exists
}

func (c *Cluster) addNode(id, addr string, status NodeStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	node := &Node{
		ID:       id,
		Address:  addr,
		Status:   status,
		LastSeen: time.Now(),
	}
	
	c.nodes[id] = node
	
	if status == NodeStatusAlive {
		c.ring.AddNode(id)
		c.logger.Info("Node added to cluster", 
			zap.String("node_id", id),
			zap.String("address", addr))
	}
}

func (c *Cluster) removeNode(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if node, exists := c.nodes[id]; exists {
		c.ring.RemoveNode(id)
		delete(c.nodes, id)
		c.logger.Info("Node removed from cluster",
			zap.String("node_id", id),
			zap.String("address", node.Address))
	}
}

func (c *Cluster) updateNodeStatus(id string, status NodeStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if node, exists := c.nodes[id]; exists {
		oldStatus := node.Status
		node.Status = status
		node.LastSeen = time.Now()
		
		if oldStatus != status {
			if status == NodeStatusAlive {
				c.ring.AddNode(id)
				c.logger.Info("Node marked as alive",
					zap.String("node_id", id))
			} else if status == NodeStatusDead {
				c.ring.RemoveNode(id)
				c.logger.Info("Node marked as dead",
					zap.String("node_id", id))
			}
		}
	}
}

func (c *Cluster) healthCheckLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.performHealthChecks()
		}
	}
}

func (c *Cluster) performHealthChecks() {
	c.mu.RLock()
	nodes := make([]*Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		if node.ID != c.selfID {
			nodes = append(nodes, node)
		}
	}
	c.mu.RUnlock()
	
	for _, node := range nodes {
		go c.checkNodeHealth(node)
	}
}

func (c *Cluster) checkNodeHealth(node *Node) {
	conn, err := net.DialTimeout("tcp", node.Address, 2*time.Second)
	if err != nil {
		if node.Status == NodeStatusAlive {
			c.updateNodeStatus(node.ID, NodeStatusDead)
		}
		return
	}
	defer conn.Close()
	
	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	
	buffer := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(buffer)
	
	if err != nil {
		if node.Status == NodeStatusAlive {
			c.updateNodeStatus(node.ID, NodeStatusDead)
		}
	} else {
		if node.Status != NodeStatusAlive {
			c.updateNodeStatus(node.ID, NodeStatusAlive)
		}
	}
}