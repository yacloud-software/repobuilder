package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBTrackerLog
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence trackerlog_seq;

Main Table:

 CREATE TABLE trackerlog (id integer primary key default nextval('trackerlog_seq'),createrequestid bigint not null  ,createtype integer not null  ,logmessage text not null  ,publicmessage text not null  ,occured integer not null  ,success boolean not null  ,task text not null  );

Alter statements:
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS createtype integer not null default 0;
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS logmessage text not null default '';
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS publicmessage text not null default '';
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS occured integer not null default 0;
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS success boolean not null default false;
ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS task text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE trackerlog_archive (id integer unique not null,createrequestid bigint not null,createtype integer not null,logmessage text not null,publicmessage text not null,occured integer not null,success boolean not null,task text not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/repobuilder"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBTrackerLog *DBTrackerLog
)

type DBTrackerLog struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBTrackerLog() *DBTrackerLog {
	if default_def_DBTrackerLog != nil {
		return default_def_DBTrackerLog
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBTrackerLog(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBTrackerLog = res
	return res
}
func NewDBTrackerLog(db *sql.DB) *DBTrackerLog {
	foo := DBTrackerLog{DB: db}
	foo.SQLTablename = "trackerlog"
	foo.SQLArchivetablename = "trackerlog_archive"
	return &foo
}

func (a *DBTrackerLog) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBTrackerLog) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

// archive. It is NOT transactionally save.
func (a *DBTrackerLog) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBTrackerLog", "insert into "+a.SQLArchivetablename+" (id,createrequestid, createtype, logmessage, publicmessage, occured, success, task) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.CreateRequestID, p.CreateType, p.LogMessage, p.PublicMessage, p.Occured, p.Success, p.Task)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBTrackerLog) buildSaveMap(ctx context.Context, p *savepb.TrackerLog) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["createrequestid"] = a.get_col_from_proto(p, "createrequestid")
	res["createtype"] = a.get_col_from_proto(p, "createtype")
	res["logmessage"] = a.get_col_from_proto(p, "logmessage")
	res["publicmessage"] = a.get_col_from_proto(p, "publicmessage")
	res["occured"] = a.get_col_from_proto(p, "occured")
	res["success"] = a.get_col_from_proto(p, "success")
	res["task"] = a.get_col_from_proto(p, "task")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBTrackerLog) Save(ctx context.Context, p *savepb.TrackerLog) (uint64, error) {
	qn := "save_DBTrackerLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBTrackerLog) SaveWithID(ctx context.Context, p *savepb.TrackerLog) error {
	qn := "insert_DBTrackerLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBTrackerLog) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.TrackerLog) (uint64, error) {
	// Save (and use database default ID generation)

	var rows *gosql.Rows
	var e error

	q_cols := ""
	q_valnames := ""
	q_vals := make([]interface{}, 0)
	deli := ""
	i := 0
	// build the 2 parts of the query (column names and value names) as well as the values themselves
	for colname, val := range smap {
		q_cols = q_cols + deli + colname
		i++
		q_valnames = q_valnames + deli + fmt.Sprintf("$%d", i)
		q_vals = append(q_vals, val)
		deli = ","
	}
	rows, e = a.DB.QueryContext(ctx, queryname, "insert into "+a.SQLTablename+" ("+q_cols+") values ("+q_valnames+") returning id", q_vals...)
	if e != nil {
		return 0, a.Error(ctx, queryname, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, queryname, errors.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, queryname, errors.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

func (a *DBTrackerLog) Update(ctx context.Context, p *savepb.TrackerLog) error {
	qn := "DBTrackerLog_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set createrequestid=$1, createtype=$2, logmessage=$3, publicmessage=$4, occured=$5, success=$6, task=$7 where id = $8", a.get_CreateRequestID(p), a.get_CreateType(p), a.get_LogMessage(p), a.get_PublicMessage(p), a.get_Occured(p), a.get_Success(p), a.get_Task(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBTrackerLog) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBTrackerLog_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBTrackerLog) ByID(ctx context.Context, p uint64) (*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No TrackerLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) TrackerLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBTrackerLog) TryByID(ctx context.Context, p uint64) (*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) TrackerLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBTrackerLog) ByIDs(ctx context.Context, p []uint64) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBTrackerLog) All(ctx context.Context) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBTrackerLog" rows with matching CreateRequestID
func (a *DBTrackerLog) ByCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching CreateRequestID
func (a *DBTrackerLog) ByMultiCreateRequestID(ctx context.Context, p []uint64) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching CreateType
func (a *DBTrackerLog) ByCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching CreateType
func (a *DBTrackerLog) ByMultiCreateType(ctx context.Context, p []uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching LogMessage
func (a *DBTrackerLog) ByLogMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLogMessage"
	l, e := a.fromQuery(ctx, qn, "logmessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching LogMessage
func (a *DBTrackerLog) ByMultiLogMessage(ctx context.Context, p []string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLogMessage"
	l, e := a.fromQuery(ctx, qn, "logmessage in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeLogMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeLogMessage"
	l, e := a.fromQuery(ctx, qn, "logmessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching PublicMessage
func (a *DBTrackerLog) ByPublicMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByPublicMessage"
	l, e := a.fromQuery(ctx, qn, "publicmessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPublicMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching PublicMessage
func (a *DBTrackerLog) ByMultiPublicMessage(ctx context.Context, p []string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByPublicMessage"
	l, e := a.fromQuery(ctx, qn, "publicmessage in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPublicMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikePublicMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikePublicMessage"
	l, e := a.fromQuery(ctx, qn, "publicmessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPublicMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching Occured
func (a *DBTrackerLog) ByOccured(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByOccured"
	l, e := a.fromQuery(ctx, qn, "occured = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByOccured: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching Occured
func (a *DBTrackerLog) ByMultiOccured(ctx context.Context, p []uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByOccured"
	l, e := a.fromQuery(ctx, qn, "occured in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByOccured: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeOccured(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeOccured"
	l, e := a.fromQuery(ctx, qn, "occured ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByOccured: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching Success
func (a *DBTrackerLog) BySuccess(ctx context.Context, p bool) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching Success
func (a *DBTrackerLog) ByMultiSuccess(ctx context.Context, p []bool) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeSuccess"
	l, e := a.fromQuery(ctx, qn, "success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching Task
func (a *DBTrackerLog) ByTask(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByTask"
	l, e := a.fromQuery(ctx, qn, "task = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTask: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with multiple matching Task
func (a *DBTrackerLog) ByMultiTask(ctx context.Context, p []string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByTask"
	l, e := a.fromQuery(ctx, qn, "task in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTask: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeTask(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeTask"
	l, e := a.fromQuery(ctx, qn, "task ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTask: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBTrackerLog) get_ID(p *savepb.TrackerLog) uint64 {
	return uint64(p.ID)
}

// getter for field "CreateRequestID" (CreateRequestID) [uint64]
func (a *DBTrackerLog) get_CreateRequestID(p *savepb.TrackerLog) uint64 {
	return uint64(p.CreateRequestID)
}

// getter for field "CreateType" (CreateType) [uint32]
func (a *DBTrackerLog) get_CreateType(p *savepb.TrackerLog) uint32 {
	return uint32(p.CreateType)
}

// getter for field "LogMessage" (LogMessage) [string]
func (a *DBTrackerLog) get_LogMessage(p *savepb.TrackerLog) string {
	return string(p.LogMessage)
}

// getter for field "PublicMessage" (PublicMessage) [string]
func (a *DBTrackerLog) get_PublicMessage(p *savepb.TrackerLog) string {
	return string(p.PublicMessage)
}

// getter for field "Occured" (Occured) [uint32]
func (a *DBTrackerLog) get_Occured(p *savepb.TrackerLog) uint32 {
	return uint32(p.Occured)
}

// getter for field "Success" (Success) [bool]
func (a *DBTrackerLog) get_Success(p *savepb.TrackerLog) bool {
	return bool(p.Success)
}

// getter for field "Task" (Task) [string]
func (a *DBTrackerLog) get_Task(p *savepb.TrackerLog) string {
	return string(p.Task)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBTrackerLog) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.TrackerLog, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		efname := fmt.Sprintf("EXTRA_FIELD_%d", i)
		query.Add(col_name+" = "+efname, QP{efname: value})
	}

	gw, paras := query.ToPostgres()
	queryname := "custom_dbquery"
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where "+gw, paras...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (a *DBTrackerLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.TrackerLog, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBTrackerLog) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.TrackerLog, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	eq := ""
	if extra_fields != nil && len(extra_fields) > 0 {
		eq = " AND ("
		// build the extraquery "eq"
		i := len(args)
		deli := ""
		for col_name, value := range extra_fields {
			i++
			eq = eq + deli + col_name + fmt.Sprintf(" = $%d", i)
			deli = " AND "
			args = append(args, value)
		}
		eq = eq + ")"
	}
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where ( "+query_where+") "+eq, args...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBTrackerLog) get_col_from_proto(p *savepb.TrackerLog, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "createrequestid" {
		return a.get_CreateRequestID(p)
	} else if colname == "createtype" {
		return a.get_CreateType(p)
	} else if colname == "logmessage" {
		return a.get_LogMessage(p)
	} else if colname == "publicmessage" {
		return a.get_PublicMessage(p)
	} else if colname == "occured" {
		return a.get_Occured(p)
	} else if colname == "success" {
		return a.get_Success(p)
	} else if colname == "task" {
		return a.get_Task(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBTrackerLog) Tablename() string {
	return a.SQLTablename
}

func (a *DBTrackerLog) SelectCols() string {
	return "id,createrequestid, createtype, logmessage, publicmessage, occured, success, task"
}
func (a *DBTrackerLog) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".createrequestid, " + a.SQLTablename + ".createtype, " + a.SQLTablename + ".logmessage, " + a.SQLTablename + ".publicmessage, " + a.SQLTablename + ".occured, " + a.SQLTablename + ".success, " + a.SQLTablename + ".task"
}

func (a *DBTrackerLog) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.TrackerLog, error) {
	var res []*savepb.TrackerLog
	for rows.Next() {
		// SCANNER:
		foo := &savepb.TrackerLog{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.CreateRequestID
		scanTarget_2 := &foo.CreateType
		scanTarget_3 := &foo.LogMessage
		scanTarget_4 := &foo.PublicMessage
		scanTarget_5 := &foo.Occured
		scanTarget_6 := &foo.Success
		scanTarget_7 := &foo.Task
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7)
		// END SCANNER

		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBTrackerLog) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null ,createtype integer not null ,logmessage text not null ,publicmessage text not null ,occured integer not null ,success boolean not null ,task text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null ,createtype integer not null ,logmessage text not null ,publicmessage text not null ,occured integer not null ,success boolean not null ,task text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createtype integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS logmessage text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS publicmessage text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS occured integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS success boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS task text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createrequestid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createtype integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS logmessage text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS publicmessage text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS occured integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS success boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS task text not null  default '';`,
	}

	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}

	// these are optional, expected to fail
	csql = []string{
		// Indices:

		// Foreign keys:

	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBTrackerLog) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

