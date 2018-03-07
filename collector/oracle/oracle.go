package oracle

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	// load oracle driver
	_ "github.com/mattn/go-oci8"
)

const (
	oracleDriver = "oci8"
)

// Query collector base
func singleQuery(session *tact.Session, query string) (events <-chan []byte) {
	outCh := make(chan []byte)
	go oracleQuery(session, query, outCh)
	return outCh
}

func oracleQuery(session *tact.Session, query string, outCh chan<- []byte) {
	defer close(outCh)
	db, err := sql.Open(oracleDriver,
		fmt.Sprintf("%s/%s@%s:%s/%s",
			session.Node().DBUser,
			session.Node().DBPassword,
			session.Node().NetAddr,
			session.Node().DBPort,
			session.Node().HostName))
	if err != nil {
		session.LogErr(err.Error())
		return
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		session.LogErr(err.Error())
		return
	}
	defer rows.Close()

	colnames, err := rows.Columns()
	if err != nil {
		session.LogErr("oracle: error getting column names: ", err.Error())
		return
	}

	// Create parallel placeholders for interface{} pointers and volues
	// This is needed to allow flexibility for not knowing column
	// number and and names ahead of time
	colcnt := len(colnames)
	values := make([]interface{}, colcnt)
	valuePtrs := make([]interface{}, colcnt)
	for i := range colnames {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		var event []byte
		rows.Scan(valuePtrs...)
		for cn := range colnames {
			event, _ = rexon.JSONSet(event, values[cn], strings.ToLower(colnames[cn]))
		}
		tact.WrapCtxSend(session.Context(), outCh, event)
	}
}
