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
	hasError := false
	for _, l := range list {
		if l == "error" {
			hasError = true
			break
		}
	}
	list[0] = "status"
	list[1] = "target"
	list[2] = "cname"
	list[3] = "provider"
	list[4] = "done"
	if hasError {
		list[5] = "error"
	}
}
