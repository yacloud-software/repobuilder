package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBRepoCreateStatus
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence repocreatestatus_seq;

Main Table:

 CREATE TABLE repocreatestatus (id integer primary key default nextval('repocreatestatus_seq'),createrequestid bigint not null  ,createtype integer not null  ,success boolean not null  ,error text not null  );

Alter statements:
ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;
ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS createtype integer not null default 0;
ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS success boolean not null default false;
ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS error text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE repocreatestatus_archive (id integer unique not null,createrequestid bigint not null,createtype integer not null,success boolean not null,error text not null);
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
	default_def_DBRepoCreateStatus *DBRepoCreateStatus
)

type DBRepoCreateStatus struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBRepoCreateStatus() *DBRepoCreateStatus {
	if default_def_DBRepoCreateStatus != nil {
		return default_def_DBRepoCreateStatus
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBRepoCreateStatus(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBRepoCreateStatus = res
	return res
}
func NewDBRepoCreateStatus(db *sql.DB) *DBRepoCreateStatus {
	foo := DBRepoCreateStatus{DB: db}
	foo.SQLTablename = "repocreatestatus"
	foo.SQLArchivetablename = "repocreatestatus_archive"
	return &foo
}

func (a *DBRepoCreateStatus) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBRepoCreateStatus) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

// archive. It is NOT transactionally save.
func (a *DBRepoCreateStatus) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBRepoCreateStatus", "insert into "+a.SQLArchivetablename+" (id,createrequestid, createtype, success, error) values ($1,$2, $3, $4, $5) ", p.ID, p.CreateRequestID, p.CreateType, p.Success, p.Error)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBRepoCreateStatus) buildSaveMap(ctx context.Context, p *savepb.RepoCreateStatus) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["createrequestid"] = a.get_col_from_proto(p, "createrequestid")
	res["createtype"] = a.get_col_from_proto(p, "createtype")
	res["success"] = a.get_col_from_proto(p, "success")
	res["error"] = a.get_col_from_proto(p, "error")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBRepoCreateStatus) Save(ctx context.Context, p *savepb.RepoCreateStatus) (uint64, error) {
	qn := "save_DBRepoCreateStatus"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBRepoCreateStatus) SaveWithID(ctx context.Context, p *savepb.RepoCreateStatus) error {
	qn := "insert_DBRepoCreateStatus"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBRepoCreateStatus) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.RepoCreateStatus) (uint64, error) {
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

func (a *DBRepoCreateStatus) Update(ctx context.Context, p *savepb.RepoCreateStatus) error {
	qn := "DBRepoCreateStatus_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set createrequestid=$1, createtype=$2, success=$3, error=$4 where id = $5", a.get_CreateRequestID(p), a.get_CreateType(p), a.get_Success(p), a.get_Error(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBRepoCreateStatus) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBRepoCreateStatus_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBRepoCreateStatus) ByID(ctx context.Context, p uint64) (*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No RepoCreateStatus with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) RepoCreateStatus with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBRepoCreateStatus) TryByID(ctx context.Context, p uint64) (*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) RepoCreateStatus with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBRepoCreateStatus) ByIDs(ctx context.Context, p []uint64) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBRepoCreateStatus) All(ctx context.Context) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBRepoCreateStatus" rows with matching CreateRequestID
func (a *DBRepoCreateStatus) ByCreateRequestID(ctx context.Context, p uint64) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with multiple matching CreateRequestID
func (a *DBRepoCreateStatus) ByMultiCreateRequestID(ctx context.Context, p []uint64) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeCreateRequestID(ctx context.Context, p uint64) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with matching CreateType
func (a *DBRepoCreateStatus) ByCreateType(ctx context.Context, p uint32) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with multiple matching CreateType
func (a *DBRepoCreateStatus) ByMultiCreateType(ctx context.Context, p []uint32) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeCreateType(ctx context.Context, p uint32) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with matching Success
func (a *DBRepoCreateStatus) BySuccess(ctx context.Context, p bool) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with multiple matching Success
func (a *DBRepoCreateStatus) ByMultiSuccess(ctx context.Context, p []bool) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeSuccess"
	l, e := a.fromQuery(ctx, qn, "success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with matching Error
func (a *DBRepoCreateStatus) ByError(ctx context.Context, p string) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByError"
	l, e := a.fromQuery(ctx, qn, "error = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByError: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with multiple matching Error
func (a *DBRepoCreateStatus) ByMultiError(ctx context.Context, p []string) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByError"
	l, e := a.fromQuery(ctx, qn, "error in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByError: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeError(ctx context.Context, p string) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeError"
	l, e := a.fromQuery(ctx, qn, "error ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByError: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBRepoCreateStatus) get_ID(p *savepb.RepoCreateStatus) uint64 {
	return uint64(p.ID)
}

// getter for field "CreateRequestID" (CreateRequestID) [uint64]
func (a *DBRepoCreateStatus) get_CreateRequestID(p *savepb.RepoCreateStatus) uint64 {
	return uint64(p.CreateRequestID)
}

// getter for field "CreateType" (CreateType) [uint32]
func (a *DBRepoCreateStatus) get_CreateType(p *savepb.RepoCreateStatus) uint32 {
	return uint32(p.CreateType)
}

// getter for field "Success" (Success) [bool]
func (a *DBRepoCreateStatus) get_Success(p *savepb.RepoCreateStatus) bool {
	return bool(p.Success)
}

// getter for field "Error" (Error) [string]
func (a *DBRepoCreateStatus) get_Error(p *savepb.RepoCreateStatus) string {
	return string(p.Error)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRepoCreateStatus) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.RepoCreateStatus, error) {
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

func (a *DBRepoCreateStatus) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.RepoCreateStatus, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBRepoCreateStatus) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.RepoCreateStatus, error) {
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
func (a *DBRepoCreateStatus) get_col_from_proto(p *savepb.RepoCreateStatus, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "createrequestid" {
		return a.get_CreateRequestID(p)
	} else if colname == "createtype" {
		return a.get_CreateType(p)
	} else if colname == "success" {
		return a.get_Success(p)
	} else if colname == "error" {
		return a.get_Error(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBRepoCreateStatus) Tablename() string {
	return a.SQLTablename
}

func (a *DBRepoCreateStatus) SelectCols() string {
	return "id,createrequestid, createtype, success, error"
}
func (a *DBRepoCreateStatus) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".createrequestid, " + a.SQLTablename + ".createtype, " + a.SQLTablename + ".success, " + a.SQLTablename + ".error"
}

func (a *DBRepoCreateStatus) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.RepoCreateStatus, error) {
	var res []*savepb.RepoCreateStatus
	for rows.Next() {
		// SCANNER:
		foo := &savepb.RepoCreateStatus{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.CreateRequestID
		scanTarget_2 := &foo.CreateType
		scanTarget_3 := &foo.Success
		scanTarget_4 := &foo.Error
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4)
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
func (a *DBRepoCreateStatus) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null ,createtype integer not null ,success boolean not null ,error text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null ,createtype integer not null ,success boolean not null ,error text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createtype integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS success boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS error text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createrequestid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createtype integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS success boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS error text not null  default '';`,
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
func (a *DBRepoCreateStatus) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

