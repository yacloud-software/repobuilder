package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBCreateWebRepoRequest
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence createwebreporequest_seq;

Main Table:

 CREATE TABLE createwebreporequest (id integer primary key default nextval('createwebreporequest_seq'),description text not null  ,name text not null  ,domain text not null  ,reponame text not null  ,servicename text not null  ,protodomain text not null  );

Alter statements:
ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS description text not null default '';
ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS name text not null default '';
ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS domain text not null default '';
ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS reponame text not null default '';
ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS servicename text not null default '';
ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS protodomain text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE createwebreporequest_archive (id integer unique not null,description text not null,name text not null,domain text not null,reponame text not null,servicename text not null,protodomain text not null);
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
	default_def_DBCreateWebRepoRequest *DBCreateWebRepoRequest
)

type DBCreateWebRepoRequest struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBCreateWebRepoRequest() *DBCreateWebRepoRequest {
	if default_def_DBCreateWebRepoRequest != nil {
		return default_def_DBCreateWebRepoRequest
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBCreateWebRepoRequest(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBCreateWebRepoRequest = res
	return res
}
func NewDBCreateWebRepoRequest(db *sql.DB) *DBCreateWebRepoRequest {
	foo := DBCreateWebRepoRequest{DB: db}
	foo.SQLTablename = "createwebreporequest"
	foo.SQLArchivetablename = "createwebreporequest_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBCreateWebRepoRequest) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBCreateWebRepoRequest", "insert into "+a.SQLArchivetablename+" (id,description, name, domain, reponame, servicename, protodomain) values ($1,$2, $3, $4, $5, $6, $7) ", p.ID, p.Description, p.Name, p.Domain, p.RepoName, p.ServiceName, p.ProtoDomain)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBCreateWebRepoRequest) Save(ctx context.Context, p *savepb.CreateWebRepoRequest) (uint64, error) {
	qn := "DBCreateWebRepoRequest_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (description, name, domain, reponame, servicename, protodomain) values ($1, $2, $3, $4, $5, $6) returning id", p.Description, p.Name, p.Domain, p.RepoName, p.ServiceName, p.ProtoDomain)
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
func (a *DBCreateWebRepoRequest) SaveWithID(ctx context.Context, p *savepb.CreateWebRepoRequest) error {
	qn := "insert_DBCreateWebRepoRequest"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,description, name, domain, reponame, servicename, protodomain) values ($1,$2, $3, $4, $5, $6, $7) ", p.ID, p.Description, p.Name, p.Domain, p.RepoName, p.ServiceName, p.ProtoDomain)
	return a.Error(ctx, qn, e)
}

func (a *DBCreateWebRepoRequest) Update(ctx context.Context, p *savepb.CreateWebRepoRequest) error {
	qn := "DBCreateWebRepoRequest_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set description=$1, name=$2, domain=$3, reponame=$4, servicename=$5, protodomain=$6 where id = $7", p.Description, p.Name, p.Domain, p.RepoName, p.ServiceName, p.ProtoDomain, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBCreateWebRepoRequest) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBCreateWebRepoRequest_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBCreateWebRepoRequest) ByID(ctx context.Context, p uint64) (*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No CreateWebRepoRequest with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) CreateWebRepoRequest with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBCreateWebRepoRequest) TryByID(ctx context.Context, p uint64) (*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) CreateWebRepoRequest with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBCreateWebRepoRequest) All(ctx context.Context) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" order by id")
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

// get all "DBCreateWebRepoRequest" rows with matching Description
func (a *DBCreateWebRepoRequest) ByDescription(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByDescription"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where description = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeDescription(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeDescription"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where description ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching Name
func (a *DBCreateWebRepoRequest) ByName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where name = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where name ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching Domain
func (a *DBCreateWebRepoRequest) ByDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByDomain"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where domain = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDomain: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDomain: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeDomain"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where domain ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDomain: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDomain: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching RepoName
func (a *DBCreateWebRepoRequest) ByRepoName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByRepoName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where reponame = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeRepoName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeRepoName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where reponame ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching ServiceName
func (a *DBCreateWebRepoRequest) ByServiceName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByServiceName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where servicename = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeServiceName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeServiceName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where servicename ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching ProtoDomain
func (a *DBCreateWebRepoRequest) ByProtoDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByProtoDomain"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where protodomain = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoDomain: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoDomain: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeProtoDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeProtoDomain"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,description, name, domain, reponame, servicename, protodomain from "+a.SQLTablename+" where protodomain ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoDomain: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoDomain: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBCreateWebRepoRequest) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.CreateWebRepoRequest, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBCreateWebRepoRequest) Tablename() string {
	return a.SQLTablename
}

func (a *DBCreateWebRepoRequest) SelectCols() string {
	return "id,description, name, domain, reponame, servicename, protodomain"
}
func (a *DBCreateWebRepoRequest) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".description, " + a.SQLTablename + ".name, " + a.SQLTablename + ".domain, " + a.SQLTablename + ".reponame, " + a.SQLTablename + ".servicename, " + a.SQLTablename + ".protodomain"
}

func (a *DBCreateWebRepoRequest) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.CreateWebRepoRequest, error) {
	var res []*savepb.CreateWebRepoRequest
	for rows.Next() {
		foo := savepb.CreateWebRepoRequest{}
		err := rows.Scan(&foo.ID, &foo.Description, &foo.Name, &foo.Domain, &foo.RepoName, &foo.ServiceName, &foo.ProtoDomain)
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
func (a *DBCreateWebRepoRequest) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),description text not null  ,name text not null  ,domain text not null  ,reponame text not null  ,servicename text not null  ,protodomain text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),description text not null  ,name text not null  ,domain text not null  ,reponame text not null  ,servicename text not null  ,protodomain text not null  );`,
		`ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS description text not null default '';`,
		`ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS name text not null default '';`,
		`ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS domain text not null default '';`,
		`ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS reponame text not null default '';`,
		`ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS servicename text not null default '';`,
		`ALTER TABLE createwebreporequest ADD COLUMN IF NOT EXISTS protodomain text not null default '';`,
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
func (a *DBCreateWebRepoRequest) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}



