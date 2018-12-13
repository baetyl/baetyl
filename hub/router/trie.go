package router

import (
	"strings"
	"sync"

	"github.com/baidu/openedge/hub/common"
)

// Trie topic tree of *common.SinkSubs
type Trie struct {
	root *node
	sync.RWMutex
}

// NewTrie creates a trie
func NewTrie() *Trie {
	return &Trie{
		root: newNode(),
	}
}

// Add adds subscription to trie
// #、+、/
func (t *Trie) Add(sub SinkSub) {
	t.Lock()
	defer t.Unlock()

	current := t.root
	nodes := strings.Split(sub.Topic(), "/")
	for _, n := range nodes {
		if _, ok := current.children[n]; !ok {
			current.children[n] = newNode()
		}
		current = current.children[n]
	}
	current.sinksubs[sub.ID()] = sub
}

// Remove removes subscription to trie
// #、+、/
func (t *Trie) Remove(id, topic string) {
	t.Lock()
	defer t.Unlock()

	stack := newDeque()
	stack.Push(t.root)
	nodes := strings.Split(topic, "/")
	for _, n := range nodes {
		current := stack.Peek().(*node).children[n]
		if current == nil {
			return
		}
		stack.Push(current)
	}

	t.doRemove(stack, nodes, id)
}

// RemoveAll removes all subscriptions by id
func (t *Trie) RemoveAll(id string) {
	t.Lock()
	defer t.Unlock()

	// TODO: remove all subscriptions by id and add unit test @wangmengtao
	var n *node
	queue := make([]*node, 0)
	queue = append(queue, t.root)
	for len(queue) > 0 {
		n = queue[0]
		delete(n.sinksubs, id)
		for _, c := range n.children {
			queue = append(queue, c)
		}
		queue = queue[1:]
	}
}

func (t *Trie) doRemove(stack *deque, nodes []string, uid string) {
	pos := len(nodes)
	delete(stack.Peek().(*node).sinksubs, uid)

	// Remove empty path in the trie
	for index := pos - 1; index >= 0; index-- {
		if stack.Peek().(*node).isEmpty() && stack.Pop() != t.root {
			delete(stack.Peek().(*node).children, nodes[index])
		}
	}

	return
}

// IsMatch IsMatch
// TODO: improve
func (t *Trie) IsMatch(topic string) (bool, uint32) {
	ss := t.Match(topic)
	ok := len(ss) != 0
	qos := uint32(0)
	for _, s := range ss {
		if s.QOS() > qos {
			qos = s.QOS()
		}
	}
	return ok, qos
}

// MatchUnique matches subscriptions merged by id
// TODO: improve
func (t *Trie) MatchUnique(topic string) map[string]SinkSub {
	matches := t.Match(topic)
	if len(matches) == 0 {
		return nil
	}
	// TODO: move dup check logic into Trie.Match
	var ok bool
	var sub, dup SinkSub
	subs := make(map[string]SinkSub)
	for _, sub = range matches {
		dup, ok = subs[sub.ID()]
		if !ok {
			subs[sub.ID()] = sub
		} else if sub.QOS() > dup.QOS() {
			// pick sub with big source qos
			subs[sub.ID()] = sub
		} else if sub.QOS() == dup.QOS() && sub.TargetQOS() > dup.TargetQOS() {
			// pick sub with big target qos if source qos equals
			subs[sub.ID()] = sub
		}
	}
	return subs
}

// Match Matches topic
func (t *Trie) Match(topic string) []SinkSub {
	t.RLock()
	defer t.RUnlock()

	matched := make([]SinkSub, 0)
	nodes := strings.Split(topic, "/")

	matchedNodes := t.doMatch(nodes)
	for _, node := range matchedNodes {
		for _, v := range node.sinksubs {
			matched = append(matched, v)
		}
	}

	return matched
}

func (t *Trie) doMatch(subjects []string) []*node {
	nodeQueue := newDeque()
	var matchedNodes []*node
	flag := false

	// Use breadth-first-search to get all matched node
	nodeQueue.Offer(t.root)
	for _, subject := range subjects {
		if !flag {
			amount := nodeQueue.Len()
			for index := amount; index > 0; index-- {
				nodeQueue.Poll().(*node).attachMultipleMatch(&matchedNodes).attachSingleMatch(nodeQueue, subject)
			}
			if nodeQueue.Empty() {
				flag = true
			}
		}
	}

	// Add the "#" match with zero level, such as "a/#" matching "a"
	for !nodeQueue.Empty() {
		matchedNodes = append(matchedNodes, nodeQueue.Poll().(*node).attachMultipleMatch(&matchedNodes))
	}

	return matchedNodes
}

func (n *node) attachMultipleMatch(nodes *[]*node) *node {
	node := n.children[common.MultipleWildCard]
	if node != nil {
		*nodes = append(*nodes, node)
	}

	return n
}

func (n *node) attachSingleMatch(nodeQueue *deque, subject string) *node {
	if strings.EqualFold(subject, common.MultipleWildCard) {
		return nil
	}

	node := n.children[subject]
	if node != nil {
		nodeQueue.Offer(node)
	}

	node = n.children[common.SingleWildCard]
	if node != nil {
		nodeQueue.Offer(node)
	}

	return n
}
