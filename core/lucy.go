package core

import (
	internal "github.com/supercmmetry/lucy/internal"
)

type Lucy struct {
	Engine  internal.Layer
	db      *internal.Database
	runtime internal.QueryRuntime
}

func (l *Lucy) DB() *internal.Database {
	l.Engine = (&internal.QueryEngine{}).NewQueryEngine()
	l.db = &internal.Database{}
	l.Engine.AttachTo(l.db)
	l.db.AddRuntime(l.runtime)
	return l.db
}

func (l *Lucy) AddRuntime(rt internal.QueryRuntime) {
	l.runtime = rt
}