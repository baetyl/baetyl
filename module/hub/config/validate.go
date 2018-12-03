package config

import (
	"github.com/baidu/openedge/module/hub/common"
	"github.com/baidu/openedge/module/hub/utils"
	"github.com/deckarep/golang-set"
	"github.com/juju/errors"
)

// principalsValidate validate principals config is valid or not
func principalsValidate(v interface{}, param string) error {
	principals := v.([]Principal)
	err := userValidate(principals)
	if err != nil {
		return errors.Trace(err)
	}
	for _, principal := range principals {
		for _, permission := range principal.Permissions {
			for _, permit := range permission.Permits {
				if !common.SubTopicValidate(permit) {
					return errors.Errorf("%s topic(%s) invalid", permission.Action, permit)
				}
			}
		}
	}
	return nil
}

// userValidate validate username duplicate or not
func userValidate(principals []Principal) error {
	userMap := make(map[string]struct{})
	for _, principal := range principals {
		if _, ok := userMap[principal.Username]; ok {
			return errors.Errorf("User name (%s) duplicate", principal.Username)
		}
		userMap[principal.Username] = struct{}{}
	}

	return nil
}

// subscriptionsValidate check the subscriptions config is valid or not
func subscriptionsValidate(v interface{}, param string) error {
	subscriptions := v.([]Subscription)
	for _, s := range subscriptions {
		if !common.SubTopicValidate(s.Source.Topic) {
			return errors.Errorf("[%+v] source topic invalid", s.Source)
		}
		if !common.PubTopicValidate(s.Target.Topic) {
			return errors.Errorf("[%+v] target topic invalid", s.Target)
		}
	}
	// duplicate source and target config validate
	_, edges := getVertexEdges(subscriptions)
	if dupSubscriptionValidate(edges) {
		return errors.Errorf("Duplicate source and target config")
	}
	// cycle found in source and target config
	return cycleFound(edges)
}

// cycleFound is used for finding cycle
func cycleFound(edges [][2]string) error {
	normalEdges, wildcardEdges := classifyEdges(edges)
	// normal source && target config cycle detect
	if !canFinish(normalEdges) {
		return errors.Errorf("Found cycle in source and target config")
	}
	// wildcard source && target config cycle detect
	targetVertexs := getTargetVertexs(edges)
	for _, e := range wildcardEdges {
		for _, v := range targetVertexs {
			if common.TopicIsMatch(v, e[0]) {
				e[0] = v
				normalEdges = append(normalEdges, e)
				if !canFinish(normalEdges) {
					return errors.Errorf("Found cycle in source and target config")
				}
			}
		}
	}
	return nil
}

// canFinish detect a directed graph has cycle or not
// vertexs store graph vertex info, edges store graph edge info
func canFinish(edges [][2]string) bool {
	if len(edges[0]) == 0 || len(edges) == 0 {
		return true
	}
	vertexs := edges2Vertexs(edges)
	graph := make(map[string][]string)
	for i := range edges {
		graph[edges[i][1]] = append(graph[edges[i][1]], edges[i][0])
	}
	path := make([]bool, len(vertexs))
	visited := make([]bool, len(vertexs))
	for i := range vertexs {
		if visited[i] {
			continue
		}
		if hasCycle(vertexs, vertexs[i], graph, path, visited) {
			return false
		}
	}
	return true
}

// hasCycle check the directed graph has cycle or not
// graph store some vertex which is associated with the given vertex
// visited represents an vertex visit-status, if true, visited; else none visited
func hasCycle(vertexs []string, start string, graph map[string][]string, path []bool, visited []bool) bool {
	for i := range graph[start] {
		if visited[getPosition(graph[start][i], vertexs)] {
			continue
		}
		if path[getPosition(graph[start][i], vertexs)] {
			return true
		}
		path[getPosition(graph[start][i], vertexs)] = true
		if hasCycle(vertexs, graph[start][i], graph, path, visited) {
			return true
		}
		path[getPosition(graph[start][i], vertexs)] = false
	}
	visited[getPosition(start, vertexs)] = true
	return false
}

// getPosition return the given data's index of the given vertexs
func getPosition(data string, vertexs []string) int {
	for pos := 0; pos < len(vertexs); pos++ {
		if vertexs[pos] == data {
			return pos
		}
	}
	return -1
}

// getVertexEdges generate vertexs && edges
func getVertexEdges(subscriptions []Subscription) ([]string, [][2]string) {
	vertexs := make(map[string]struct{})
	edges := make([][2]string, 0)
	for _, s := range subscriptions {
		st := s.Source.Topic
		tt := s.Target.Topic
		vertexs[st] = struct{}{}
		vertexs[tt] = struct{}{}
		edges = append(edges, [2]string{st, tt})
	}
	return utils.GetKeys(vertexs), edges
}

// classifyEdges generate normalEges && wildcardEdges
func classifyEdges(edges [][2]string) ([][2]string, [][2]string) {
	normalEdges := make([][2]string, 0)
	wildcardEdges := make([][2]string, 0)
	for _, e := range edges {
		if common.ContainsWildcard(e[0]) {
			wildcardEdges = append(wildcardEdges, e)
		} else {
			normalEdges = append(normalEdges, e)
		}
	}
	return normalEdges, wildcardEdges
}

// edges2Vertexs generate vertexs from given edges
func edges2Vertexs(edges [][2]string) []string {
	vertexs := make(map[string]struct{})
	for _, e := range edges {
		vertexs[e[0]] = struct{}{}
		vertexs[e[1]] = struct{}{}
	}
	return utils.GetKeys(vertexs)
}

// getTargetVertexs generate target vertexs from given edges
func getTargetVertexs(edges [][2]string) []string {
	vertexs := make(map[string]struct{})
	for _, e := range edges {
		vertexs[e[1]] = struct{}{}
	}
	return utils.GetKeys(vertexs)
}

// dupSubscriptionValidate check subscription config has duplicate config or not
func dupSubscriptionValidate(edges [][2]string) bool {
	_edges := mapset.NewSet()
	for _, element := range edges {
		_edges.Add(element)
	}
	if _edges.Cardinality() != len(edges) {
		return true
	}
	return false
}
