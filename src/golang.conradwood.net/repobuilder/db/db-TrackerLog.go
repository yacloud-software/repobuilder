package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBTrackerLog
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBTrackerLog *DBTrackerLog
)

type DBTrackerLog struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBTrackerLog) Save(ctx context.Context, p *savepb.TrackerLog) (uint64, error) {
	qn := "DBTrackerLog_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (createrequestid, createtype, logmessage, publicmessage, occured, success, task) values ($1, $2, $3, $4, $5, $6, $7) returning id", p.CreateRequestID, p.CreateType, p.LogMessage, p.PublicMessage, p.Occured, p.Success, p.Task)
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
func (a *DBTrackerLog) SaveWithID(ctx context.Context, p *savepb.TrackerLog) error {
	qn := "insert_DBTrackerLog"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,createrequestid, createtype, logmessage, publicmessage, occured, success, task) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.CreateRequestID, p.CreateType, p.LogMessage, p.PublicMessage, p.Occured, p.Success, p.Task)
	return a.Error(ctx, qn, e)
}

func (a *DBTrackerLog) Update(ctx context.Context, p *savepb.TrackerLog) error {
	qn := "DBTrackerLog_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set createrequestid=$1, createtype=$2, logmessage=$3, publicmessage=$4, occured=$5, success=$6, task=$7 where id = $8", p.CreateRequestID, p.CreateType, p.LogMessage, p.PublicMessage, p.Occured, p.Success, p.Task, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No TrackerLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) TrackerLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBTrackerLog) TryByID(ctx context.Context, p uint64) (*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) TrackerLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBTrackerLog) All(ctx context.Context) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" order by id")
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

// get all "DBTrackerLog" rows with matching CreateRequestID
func (a *DBTrackerLog) ByCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByCreateRequestID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where createrequestid = $1", p)
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
func (a *DBTrackerLog) ByLikeCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeCreateRequestID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where createrequestid ilike $1", p)
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

// get all "DBTrackerLog" rows with matching CreateType
func (a *DBTrackerLog) ByCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByCreateType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where createtype = $1", p)
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
func (a *DBTrackerLog) ByLikeCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeCreateType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where createtype ilike $1", p)
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

// get all "DBTrackerLog" rows with matching LogMessage
func (a *DBTrackerLog) ByLogMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLogMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where logmessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeLogMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeLogMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where logmessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching PublicMessage
func (a *DBTrackerLog) ByPublicMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByPublicMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where publicmessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublicMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublicMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikePublicMessage(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikePublicMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where publicmessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublicMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPublicMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching Occured
func (a *DBTrackerLog) ByOccured(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByOccured"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where occured = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOccured: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOccured: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeOccured(ctx context.Context, p uint32) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeOccured"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where occured ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOccured: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOccured: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerLog" rows with matching Success
func (a *DBTrackerLog) BySuccess(ctx context.Context, p bool) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_BySuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where success = $1", p)
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
func (a *DBTrackerLog) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeSuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where success ilike $1", p)
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

// get all "DBTrackerLog" rows with matching Task
func (a *DBTrackerLog) ByTask(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByTask"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where task = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTask: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTask: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerLog) ByLikeTask(ctx context.Context, p string) ([]*savepb.TrackerLog, error) {
	qn := "DBTrackerLog_ByLikeTask"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, logmessage, publicmessage, occured, success, task from "+a.SQLTablename+" where task ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTask: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTask: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBTrackerLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.TrackerLog, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
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
		foo := savepb.TrackerLog{}
		err := rows.Scan(&foo.ID, &foo.CreateRequestID, &foo.CreateType, &foo.LogMessage, &foo.PublicMessage, &foo.Occured, &foo.Success, &foo.Task)
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
func (a *DBTrackerLog) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null  ,createtype integer not null  ,logmessage text not null  ,publicmessage text not null  ,occured integer not null  ,success boolean not null  ,task text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null  ,createtype integer not null  ,logmessage text not null  ,publicmessage text not null  ,occured integer not null  ,success boolean not null  ,task text not null  );`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS createtype integer not null default 0;`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS logmessage text not null default '';`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS publicmessage text not null default '';`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS occured integer not null default 0;`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS success boolean not null default false;`,
		`ALTER TABLE trackerlog ADD COLUMN IF NOT EXISTS task text not null default '';`,
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
func (a *DBTrackerLog) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

