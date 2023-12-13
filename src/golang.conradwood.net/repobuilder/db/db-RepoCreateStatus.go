package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBRepoCreateStatus
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBRepoCreateStatus *DBRepoCreateStatus
)

type DBRepoCreateStatus struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBRepoCreateStatus) Save(ctx context.Context, p *savepb.RepoCreateStatus) (uint64, error) {
	qn := "DBRepoCreateStatus_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (createrequestid, createtype, success, error) values ($1, $2, $3, $4) returning id", p.CreateRequestID, p.CreateType, p.Success, p.Error)
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
func (a *DBRepoCreateStatus) SaveWithID(ctx context.Context, p *savepb.RepoCreateStatus) error {
	qn := "insert_DBRepoCreateStatus"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,createrequestid, createtype, success, error) values ($1,$2, $3, $4, $5) ", p.ID, p.CreateRequestID, p.CreateType, p.Success, p.Error)
	return a.Error(ctx, qn, e)
}

func (a *DBRepoCreateStatus) Update(ctx context.Context, p *savepb.RepoCreateStatus) error {
	qn := "DBRepoCreateStatus_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set createrequestid=$1, createtype=$2, success=$3, error=$4 where id = $5", p.CreateRequestID, p.CreateType, p.Success, p.Error, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No RepoCreateStatus with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) RepoCreateStatus with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBRepoCreateStatus) TryByID(ctx context.Context, p uint64) (*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) RepoCreateStatus with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBRepoCreateStatus) All(ctx context.Context) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" order by id")
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

// get all "DBRepoCreateStatus" rows with matching CreateRequestID
func (a *DBRepoCreateStatus) ByCreateRequestID(ctx context.Context, p uint64) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByCreateRequestID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where createrequestid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateRequestID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeCreateRequestID(ctx context.Context, p uint64) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeCreateRequestID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where createrequestid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateRequestID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with matching CreateType
func (a *DBRepoCreateStatus) ByCreateType(ctx context.Context, p uint32) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByCreateType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where createtype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateType: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeCreateType(ctx context.Context, p uint32) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeCreateType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where createtype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateType: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with matching Success
func (a *DBRepoCreateStatus) BySuccess(ctx context.Context, p bool) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_BySuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeSuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepoCreateStatus" rows with matching Error
func (a *DBRepoCreateStatus) ByError(ctx context.Context, p string) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByError"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where error = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByError: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByError: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepoCreateStatus) ByLikeError(ctx context.Context, p string) ([]*savepb.RepoCreateStatus, error) {
	qn := "DBRepoCreateStatus_ByLikeError"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, success, error from "+a.SQLTablename+" where error ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByError: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByError: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRepoCreateStatus) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.RepoCreateStatus, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
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
		foo := savepb.RepoCreateStatus{}
		err := rows.Scan(&foo.ID, &foo.CreateRequestID, &foo.CreateType, &foo.Success, &foo.Error)
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
func (a *DBRepoCreateStatus) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null  ,createtype integer not null  ,success boolean not null  ,error text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null  ,createtype integer not null  ,success boolean not null  ,error text not null  );`,
		`ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;`,
		`ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS createtype integer not null default 0;`,
		`ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS success boolean not null default false;`,
		`ALTER TABLE repocreatestatus ADD COLUMN IF NOT EXISTS error text not null default '';`,
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
func (a *DBRepoCreateStatus) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}




