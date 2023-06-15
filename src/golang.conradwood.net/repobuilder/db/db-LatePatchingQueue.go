package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBLatePatchingQueue
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBLatePatchingQueue *DBLatePatchingQueue
)

type DBLatePatchingQueue struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBLatePatchingQueue) Save(ctx context.Context, p *savepb.LatePatchingQueue) (uint64, error) {
	qn := "DBLatePatchingQueue_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (repositoryid, entrycreated, lastattempt) values ($1, $2, $3) returning id", p.RepositoryID, p.EntryCreated, p.LastAttempt)
	if e != nil {
		return 0, a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBLatePatchingQueue) SaveWithID(ctx context.Context, p *savepb.LatePatchingQueue) error {
	qn := "insert_DBLatePatchingQueue"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,repositoryid, entrycreated, lastattempt) values ($1,$2, $3, $4) ", p.ID, p.RepositoryID, p.EntryCreated, p.LastAttempt)
	return a.Error(ctx, qn, e)
}

func (a *DBLatePatchingQueue) Update(ctx context.Context, p *savepb.LatePatchingQueue) error {
	qn := "DBLatePatchingQueue_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, entrycreated=$2, lastattempt=$3 where id = $4", p.RepositoryID, p.EntryCreated, p.LastAttempt, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No LatePatchingQueue with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) LatePatchingQueue with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBLatePatchingQueue) TryByID(ctx context.Context, p uint64) (*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) LatePatchingQueue with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBLatePatchingQueue) All(ctx context.Context) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" order by id")
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("All: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, fmt.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBLatePatchingQueue" rows with matching RepositoryID
func (a *DBLatePatchingQueue) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLatePatchingQueue) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with matching EntryCreated
func (a *DBLatePatchingQueue) ByEntryCreated(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByEntryCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where entrycreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEntryCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEntryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLatePatchingQueue) ByLikeEntryCreated(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLikeEntryCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where entrycreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEntryCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEntryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLatePatchingQueue" rows with matching LastAttempt
func (a *DBLatePatchingQueue) ByLastAttempt(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLastAttempt"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where lastattempt = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastAttempt: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastAttempt: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLatePatchingQueue) ByLikeLastAttempt(ctx context.Context, p uint32) ([]*savepb.LatePatchingQueue, error) {
	qn := "DBLatePatchingQueue_ByLikeLastAttempt"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, entrycreated, lastattempt from "+a.SQLTablename+" where lastattempt ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastAttempt: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastAttempt: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBLatePatchingQueue) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.LatePatchingQueue, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
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
		foo := savepb.LatePatchingQueue{}
		err := rows.Scan(&foo.ID, &foo.RepositoryID, &foo.EntryCreated, &foo.LastAttempt)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBLatePatchingQueue) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null  unique  ,entrycreated integer not null  ,lastattempt integer not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null  unique  ,entrycreated integer not null  ,lastattempt integer not null  );`,
		`ALTER TABLE latepatchingqueue ADD COLUMN IF NOT EXISTS repositoryid bigint not null unique  default 0;`,
		`ALTER TABLE latepatchingqueue ADD COLUMN IF NOT EXISTS entrycreated integer not null default 0;`,
		`ALTER TABLE latepatchingqueue ADD COLUMN IF NOT EXISTS lastattempt integer not null default 0;`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
