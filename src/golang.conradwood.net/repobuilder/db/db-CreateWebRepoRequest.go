package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBCreateWebRepoRequest
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
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBCreateWebRepoRequest *DBCreateWebRepoRequest
)

type DBCreateWebRepoRequest struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
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

func (a *DBCreateWebRepoRequest) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBCreateWebRepoRequest) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
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

// return a map with columnname -> value_from_proto
func (a *DBCreateWebRepoRequest) buildSaveMap(ctx context.Context, p *savepb.CreateWebRepoRequest) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["description"] = a.get_col_from_proto(p, "description")
	res["name"] = a.get_col_from_proto(p, "name")
	res["domain"] = a.get_col_from_proto(p, "domain")
	res["reponame"] = a.get_col_from_proto(p, "reponame")
	res["servicename"] = a.get_col_from_proto(p, "servicename")
	res["protodomain"] = a.get_col_from_proto(p, "protodomain")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBCreateWebRepoRequest) Save(ctx context.Context, p *savepb.CreateWebRepoRequest) (uint64, error) {
	qn := "save_DBCreateWebRepoRequest"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBCreateWebRepoRequest) SaveWithID(ctx context.Context, p *savepb.CreateWebRepoRequest) error {
	qn := "insert_DBCreateWebRepoRequest"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBCreateWebRepoRequest) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.CreateWebRepoRequest) (uint64, error) {
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

func (a *DBCreateWebRepoRequest) Update(ctx context.Context, p *savepb.CreateWebRepoRequest) error {
	qn := "DBCreateWebRepoRequest_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set description=$1, name=$2, domain=$3, reponame=$4, servicename=$5, protodomain=$6 where id = $7", a.get_Description(p), a.get_Name(p), a.get_Domain(p), a.get_RepoName(p), a.get_ServiceName(p), a.get_ProtoDomain(p), p.ID)

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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No CreateWebRepoRequest with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) CreateWebRepoRequest with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBCreateWebRepoRequest) TryByID(ctx context.Context, p uint64) (*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) CreateWebRepoRequest with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBCreateWebRepoRequest) ByIDs(ctx context.Context, p []uint64) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBCreateWebRepoRequest) All(ctx context.Context) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBCreateWebRepoRequest" rows with matching Description
func (a *DBCreateWebRepoRequest) ByDescription(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByDescription"
	l, e := a.fromQuery(ctx, qn, "description = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with multiple matching Description
func (a *DBCreateWebRepoRequest) ByMultiDescription(ctx context.Context, p []string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByDescription"
	l, e := a.fromQuery(ctx, qn, "description in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeDescription(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeDescription"
	l, e := a.fromQuery(ctx, qn, "description ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching Name
func (a *DBCreateWebRepoRequest) ByName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByName"
	l, e := a.fromQuery(ctx, qn, "name = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with multiple matching Name
func (a *DBCreateWebRepoRequest) ByMultiName(ctx context.Context, p []string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByName"
	l, e := a.fromQuery(ctx, qn, "name in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeName"
	l, e := a.fromQuery(ctx, qn, "name ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching Domain
func (a *DBCreateWebRepoRequest) ByDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByDomain"
	l, e := a.fromQuery(ctx, qn, "domain = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDomain: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with multiple matching Domain
func (a *DBCreateWebRepoRequest) ByMultiDomain(ctx context.Context, p []string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByDomain"
	l, e := a.fromQuery(ctx, qn, "domain in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDomain: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeDomain"
	l, e := a.fromQuery(ctx, qn, "domain ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDomain: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching RepoName
func (a *DBCreateWebRepoRequest) ByRepoName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByRepoName"
	l, e := a.fromQuery(ctx, qn, "reponame = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with multiple matching RepoName
func (a *DBCreateWebRepoRequest) ByMultiRepoName(ctx context.Context, p []string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByRepoName"
	l, e := a.fromQuery(ctx, qn, "reponame in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeRepoName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeRepoName"
	l, e := a.fromQuery(ctx, qn, "reponame ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching ServiceName
func (a *DBCreateWebRepoRequest) ByServiceName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByServiceName"
	l, e := a.fromQuery(ctx, qn, "servicename = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with multiple matching ServiceName
func (a *DBCreateWebRepoRequest) ByMultiServiceName(ctx context.Context, p []string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByServiceName"
	l, e := a.fromQuery(ctx, qn, "servicename in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeServiceName(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeServiceName"
	l, e := a.fromQuery(ctx, qn, "servicename ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with matching ProtoDomain
func (a *DBCreateWebRepoRequest) ByProtoDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByProtoDomain"
	l, e := a.fromQuery(ctx, qn, "protodomain = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoDomain: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateWebRepoRequest" rows with multiple matching ProtoDomain
func (a *DBCreateWebRepoRequest) ByMultiProtoDomain(ctx context.Context, p []string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByProtoDomain"
	l, e := a.fromQuery(ctx, qn, "protodomain in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoDomain: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateWebRepoRequest) ByLikeProtoDomain(ctx context.Context, p string) ([]*savepb.CreateWebRepoRequest, error) {
	qn := "DBCreateWebRepoRequest_ByLikeProtoDomain"
	l, e := a.fromQuery(ctx, qn, "protodomain ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoDomain: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBCreateWebRepoRequest) get_ID(p *savepb.CreateWebRepoRequest) uint64 {
	return uint64(p.ID)
}

// getter for field "Description" (Description) [string]
func (a *DBCreateWebRepoRequest) get_Description(p *savepb.CreateWebRepoRequest) string {
	return string(p.Description)
}

// getter for field "Name" (Name) [string]
func (a *DBCreateWebRepoRequest) get_Name(p *savepb.CreateWebRepoRequest) string {
	return string(p.Name)
}

// getter for field "Domain" (Domain) [string]
func (a *DBCreateWebRepoRequest) get_Domain(p *savepb.CreateWebRepoRequest) string {
	return string(p.Domain)
}

// getter for field "RepoName" (RepoName) [string]
func (a *DBCreateWebRepoRequest) get_RepoName(p *savepb.CreateWebRepoRequest) string {
	return string(p.RepoName)
}

// getter for field "ServiceName" (ServiceName) [string]
func (a *DBCreateWebRepoRequest) get_ServiceName(p *savepb.CreateWebRepoRequest) string {
	return string(p.ServiceName)
}

// getter for field "ProtoDomain" (ProtoDomain) [string]
func (a *DBCreateWebRepoRequest) get_ProtoDomain(p *savepb.CreateWebRepoRequest) string {
	return string(p.ProtoDomain)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBCreateWebRepoRequest) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.CreateWebRepoRequest, error) {
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

func (a *DBCreateWebRepoRequest) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.CreateWebRepoRequest, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBCreateWebRepoRequest) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.CreateWebRepoRequest, error) {
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
func (a *DBCreateWebRepoRequest) get_col_from_proto(p *savepb.CreateWebRepoRequest, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "description" {
		return a.get_Description(p)
	} else if colname == "name" {
		return a.get_Name(p)
	} else if colname == "domain" {
		return a.get_Domain(p)
	} else if colname == "reponame" {
		return a.get_RepoName(p)
	} else if colname == "servicename" {
		return a.get_ServiceName(p)
	} else if colname == "protodomain" {
		return a.get_ProtoDomain(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

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
		// SCANNER:
		foo := &savepb.CreateWebRepoRequest{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Description
		scanTarget_2 := &foo.Name
		scanTarget_3 := &foo.Domain
		scanTarget_4 := &foo.RepoName
		scanTarget_5 := &foo.ServiceName
		scanTarget_6 := &foo.ProtoDomain
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6)
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
func (a *DBCreateWebRepoRequest) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),description text not null ,name text not null ,domain text not null ,reponame text not null ,servicename text not null ,protodomain text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),description text not null ,name text not null ,domain text not null ,reponame text not null ,servicename text not null ,protodomain text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS description text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS name text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS domain text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS reponame text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS servicename text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS protodomain text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS description text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS name text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS domain text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS reponame text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS servicename text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS protodomain text not null  default '';`,
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
func (a *DBCreateWebRepoRequest) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

