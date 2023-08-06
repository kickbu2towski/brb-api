package main

import (
	"github.com/kickbu2towski/brb-api/internal/data"
)

func Includes(input []string, key string) bool {
	var exists bool
	for _, v := range input {
		if v == key {
			exists = true
			break
		}
	}
	return exists
}

func GetMessageEvent(event map[string]any) (*data.Event, error) {
	var m data.Event
	if t, ok := (event["type"]).(string); ok {
		m.Type = t
	}
	if p, ok := (event["payload"]).(map[string]any); ok {
		m.Payload = p
	}
	return &m, nil
}

func GetBroadcastTo(event map[string]any) ([]string, error) {
	var bc []string
	if b, ok := (event["broadcastTo"]).([]any); ok {
		for _, v := range b {
			c, ok := (v).(string)
			if ok {
				bc = append(bc, c)
			}
		}
	}
	return bc, nil
}

func Filter(items []any, predicate func(item string) bool) []any {
	var res []any
	for _, v := range items {
		s := v.(string)
		if predicate(s) {
			res = append(res, v)
		}
	}
	return res
}
