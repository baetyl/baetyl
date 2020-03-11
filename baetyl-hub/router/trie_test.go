package router

import (
	"fmt"
	"testing"

	"github.com/256dpi/gomqtt/topic"
	"github.com/stretchr/testify/assert"
)

func TestTrieMatch(t *testing.T) {
	tcs := []struct {
		input    string
		expected []string
	}{
		{" ", []string{"#", "+", "+/#", " "}},
		{"/", []string{"#", "/", "/+", "+/", "+/+", "+/#", "/#"}},
		{"//", []string{"#", "//", "/#", "+/#"}},
		{" abc", []string{"#", "+", " abc", "+/#"}},
		{"abc", []string{"#", "+", "abc1", "abc2", "abc/#", "+/#"}},
		{"abc/", []string{"#", "abc/", "abc/#", "abc/+", "+/", "+/#", "+/+"}},
		{"/ldf", []string{"#", "/ldf", "/+", "/#", "+/+", "+/#"}},
		{"abc/ldf", []string{"#", "abc/+", "abc/#", "abc/ldf", "+/+", "+/#"}},
		{"abc/ldf/rmb/", []string{"#", "abc/ldf/rmb/", "abc/+/rmb/", "abc/#", "+/#"}},
		{"/abc/ldf/rmb", []string{"#", "/abc/ldf/rmb", "+/abc/#", "/#", "+/#"}},
		{"abc/ldf//rmb", []string{"#", "abc/ldf//rmb", "abc/#", "+/#"}},
		{"test/abc/ldf/rmb", []string{"#", "+/abc/#", "+/#"}},
	}

	root := NewTrie()

	root.Add(NewNopSinkSub(" ", 0, " ", 0, ""))
	root.Add(NewNopSinkSub(" abc", 0, " abc", 0, ""))
	root.Add(NewNopSinkSub("//", 0, "//", 0, ""))
	root.Add(NewNopSinkSub("#", 0, "#", 0, ""))
	root.Add(NewNopSinkSub("+", 0, "+", 0, ""))
	root.Add(NewNopSinkSub("+/", 0, "+/", 0, ""))
	root.Add(NewNopSinkSub("+/+", 0, "+/+", 0, ""))
	root.Add(NewNopSinkSub("/+", 0, "/+", 0, ""))
	root.Add(NewNopSinkSub("/#", 0, "/#", 0, ""))
	root.Add(NewNopSinkSub("+/#", 0, "+/#", 0, ""))
	root.Add(NewNopSinkSub("abc1", 0, "abc", 0, ""))
	root.Add(NewNopSinkSub("abc2", 0, "abc", 0, ""))
	root.Add(NewNopSinkSub("/", 0, "/", 0, ""))
	root.Add(NewNopSinkSub("/ldf", 0, "/ldf", 0, ""))
	root.Add(NewNopSinkSub("abc/", 0, "abc/", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf", 0, "abc/ldf", 0, ""))
	root.Add(NewNopSinkSub("abc/#", 0, "abc/#", 0, ""))
	root.Add(NewNopSinkSub("abc/+", 0, "abc/+", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf/rmb/", 0, "abc/ldf/rmb/", 0, ""))
	root.Add(NewNopSinkSub("/abc/ldf/rmb", 0, "/abc/ldf/rmb", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf/ /rmb", 0, "abc/ldf/ /rmb", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf//rmb", 0, "abc/ldf//rmb", 0, ""))
	root.Add(NewNopSinkSub("abc/+/rmb/", 0, "abc/+/rmb/", 0, ""))
	root.Add(NewNopSinkSub("abc/+/ldf/+/rmb", 0, "abc/+/ldf/+/rmb", 0, ""))
	root.Add(NewNopSinkSub("+/abc/#", 0, "+/abc/#", 0, ""))

	for _, tc := range tcs {
		subs := root.Match(tc.input)
		assert.Len(t, subs, len(tc.expected))
		for _, v := range subs {
			assert.Contains(t, tc.expected, v.ID())
		}
	}

	root.Remove(" ", " ")
	root.Remove(" abc", " abc")
	root.Remove("//", "//")
	root.Remove("#", "#")
	root.Remove("+", "+")
	root.Remove("+/", "+/")
	root.Remove("/+", "/+")
	root.Remove("/#", "/#")
	root.Remove("+/+", "+/+")
	root.Remove("+/#", "+/#")
	root.Remove("abc1", "abc")
	root.Remove("abc2", "abc")
	root.Remove("/", "/")
	root.Remove("/ldf", "/ldf")
	root.Remove("abc/", "abc/")
	root.Remove("abc/ldf", "abc/ldf")
	root.Remove("abc/#", "abc/#")
	root.Remove("abc/+", "abc/+")
	root.Remove("abc/ldf/rmb/", "abc/ldf/rmb/")
	root.Remove("/abc/ldf/rmb", "/abc/ldf/rmb")
	root.Remove("abc/ldf/ /rmb", "abc/ldf/ /rmb")
	root.Remove("abc/ldf//rmb", "abc/ldf//rmb")
	root.Remove("abc/+/rmb/", "abc/+/rmb/")
	root.Remove("abc/+/ldf/+/rmb", "abc/+/ldf/+/rmb")
	root.Remove("+/abc/#", "+/abc/#")

	for _, tc := range tcs {
		subs := root.Match(tc.input)
		assert.Len(t, subs, 0)
	}

	root.Remove("remove again", "remove again")
}

func TestTrieMatchUnique(t *testing.T) {
	r := NewTrie()
	// No sink
	matched := r.MatchUnique("test")
	assert.Len(t, matched, 0)
	// add all sinks
	subs := []*NopSinkSub{
		NewNopSinkSub("a", 1, "#", 1, "a11"),
		NewNopSinkSub("a", 0, "test", 0, "a00"),
		NewNopSinkSub("b", 1, "test", 1, "b11"),
		NewNopSinkSub("b", 0, "#", 0, "b00"),
		NewNopSinkSub("c", 0, "test", 0, "c00"),
		NewNopSinkSub("c", 1, "+", 1, "c11"),
		NewNopSinkSub("d", 0, "+", 0, "d00"),
		NewNopSinkSub("d", 1, "test", 1, "d11"),

		NewNopSinkSub("e", 1, "#", 1, "e11"),
		NewNopSinkSub("e", 1, "test", 0, "e10"),
		NewNopSinkSub("f", 1, "test", 1, "f11"),
		NewNopSinkSub("f", 1, "#", 0, "f10"),
		NewNopSinkSub("g", 0, "test", 0, "g00"),
		NewNopSinkSub("g", 0, "+", 1, "g01"),
		NewNopSinkSub("h", 0, "+", 0, "h00"),
		NewNopSinkSub("h", 0, "test", 1, "h01"),

		NewNopSinkSub("i", 0, "test", 0, "i0"),
		NewNopSinkSub("i", 1, "test", 0, "i1"),
		NewNopSinkSub("i", 1, "+", 1, "i2"),
		NewNopSinkSub("i", 1, "#", 0, "i3"),

		NewNopSinkSub("j", 0, "#", 0, "j0"),
		NewNopSinkSub("j", 1, "+", 0, "j1"),
		NewNopSinkSub("j", 1, "#", 0, "j2"),

		NewNopSinkSub("k", 0, "+", 0, "k0"),
		NewNopSinkSub("k", 1, "#", 0, "k1"),
		NewNopSinkSub("k", 1, "+", 0, "k2"),
	}
	for _, s := range subs {
		r.Add(s)
	}
	expected := []struct {
		id     string
		sqos   uint32
		tqos   uint32
		ttopic string
	}{
		{"a", 1, 1, "a11"},
		{"b", 1, 1, "b11"},
		{"c", 1, 1, "c11"},
		{"d", 1, 1, "d11"},
		{"e", 1, 1, "e11"},
		{"f", 1, 1, "f11"},
		{"g", 0, 1, "g01"},
		{"h", 0, 1, "h01"},
		{"i", 1, 1, "i2"},
		{"j", 1, 0, "j2"}, // ?
		{"k", 1, 0, "k1"}, // ?
	}
	matched = r.MatchUnique("test")
	assert.Len(t, matched, len(expected))
	for _, v := range expected {
		msg := fmt.Sprintf("1-%v", v)
		assert.Equal(t, v.sqos, matched[v.id].QOS(), msg)
		assert.Equal(t, v.tqos, matched[v.id].TargetQOS(), msg)
		assert.Equal(t, v.ttopic, matched[v.id].TargetTopic(), msg)
	}
	// remove 2 sub1 sub2
	r.Remove("k", "+")
	matched = r.MatchUnique("test")
	expected[len(expected)-1].ttopic = "k1"
	assert.Len(t, matched, len(expected))
	for _, v := range expected {
		msg := fmt.Sprintf("2-%v", v)
		assert.Equal(t, v.sqos, matched[v.id].QOS(), msg)
		assert.Equal(t, v.tqos, matched[v.id].TargetQOS(), msg)
		assert.Equal(t, v.ttopic, matched[v.id].TargetTopic(), msg)
	}
	r.RemoveAll("k")
	r.RemoveAll("j")
	expected = expected[:len(expected)-2]
	matched = r.MatchUnique("test")
	assert.Len(t, matched, len(expected))
	for _, v := range expected {
		msg := fmt.Sprintf("3-%v", v)
		assert.Equal(t, v.sqos, matched[v.id].QOS(), msg)
		assert.Equal(t, v.tqos, matched[v.id].TargetQOS(), msg)
		assert.Equal(t, v.ttopic, matched[v.id].TargetTopic(), msg)
	}
}

func BenchmarkTrieMatch(b *testing.B) {
	root := NewTrie()
	root.Add(NewNopSinkSub(" ", 0, " ", 0, ""))
	root.Add(NewNopSinkSub(" abc", 0, " abc", 0, ""))
	root.Add(NewNopSinkSub("//", 0, "//", 0, ""))
	root.Add(NewNopSinkSub("#", 0, "#", 0, ""))
	root.Add(NewNopSinkSub("+", 0, "+", 0, ""))
	root.Add(NewNopSinkSub("+/", 0, "+/", 0, ""))
	root.Add(NewNopSinkSub("+/+", 0, "+/+", 0, ""))
	root.Add(NewNopSinkSub("/+", 0, "/+", 0, ""))
	root.Add(NewNopSinkSub("/#", 0, "/#", 0, ""))
	root.Add(NewNopSinkSub("+/#", 0, "+/#", 0, ""))
	root.Add(NewNopSinkSub("abc1", 0, "abc", 0, ""))
	root.Add(NewNopSinkSub("abc2", 0, "abc", 0, ""))
	root.Add(NewNopSinkSub("/", 0, "/", 0, ""))
	root.Add(NewNopSinkSub("/ldf", 0, "/ldf", 0, ""))
	root.Add(NewNopSinkSub("abc/", 0, "abc/", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf", 0, "abc/ldf", 0, ""))
	root.Add(NewNopSinkSub("abc/#", 0, "abc/#", 0, ""))
	root.Add(NewNopSinkSub("abc/+", 0, "abc/+", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf/rmb/", 0, "abc/ldf/rmb/", 0, ""))
	root.Add(NewNopSinkSub("/abc/ldf/rmb", 0, "/abc/ldf/rmb", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf/ /rmb", 0, "abc/ldf/ /rmb", 0, ""))
	root.Add(NewNopSinkSub("abc/ldf//rmb", 0, "abc/ldf//rmb", 0, ""))
	root.Add(NewNopSinkSub("abc/+/rmb/", 0, "abc/+/rmb/", 0, ""))
	root.Add(NewNopSinkSub("abc/+/ldf/+/rmb", 0, "abc/+/ldf/+/rmb", 0, ""))
	root.Add(NewNopSinkSub("+/abc/#", 0, "+/abc/#", 0, ""))

	b.ResetTimer()
	topic := "abc/ldf/rmb/"
	for index := 0; index < b.N; index++ {
		root.Match(topic)
	}
}

func BenchmarkTreeMatch256dpi(b *testing.B) {
	root := topic.NewStandardTree()
	root.Add(" ", NewNopSinkSub(" ", 0, " ", 0, ""))
	root.Add(" abc", NewNopSinkSub(" abc", 0, " abc", 0, ""))
	root.Add("//", NewNopSinkSub("//", 0, "//", 0, ""))
	root.Add("#", NewNopSinkSub("#", 0, "#", 0, ""))
	root.Add("+", NewNopSinkSub("+", 0, "+", 0, ""))
	root.Add("+/", NewNopSinkSub("+/", 0, "+/", 0, ""))
	root.Add("+/+", NewNopSinkSub("+/+", 0, "+/+", 0, ""))
	root.Add("/+", NewNopSinkSub("/+", 0, "/+", 0, ""))
	root.Add("/#", NewNopSinkSub("/#", 0, "/#", 0, ""))
	root.Add("+/#", NewNopSinkSub("+/#", 0, "+/#", 0, ""))
	root.Add("abc", NewNopSinkSub("abc1", 0, "abc", 0, ""))
	root.Add("abc", NewNopSinkSub("abc2", 0, "abc", 0, ""))
	root.Add("/", NewNopSinkSub("/", 0, "/", 0, ""))
	root.Add("/ldf", NewNopSinkSub("/ldf", 0, "/ldf", 0, ""))
	root.Add("abc/", NewNopSinkSub("abc/", 0, "abc/", 0, ""))
	root.Add("abc/ldf", NewNopSinkSub("abc/ldf", 0, "abc/ldf", 0, ""))
	root.Add("abc/#", NewNopSinkSub("abc/#", 0, "abc/#", 0, ""))
	root.Add("abc/+", NewNopSinkSub("abc/+", 0, "abc/+", 0, ""))
	root.Add("abc/ldf/rmb/", NewNopSinkSub("abc/ldf/rmb/", 0, "abc/ldf/rmb/", 0, ""))
	root.Add("/abc/ldf/rmb", NewNopSinkSub("/abc/ldf/rmb", 0, "/abc/ldf/rmb", 0, ""))
	root.Add("abc/ldf/ /rmb", NewNopSinkSub("abc/ldf/ /rmb", 0, "abc/ldf/ /rmb", 0, ""))
	root.Add("abc/ldf//rmb", NewNopSinkSub("abc/ldf//rmb", 0, "abc/ldf//rmb", 0, ""))
	root.Add("abc/+/rmb/", NewNopSinkSub("abc/+/rmb/", 0, "abc/+/rmb/", 0, ""))
	root.Add("abc/+/ldf/+/rmb", NewNopSinkSub("abc/+/ldf/+/rmb", 0, "abc/+/ldf/+/rmb", 0, ""))
	root.Add("+/abc/#", NewNopSinkSub("+/abc/#", 0, "+/abc/#", 0, ""))

	b.ResetTimer()
	topic := "abc/ldf/rmb/"
	for index := 0; index < b.N; index++ {
		root.Match(topic)
	}
}
