package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/russellslater/protohackers/internal/db"
)

type DbCommand interface {
	execute() (result string)
}

type InsertCommand struct {
	db    *db.UnusualDatabase
	key   string
	value string
}

func (i *InsertCommand) execute() (result string) {
	log.Printf("setting key %q to %q", i.key, i.value)

	i.db.Set(i.key, i.value)
	result = ""

	return
}

type RetrieveCommand struct {
	db  *db.UnusualDatabase
	key string
}

func (r *RetrieveCommand) execute() (result string) {
	value, _ := r.db.Get(r.key)
	result = fmt.Sprintf("%s=%s", r.key, value)

	log.Printf("retrieving %q for %q", value, r.key)

	return
}

func NewDbCommand(db *db.UnusualDatabase, cmd string) DbCommand {
	eqIndex := strings.Index(cmd, "=")
	if eqIndex != -1 {
		key := cmd[:eqIndex]
		value := cmd[eqIndex+1:]
		return &InsertCommand{
			db:    db,
			key:   key,
			value: value,
		}
	} else {
		return &RetrieveCommand{
			db:  db,
			key: cmd,
		}
	}
}
