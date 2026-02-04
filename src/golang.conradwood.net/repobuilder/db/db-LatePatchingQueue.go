package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBLatePatchingQueue
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence latepatchingqueue_seq;

Main Table:

 CREATE TABLE latepatchingqueue (id integer primary key default nextval('latepatchingqueue_seq'),repositoryid bigint not null  unique  ,entrycreated integer not null  ,lastattempt integer not null  );

Alter statements:
ALTER TABLE latepatchingqueue ADD COLUMN IF NOT EXISTS repositoryid bigint not null unique  default 0;
ALTER TABLE latepatchingqueue ADD COLUMN IF NOT EXISTS entrycreated integer not null default 0;
ALTER TABLE latepatchingqueue ADD COLUMN IF NOT EXISTS lastattempt integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE latepatchingqueue_archive (id integer unique not null,repositoryid bigint not null,entrycreated integer not null,lastattempt integer not null);
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
	default_def_DBLatePatchingQueue *DBLatePatchingQueue
)

type DBLatePatchingQueue struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBLatePatchingQueue()
	})
}

func DefaultDBLatePatchingQueue() *DBLatePatchingQueue {
	if default_def_DBLatePatchingQueue != nil {
		return default_def_DBLatePatchingQueue
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBLatePatchingQueue(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBLatePatchingQueue = res
	return res
}
func NewDBLatePatchingQueue(db *sql.DB) *DBLatePatchingQueue {
	foo := DBLatePatchingQueue{DB: db}
	foo.SQLTablename = "latepatchingqueue"
	foo.SQLArchivetablename = "latepatchingqueue_archive"
	return &foo
}

func (a *DBLatePatchingQueue) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBLatePatchingQueue) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBLatePatchingQueue) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBLatePatchingQueue) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBLatePatchingQueue", "insert into "+a.SQLArchivetablename+" (id,repositoryid, entrycreated, lastattempt) values ($1,$2, $3, $4) ", p.ID, p.RepositoryID, p.EntryCreated, p.LastAttempt)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBLatePatchingQueue) buildSaveMap(ctx context.Context, p *savepb.LatePatchingQueue) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["entrycreated"] = a.get_col_from_proto(p, "entrycreated")
	res["lastattempt"] = a.get_col_from_proto(p, "lastattempt")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBLatePatchingQueue) Save(ctx context.Context, p *savepb.LatePatchingQueue) (uint64, error) {
	qn := "save_DBLatePatchingQueue"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBLatePatchingQueue) SaveWithID(ctx context.Context, p *savepb.LatePatchingQueue) error {
	qn := "insert_DBLatePatchingQueue"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBLatePatchingQueue) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.LatePatchingQueue) (uint64, error) {
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

// if ID==0 save, otherwise update
func (a *DBLatePatchingQueue) SaveOrUpdate(ctx context.Context, p *savepb.LatePatchingQueue) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBLatePatchingQueue) Update(ctx context.Context, p *savepb.LatePatchingQueue) error {
	qn := "DBLatePatchingQueue_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, entrycreated=$2, lastattempt=$3 where id = $4", a.get_RepositoryID(p), a.get_EntryCreated(p), a.get_LastAttempt(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBLatePatchingQueue) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBLatePatchingQueue_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBLatePatchingQueue) ByID(ctx context.Context, p uint64) (*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No LatePatchingQueue with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) LatePatchingQueue with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBLatePatchingQueue) TryByID(ctx context.Context, p uint64) (*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) LatePatchingQueue with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBLatePatchingQueue) ByIDs(ctx context.Context, p []uint64) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBLatePatchingQueue) All(ctx context.Context) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBLatePatchingQueue" rows with matching RepositoryID
func (a *DBLatePatchingQueue) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with multiple matching RepositoryID
func (a *DBLatePatchingQueue) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLatePatchingQueue) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with matching EntryCreated
func (a *DBLatePatchingQueue) ByEntryCreated(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByEntryCreated"
	l, e := a.fromQuery(ctx, qn, "entrycreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByEntryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with multiple matching EntryCreated
func (a *DBLatePatchingQueue) ByMultiEntryCreated(ctx context.Context, p []uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByEntryCreated"
	l, e := a.fromQuery(ctx, qn, "entrycreated in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByEntryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLatePatchingQueue) ByLikeEntryCreated(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLikeEntryCreated"
	l, e := a.fromQuery(ctx, qn, "entrycreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByEntryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with matching LastAttempt
func (a *DBLatePatchingQueue) ByLastAttempt(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLastAttempt"
	l, e := a.fromQuery(ctx, qn, "lastattempt = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastAttempt: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with multiple matching LastAttempt
func (a *DBLatePatchingQueue) ByMultiLastAttempt(ctx context.Context, p []uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLastAttempt"
	l, e := a.fromQuery(ctx, qn, "lastattempt in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastAttempt: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLatePatchingQueue) ByLikeLastAttempt(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLikeLastAttempt"
	l, e := a.fromQuery(ctx, qn, "lastattempt ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastAttempt: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBLatePatchingQueue) get_ID(p *savepb.LatePatchingQueue) uint64 {
	return uint64(p.ID)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBLatePatchingQueue) get_RepositoryID(p *savepb.LatePatchingQueue) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "EntryCreated" (EntryCreated) [uint32]
func (a *DBLatePatchingQueue) get_EntryCreated(p *savepb.LatePatchingQueue) uint32 {
	return uint32(p.EntryCreated)
}

// getter for field "LastAttempt" (LastAttempt) [uint32]
func (a *DBLatePatchingQueue) get_LastAttempt(p *savepb.LatePatchingQueue) uint32 {
	return uint32(p.LastAttempt)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBLatePatchingQueue) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.LatePatchingQueue, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		/*
		   efname:=fmt.Sprintf("EXTRA_FIELD_%d",i)
		   query.Add(col_name+" = "+efname,QP{efname:value})
		*/
		query.AddEqual(col_name, value)
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

func (a *DBLatePatchingQueue) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.LatePatchingQueue, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBLatePatchingQueue) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.LatePatchingQueue, error) {
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
func (a *DBLatePatchingQueue) get_col_from_proto(p *savepb.LatePatchingQueue, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "entrycreated" {
		return a.get_EntryCreated(p)
	} else if colname == "lastattempt" {
		return a.get_LastAttempt(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBLatePatchingQueue) Tablename() string {
	return a.SQLTablename
}

func (a *DBLatePatchingQueue) SelectCols() string {
	return "id,repositoryid, entrycreated, lastattempt"
}
func (a *DBLatePatchingQueue) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".entrycreated, " + a.SQLTablename + ".lastattempt"
}

func (a *DBLatePatchingQueue) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.LatePatchingQueue, error) {
	var res []*savepb.LatePatchingQueue
	for rows.Next() {
		// SCANNER:
		foo := &savepb.LatePatchingQueue{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepositoryID
		scanTarget_2 := &foo.EntryCreated
		scanTarget_3 := &foo.LastAttempt
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3)
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
func (a *DBLatePatchingQueue) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,entrycreated integer not null ,lastattempt integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,entrycreated integer not null ,lastattempt integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS entrycreated integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS lastattempt integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS entrycreated integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS lastattempt integer not null  default 0;`,
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
		`create unique index if not exists uniq_latepatchingqueue_repositoryid on latepatchingqueue (repositoryid);`,
		`alter table latepatchingqueue add constraint uniq_latepatchingqueue_repositoryid unique using index uniq_latepatchingqueue_repositoryid;`,

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
func (a *DBLatePatchingQueue) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

