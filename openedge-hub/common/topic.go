package common

import (
	"strings"
)

type subject interface {
	getValue() string
}

// normalSubject
type normalSubject struct {
	value string
}

// singleWildcardSubject
type singleWildcardSubject struct {
	value string
}

// multipleWildcardSubject
type multipleWildcardSubject struct {
	value string
}

func (n normalSubject) getValue() string {
	return n.value
}

func (s singleWildcardSubject) getValue() string {
	return s.value
}

func (m multipleWildcardSubject) getValue() string {
	return m.value
}

// getTopicSubjects convert topic to []subject by subject kind
func getTopicSubjects(topic string) []subject {
	var subjectList []subject
	var topicList = strings.Split(topic, "/")
	for _, value := range topicList {
		switch value {
		case SingleWildCard:
			subjectList = append(subjectList, interface{}(singleWildcardSubject{value}).(subject))
			break
		case MultipleWildCard:
			subjectList = append(subjectList, interface{}(multipleWildcardSubject{value}).(subject))
			break
		default:
			subjectList = append(subjectList, interface{}(normalSubject{value}).(subject))
			break
		}
	}
	return subjectList
}

// TopicIsMatch check the given topicRule is matched the given topic or not
func TopicIsMatch(topic string, topicRule string) bool {
	topicSubjects := getTopicSubjects(topic)
	topicRuleSubjects := getTopicSubjects(topicRule)
	topicSubjectsLength := len(topicSubjects)
	topicRuleSubjectsLength := len(topicRuleSubjects)
	var minLength int
	if topicSubjectsLength < topicRuleSubjectsLength {
		minLength = topicSubjectsLength
	} else {
		minLength = topicRuleSubjectsLength
	}
	for i := 0; i < minLength; i++ {
		topicSubject := topicSubjects[i]
		topicRuleSubject := topicRuleSubjects[i]
		if strings.Compare(topicRuleSubject.getValue(), MultipleWildCard) == 0 {
			return true
		}
		if strings.Compare(topicRuleSubject.getValue(), SingleWildCard) != 0 &&
			strings.Compare(topicRuleSubject.getValue(), topicSubject.getValue()) != 0 {
			return false
		}
	}
	if topicSubjectsLength > minLength {
		return false
	}
	if topicRuleSubjectsLength > minLength &&
		strings.Compare(topicRuleSubjects[minLength].getValue(), MultipleWildCard) != 0 {
		return false
	}
	return true
}

// ContainsWildcard check topic contains wildCard("#" or "+") or not
func ContainsWildcard(topic string) bool {
	return strings.Contains(topic, SingleWildCard) || strings.Contains(topic, MultipleWildCard)
}

// isSysTopic check topic is SysTopic or not
func isSysTopic(topic string) bool {
	return strings.HasPrefix(topic, SysCmdPrefix)
}

// PubTopicValidate validate MQTT publish topic
func PubTopicValidate(topic string) bool {
	if topic == "" {
		return false
	}
	if len(topic) > MaxTopicNameLen || strings.Contains(topic, "\u0000") ||
		strings.Count(topic, TopicSeparator) > MaxSlashCount {
		return false
	}
	if ContainsWildcard(topic) {
		return false
	}
	if isSysTopic(topic) {
		return false
	}
	return true
}

// SubTopicValidate validate MQTT subscribe topic
func SubTopicValidate(topic string) bool {
	if topic == "" {
		return false
	}
	if len(topic) > MaxTopicNameLen || strings.Contains(topic, "\u0000") ||
		strings.Count(topic, TopicSeparator) > MaxSlashCount {
		return false
	}
	if isSysTopic(topic) {
		return false
	}
	splited := strings.Split(topic, TopicSeparator)
	for index := 0; index < len(splited); index++ {
		s := splited[index]
		if strings.EqualFold(s, MultipleWildCard) {
			// check that multi is the last symbol
			if index != len(splited)-1 {
				return false
			}
		} else if strings.Contains(s, MultipleWildCard) {
			return false
		} else if !strings.EqualFold(s, SingleWildCard) &&
			strings.Contains(s, SingleWildCard) {
			return false
		}
	}
	return true
}
