package oracle

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/brunotm/tact"
	"github.com/brunotm/tact/js" // load oracle driver
	_ "github.com/mattn/go-oci8"
)

const (
	oracleDriver = "oci8"
)

// Client creates a new ssh client
func Client(ctx *tact.Context) (client *sql.DB, err error) {
	connStr := fmt.Sprintf("%s/%s@%s:%s/%s",
		ctx.Node().DBUser,
		ctx.Node().DBPassword,
		ctx.Node().NetAddr,
		ctx.Node().DBPort,
		ctx.Node().HostName)
	client, err = sql.Open(oracleDriver, connStr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// SingleQuery collector base
func SingleQuery(ctx *tact.Context, query string) (events <-chan []byte) {
	outCh := make(chan []byte)
	go oracleQuery(ctx, query, outCh)
	return outCh
}

func oracleQuery(ctx *tact.Context, query string, outCh chan<- []byte) {
	defer close(outCh)
	client, err := Client(ctx)
	if err != nil {
		ctx.LogError(err.Error())
		return
	}
	defer client.Close()

	rows, err := client.QueryContext(ctx.Context(), query)
	if err != nil {
		ctx.LogError(err.Error())
		return
	}
	defer rows.Close()

	colnames, err := rows.Columns()
	if err != nil {
		ctx.LogError("oracle: error getting column names: ", err.Error())
		return
	}

	// Create parallel placeholders for interface{} pointers and values
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
			event, _ = js.Set(event, values[cn], strings.ToLower(colnames[cn]))
		}
		tact.WrapCtxSend(ctx.Context(), outCh, event)
	}
}
