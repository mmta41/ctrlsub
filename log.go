package main

import (
	log "github.com/sirupsen/logrus"
)

type LogFormat struct {
	Status   bool
	Target   string
	Cname    string
	Provider string
	IsDone   bool
}

func (l *LogFormat) toField(err error) log.Fields {
	f := log.Fields{
		"status":   l.Status,
		"target":   l.Target,
		"cname":    l.Cname,
		"provider": l.Provider,
		"done":     l.IsDone,
	}
	if err != nil {
		f["error"] = err
	}
	return f
}

func (l *LogFormat) Done() *LogFormat {
	l.IsDone = true
	return l
}

func SortLog(list []string) {
	if len(list) < 5 {
		return
	}

	order := map[string]int{
		"status":   0,
		"done":     1,
		"target":   2,
		"cname":    3,
		"provider": 4,
		"error":    5,
		"msg":      6,
		"level":    7,
	}

	for i, l := range list {
		for j, l2 := range list {
			if order[l] > order[l2] {
				t := list[i]
				list[i] = list[j]
				list[j] = t
			}
		}
	}
}
