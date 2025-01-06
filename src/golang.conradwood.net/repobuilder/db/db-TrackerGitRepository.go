package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBTrackerGitRepository
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence trackergitrepository_seq;

Main Table:

 CREATE TABLE trackergitrepository (id integer primary key default nextval('trackergitrepository_seq'),createrequestid bigint not null  ,createtype integer not null  ,repositoryid bigint not null  ,urlhost text not null  ,urlpath text not null  ,repositorycreated boolean not null  ,sourceinstalled boolean not null  ,packageid text not null  ,packagename text not null  ,protofilename text not null  ,protosubmitted boolean not null  ,protocommitted boolean not null  ,minprotoversion bigint not null  ,userid text not null  ,permissionscreated boolean not null  ,secureargscreated boolean not null  ,serviceid text not null  ,serviceuserid text not null  ,servicetoken text not null  ,finalised boolean not null  ,patchrepo boolean not null  ,sourcerepositoryid bigint not null  ,notificationsent boolean not null  ,artefactid bigint not null  );

Alter statements:
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS createtype integer not null default 0;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS urlhost text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS urlpath text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS repositorycreated boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS sourceinstalled boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS packageid text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS packagename text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS protofilename text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS protosubmitted boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS protocommitted boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS minprotoversion bigint not null default 0;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS permissionscreated boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS secureargscreated boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS serviceid text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS serviceuserid text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS servicetoken text not null default '';
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS finalised boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS patchrepo boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS sourcerepositoryid bigint not null default 0;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS notificationsent boolean not null default false;
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS artefactid bigint not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE trackergitrepository_archive (id integer unique not null,createrequestid bigint not null,createtype integer not null,repositoryid bigint not null,urlhost text not null,urlpath text not null,repositorycreated boolean not null,sourceinstalled boolean not null,packageid text not null,packagename text not null,protofilename text not null,protosubmitted boolean not null,protocommitted boolean not null,minprotoversion bigint not null,userid text not null,permissionscreated boolean not null,secureargscreated boolean not null,serviceid text not null,serviceuserid text not null,servicetoken text not null,finalised boolean not null,patchrepo boolean not null,sourcerepositoryid bigint not null,notificationsent boolean not null,artefactid bigint not null);
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
	default_def_DBTrackerGitRepository *DBTrackerGitRepository
)

type DBTrackerGitRepository struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBTrackerGitRepository() *DBTrackerGitRepository {
	if default_def_DBTrackerGitRepository != nil {
		return default_def_DBTrackerGitRepository
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBTrackerGitRepository(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBTrackerGitRepository = res
	return res
}
func NewDBTrackerGitRepository(db *sql.DB) *DBTrackerGitRepository {
	foo := DBTrackerGitRepository{DB: db}
	foo.SQLTablename = "trackergitrepository"
	foo.SQLArchivetablename = "trackergitrepository_archive"
	return &foo
}

func (a *DBTrackerGitRepository) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBTrackerGitRepository) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

// archive. It is NOT transactionally save.
func (a *DBTrackerGitRepository) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBTrackerGitRepository", "insert into "+a.SQLArchivetablename+" (id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent, artefactid) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25) ", p.ID, p.CreateRequestID, p.CreateType, p.RepositoryID, p.URLHost, p.URLPath, p.RepositoryCreated, p.SourceInstalled, p.PackageID, p.PackageName, p.ProtoFilename, p.ProtoSubmitted, p.ProtoCommitted, p.MinProtoVersion, p.UserID, p.PermissionsCreated, p.SecureArgsCreated, p.ServiceID, p.ServiceUserID, p.ServiceToken, p.Finalised, p.PatchRepo, p.SourceRepositoryID, p.NotificationSent, p.ArtefactID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBTrackerGitRepository) buildSaveMap(ctx context.Context, p *savepb.TrackerGitRepository) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["createrequestid"] = a.get_col_from_proto(p, "createrequestid")
	res["createtype"] = a.get_col_from_proto(p, "createtype")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["urlhost"] = a.get_col_from_proto(p, "urlhost")
	res["urlpath"] = a.get_col_from_proto(p, "urlpath")
	res["repositorycreated"] = a.get_col_from_proto(p, "repositorycreated")
	res["sourceinstalled"] = a.get_col_from_proto(p, "sourceinstalled")
	res["packageid"] = a.get_col_from_proto(p, "packageid")
	res["packagename"] = a.get_col_from_proto(p, "packagename")
	res["protofilename"] = a.get_col_from_proto(p, "protofilename")
	res["protosubmitted"] = a.get_col_from_proto(p, "protosubmitted")
	res["protocommitted"] = a.get_col_from_proto(p, "protocommitted")
	res["minprotoversion"] = a.get_col_from_proto(p, "minprotoversion")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["permissionscreated"] = a.get_col_from_proto(p, "permissionscreated")
	res["secureargscreated"] = a.get_col_from_proto(p, "secureargscreated")
	res["serviceid"] = a.get_col_from_proto(p, "serviceid")
	res["serviceuserid"] = a.get_col_from_proto(p, "serviceuserid")
	res["servicetoken"] = a.get_col_from_proto(p, "servicetoken")
	res["finalised"] = a.get_col_from_proto(p, "finalised")
	res["patchrepo"] = a.get_col_from_proto(p, "patchrepo")
	res["sourcerepositoryid"] = a.get_col_from_proto(p, "sourcerepositoryid")
	res["notificationsent"] = a.get_col_from_proto(p, "notificationsent")
	res["artefactid"] = a.get_col_from_proto(p, "artefactid")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBTrackerGitRepository) Save(ctx context.Context, p *savepb.TrackerGitRepository) (uint64, error) {
	qn := "save_DBTrackerGitRepository"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBTrackerGitRepository) SaveWithID(ctx context.Context, p *savepb.TrackerGitRepository) error {
	qn := "insert_DBTrackerGitRepository"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBTrackerGitRepository) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.TrackerGitRepository) (uint64, error) {
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

func (a *DBTrackerGitRepository) Update(ctx context.Context, p *savepb.TrackerGitRepository) error {
	qn := "DBTrackerGitRepository_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set createrequestid=$1, createtype=$2, repositoryid=$3, urlhost=$4, urlpath=$5, repositorycreated=$6, sourceinstalled=$7, packageid=$8, packagename=$9, protofilename=$10, protosubmitted=$11, protocommitted=$12, minprotoversion=$13, userid=$14, permissionscreated=$15, secureargscreated=$16, serviceid=$17, serviceuserid=$18, servicetoken=$19, finalised=$20, patchrepo=$21, sourcerepositoryid=$22, notificationsent=$23, artefactid=$24 where id = $25", a.get_CreateRequestID(p), a.get_CreateType(p), a.get_RepositoryID(p), a.get_URLHost(p), a.get_URLPath(p), a.get_RepositoryCreated(p), a.get_SourceInstalled(p), a.get_PackageID(p), a.get_PackageName(p), a.get_ProtoFilename(p), a.get_ProtoSubmitted(p), a.get_ProtoCommitted(p), a.get_MinProtoVersion(p), a.get_UserID(p), a.get_PermissionsCreated(p), a.get_SecureArgsCreated(p), a.get_ServiceID(p), a.get_ServiceUserID(p), a.get_ServiceToken(p), a.get_Finalised(p), a.get_PatchRepo(p), a.get_SourceRepositoryID(p), a.get_NotificationSent(p), a.get_ArtefactID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBTrackerGitRepository) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBTrackerGitRepository_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBTrackerGitRepository) ByID(ctx context.Context, p uint64) (*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No TrackerGitRepository with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) TrackerGitRepository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBTrackerGitRepository) TryByID(ctx context.Context, p uint64) (*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) TrackerGitRepository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBTrackerGitRepository) ByIDs(ctx context.Context, p []uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBTrackerGitRepository) All(ctx context.Context) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBTrackerGitRepository" rows with matching CreateRequestID
func (a *DBTrackerGitRepository) ByCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching CreateRequestID
func (a *DBTrackerGitRepository) ByMultiCreateRequestID(ctx context.Context, p []uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeCreateRequestID"
	l, e := a.fromQuery(ctx, qn, "createrequestid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateRequestID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching CreateType
func (a *DBTrackerGitRepository) ByCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching CreateType
func (a *DBTrackerGitRepository) ByMultiCreateType(ctx context.Context, p []uint32) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeCreateType"
	l, e := a.fromQuery(ctx, qn, "createtype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateType: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching RepositoryID
func (a *DBTrackerGitRepository) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching RepositoryID
func (a *DBTrackerGitRepository) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching URLHost
func (a *DBTrackerGitRepository) ByURLHost(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByURLHost"
	l, e := a.fromQuery(ctx, qn, "urlhost = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURLHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching URLHost
func (a *DBTrackerGitRepository) ByMultiURLHost(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByURLHost"
	l, e := a.fromQuery(ctx, qn, "urlhost in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURLHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeURLHost(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeURLHost"
	l, e := a.fromQuery(ctx, qn, "urlhost ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURLHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching URLPath
func (a *DBTrackerGitRepository) ByURLPath(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByURLPath"
	l, e := a.fromQuery(ctx, qn, "urlpath = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURLPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching URLPath
func (a *DBTrackerGitRepository) ByMultiURLPath(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByURLPath"
	l, e := a.fromQuery(ctx, qn, "urlpath in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURLPath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeURLPath(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeURLPath"
	l, e := a.fromQuery(ctx, qn, "urlpath ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByURLPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching RepositoryCreated
func (a *DBTrackerGitRepository) ByRepositoryCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByRepositoryCreated"
	l, e := a.fromQuery(ctx, qn, "repositorycreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching RepositoryCreated
func (a *DBTrackerGitRepository) ByMultiRepositoryCreated(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByRepositoryCreated"
	l, e := a.fromQuery(ctx, qn, "repositorycreated in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeRepositoryCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeRepositoryCreated"
	l, e := a.fromQuery(ctx, qn, "repositorycreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching SourceInstalled
func (a *DBTrackerGitRepository) BySourceInstalled(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySourceInstalled"
	l, e := a.fromQuery(ctx, qn, "sourceinstalled = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceInstalled: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching SourceInstalled
func (a *DBTrackerGitRepository) ByMultiSourceInstalled(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySourceInstalled"
	l, e := a.fromQuery(ctx, qn, "sourceinstalled in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceInstalled: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeSourceInstalled(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeSourceInstalled"
	l, e := a.fromQuery(ctx, qn, "sourceinstalled ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceInstalled: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PackageID
func (a *DBTrackerGitRepository) ByPackageID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPackageID"
	l, e := a.fromQuery(ctx, qn, "packageid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPackageID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching PackageID
func (a *DBTrackerGitRepository) ByMultiPackageID(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPackageID"
	l, e := a.fromQuery(ctx, qn, "packageid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPackageID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePackageID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePackageID"
	l, e := a.fromQuery(ctx, qn, "packageid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPackageID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PackageName
func (a *DBTrackerGitRepository) ByPackageName(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPackageName"
	l, e := a.fromQuery(ctx, qn, "packagename = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPackageName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching PackageName
func (a *DBTrackerGitRepository) ByMultiPackageName(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPackageName"
	l, e := a.fromQuery(ctx, qn, "packagename in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPackageName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePackageName(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePackageName"
	l, e := a.fromQuery(ctx, qn, "packagename ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPackageName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ProtoFilename
func (a *DBTrackerGitRepository) ByProtoFilename(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoFilename"
	l, e := a.fromQuery(ctx, qn, "protofilename = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoFilename: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ProtoFilename
func (a *DBTrackerGitRepository) ByMultiProtoFilename(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoFilename"
	l, e := a.fromQuery(ctx, qn, "protofilename in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoFilename: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeProtoFilename(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeProtoFilename"
	l, e := a.fromQuery(ctx, qn, "protofilename ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoFilename: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ProtoSubmitted
func (a *DBTrackerGitRepository) ByProtoSubmitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoSubmitted"
	l, e := a.fromQuery(ctx, qn, "protosubmitted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoSubmitted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ProtoSubmitted
func (a *DBTrackerGitRepository) ByMultiProtoSubmitted(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoSubmitted"
	l, e := a.fromQuery(ctx, qn, "protosubmitted in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoSubmitted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeProtoSubmitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeProtoSubmitted"
	l, e := a.fromQuery(ctx, qn, "protosubmitted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoSubmitted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ProtoCommitted
func (a *DBTrackerGitRepository) ByProtoCommitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoCommitted"
	l, e := a.fromQuery(ctx, qn, "protocommitted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoCommitted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ProtoCommitted
func (a *DBTrackerGitRepository) ByMultiProtoCommitted(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoCommitted"
	l, e := a.fromQuery(ctx, qn, "protocommitted in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoCommitted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeProtoCommitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeProtoCommitted"
	l, e := a.fromQuery(ctx, qn, "protocommitted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByProtoCommitted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching MinProtoVersion
func (a *DBTrackerGitRepository) ByMinProtoVersion(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByMinProtoVersion"
	l, e := a.fromQuery(ctx, qn, "minprotoversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMinProtoVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching MinProtoVersion
func (a *DBTrackerGitRepository) ByMultiMinProtoVersion(ctx context.Context, p []uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByMinProtoVersion"
	l, e := a.fromQuery(ctx, qn, "minprotoversion in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMinProtoVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeMinProtoVersion(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeMinProtoVersion"
	l, e := a.fromQuery(ctx, qn, "minprotoversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMinProtoVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching UserID
func (a *DBTrackerGitRepository) ByUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching UserID
func (a *DBTrackerGitRepository) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PermissionsCreated
func (a *DBTrackerGitRepository) ByPermissionsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPermissionsCreated"
	l, e := a.fromQuery(ctx, qn, "permissionscreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPermissionsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching PermissionsCreated
func (a *DBTrackerGitRepository) ByMultiPermissionsCreated(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPermissionsCreated"
	l, e := a.fromQuery(ctx, qn, "permissionscreated in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPermissionsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePermissionsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePermissionsCreated"
	l, e := a.fromQuery(ctx, qn, "permissionscreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPermissionsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching SecureArgsCreated
func (a *DBTrackerGitRepository) BySecureArgsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySecureArgsCreated"
	l, e := a.fromQuery(ctx, qn, "secureargscreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySecureArgsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching SecureArgsCreated
func (a *DBTrackerGitRepository) ByMultiSecureArgsCreated(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySecureArgsCreated"
	l, e := a.fromQuery(ctx, qn, "secureargscreated in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySecureArgsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeSecureArgsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeSecureArgsCreated"
	l, e := a.fromQuery(ctx, qn, "secureargscreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySecureArgsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ServiceID
func (a *DBTrackerGitRepository) ByServiceID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceID"
	l, e := a.fromQuery(ctx, qn, "serviceid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ServiceID
func (a *DBTrackerGitRepository) ByMultiServiceID(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceID"
	l, e := a.fromQuery(ctx, qn, "serviceid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeServiceID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeServiceID"
	l, e := a.fromQuery(ctx, qn, "serviceid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ServiceUserID
func (a *DBTrackerGitRepository) ByServiceUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceUserID"
	l, e := a.fromQuery(ctx, qn, "serviceuserid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ServiceUserID
func (a *DBTrackerGitRepository) ByMultiServiceUserID(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceUserID"
	l, e := a.fromQuery(ctx, qn, "serviceuserid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeServiceUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeServiceUserID"
	l, e := a.fromQuery(ctx, qn, "serviceuserid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ServiceToken
func (a *DBTrackerGitRepository) ByServiceToken(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceToken"
	l, e := a.fromQuery(ctx, qn, "servicetoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ServiceToken
func (a *DBTrackerGitRepository) ByMultiServiceToken(ctx context.Context, p []string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceToken"
	l, e := a.fromQuery(ctx, qn, "servicetoken in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeServiceToken(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeServiceToken"
	l, e := a.fromQuery(ctx, qn, "servicetoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByServiceToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching Finalised
func (a *DBTrackerGitRepository) ByFinalised(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByFinalised"
	l, e := a.fromQuery(ctx, qn, "finalised = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinalised: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching Finalised
func (a *DBTrackerGitRepository) ByMultiFinalised(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByFinalised"
	l, e := a.fromQuery(ctx, qn, "finalised in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinalised: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeFinalised(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeFinalised"
	l, e := a.fromQuery(ctx, qn, "finalised ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinalised: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PatchRepo
func (a *DBTrackerGitRepository) ByPatchRepo(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPatchRepo"
	l, e := a.fromQuery(ctx, qn, "patchrepo = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPatchRepo: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching PatchRepo
func (a *DBTrackerGitRepository) ByMultiPatchRepo(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPatchRepo"
	l, e := a.fromQuery(ctx, qn, "patchrepo in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPatchRepo: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePatchRepo(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePatchRepo"
	l, e := a.fromQuery(ctx, qn, "patchrepo ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPatchRepo: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching SourceRepositoryID
func (a *DBTrackerGitRepository) BySourceRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySourceRepositoryID"
	l, e := a.fromQuery(ctx, qn, "sourcerepositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching SourceRepositoryID
func (a *DBTrackerGitRepository) ByMultiSourceRepositoryID(ctx context.Context, p []uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySourceRepositoryID"
	l, e := a.fromQuery(ctx, qn, "sourcerepositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeSourceRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeSourceRepositoryID"
	l, e := a.fromQuery(ctx, qn, "sourcerepositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching NotificationSent
func (a *DBTrackerGitRepository) ByNotificationSent(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByNotificationSent"
	l, e := a.fromQuery(ctx, qn, "notificationsent = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNotificationSent: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching NotificationSent
func (a *DBTrackerGitRepository) ByMultiNotificationSent(ctx context.Context, p []bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByNotificationSent"
	l, e := a.fromQuery(ctx, qn, "notificationsent in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNotificationSent: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeNotificationSent(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeNotificationSent"
	l, e := a.fromQuery(ctx, qn, "notificationsent ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNotificationSent: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ArtefactID
func (a *DBTrackerGitRepository) ByArtefactID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByArtefactID"
	l, e := a.fromQuery(ctx, qn, "artefactid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with multiple matching ArtefactID
func (a *DBTrackerGitRepository) ByMultiArtefactID(ctx context.Context, p []uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByArtefactID"
	l, e := a.fromQuery(ctx, qn, "artefactid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeArtefactID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeArtefactID"
	l, e := a.fromQuery(ctx, qn, "artefactid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactID: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBTrackerGitRepository) get_ID(p *savepb.TrackerGitRepository) uint64 {
	return uint64(p.ID)
}

// getter for field "CreateRequestID" (CreateRequestID) [uint64]
func (a *DBTrackerGitRepository) get_CreateRequestID(p *savepb.TrackerGitRepository) uint64 {
	return uint64(p.CreateRequestID)
}

// getter for field "CreateType" (CreateType) [uint32]
func (a *DBTrackerGitRepository) get_CreateType(p *savepb.TrackerGitRepository) uint32 {
	return uint32(p.CreateType)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBTrackerGitRepository) get_RepositoryID(p *savepb.TrackerGitRepository) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "URLHost" (URLHost) [string]
func (a *DBTrackerGitRepository) get_URLHost(p *savepb.TrackerGitRepository) string {
	return string(p.URLHost)
}

// getter for field "URLPath" (URLPath) [string]
func (a *DBTrackerGitRepository) get_URLPath(p *savepb.TrackerGitRepository) string {
	return string(p.URLPath)
}

// getter for field "RepositoryCreated" (RepositoryCreated) [bool]
func (a *DBTrackerGitRepository) get_RepositoryCreated(p *savepb.TrackerGitRepository) bool {
	return bool(p.RepositoryCreated)
}

// getter for field "SourceInstalled" (SourceInstalled) [bool]
func (a *DBTrackerGitRepository) get_SourceInstalled(p *savepb.TrackerGitRepository) bool {
	return bool(p.SourceInstalled)
}

// getter for field "PackageID" (PackageID) [string]
func (a *DBTrackerGitRepository) get_PackageID(p *savepb.TrackerGitRepository) string {
	return string(p.PackageID)
}

// getter for field "PackageName" (PackageName) [string]
func (a *DBTrackerGitRepository) get_PackageName(p *savepb.TrackerGitRepository) string {
	return string(p.PackageName)
}

// getter for field "ProtoFilename" (ProtoFilename) [string]
func (a *DBTrackerGitRepository) get_ProtoFilename(p *savepb.TrackerGitRepository) string {
	return string(p.ProtoFilename)
}

// getter for field "ProtoSubmitted" (ProtoSubmitted) [bool]
func (a *DBTrackerGitRepository) get_ProtoSubmitted(p *savepb.TrackerGitRepository) bool {
	return bool(p.ProtoSubmitted)
}

// getter for field "ProtoCommitted" (ProtoCommitted) [bool]
func (a *DBTrackerGitRepository) get_ProtoCommitted(p *savepb.TrackerGitRepository) bool {
	return bool(p.ProtoCommitted)
}

// getter for field "MinProtoVersion" (MinProtoVersion) [uint64]
func (a *DBTrackerGitRepository) get_MinProtoVersion(p *savepb.TrackerGitRepository) uint64 {
	return uint64(p.MinProtoVersion)
}

// getter for field "UserID" (UserID) [string]
func (a *DBTrackerGitRepository) get_UserID(p *savepb.TrackerGitRepository) string {
	return string(p.UserID)
}

// getter for field "PermissionsCreated" (PermissionsCreated) [bool]
func (a *DBTrackerGitRepository) get_PermissionsCreated(p *savepb.TrackerGitRepository) bool {
	return bool(p.PermissionsCreated)
}

// getter for field "SecureArgsCreated" (SecureArgsCreated) [bool]
func (a *DBTrackerGitRepository) get_SecureArgsCreated(p *savepb.TrackerGitRepository) bool {
	return bool(p.SecureArgsCreated)
}

// getter for field "ServiceID" (ServiceID) [string]
func (a *DBTrackerGitRepository) get_ServiceID(p *savepb.TrackerGitRepository) string {
	return string(p.ServiceID)
}

// getter for field "ServiceUserID" (ServiceUserID) [string]
func (a *DBTrackerGitRepository) get_ServiceUserID(p *savepb.TrackerGitRepository) string {
	return string(p.ServiceUserID)
}

// getter for field "ServiceToken" (ServiceToken) [string]
func (a *DBTrackerGitRepository) get_ServiceToken(p *savepb.TrackerGitRepository) string {
	return string(p.ServiceToken)
}

// getter for field "Finalised" (Finalised) [bool]
func (a *DBTrackerGitRepository) get_Finalised(p *savepb.TrackerGitRepository) bool {
	return bool(p.Finalised)
}

// getter for field "PatchRepo" (PatchRepo) [bool]
func (a *DBTrackerGitRepository) get_PatchRepo(p *savepb.TrackerGitRepository) bool {
	return bool(p.PatchRepo)
}

// getter for field "SourceRepositoryID" (SourceRepositoryID) [uint64]
func (a *DBTrackerGitRepository) get_SourceRepositoryID(p *savepb.TrackerGitRepository) uint64 {
	return uint64(p.SourceRepositoryID)
}

// getter for field "NotificationSent" (NotificationSent) [bool]
func (a *DBTrackerGitRepository) get_NotificationSent(p *savepb.TrackerGitRepository) bool {
	return bool(p.NotificationSent)
}

// getter for field "ArtefactID" (ArtefactID) [uint64]
func (a *DBTrackerGitRepository) get_ArtefactID(p *savepb.TrackerGitRepository) uint64 {
	return uint64(p.ArtefactID)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBTrackerGitRepository) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.TrackerGitRepository, error) {
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

func (a *DBTrackerGitRepository) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.TrackerGitRepository, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBTrackerGitRepository) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.TrackerGitRepository, error) {
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
func (a *DBTrackerGitRepository) get_col_from_proto(p *savepb.TrackerGitRepository, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "createrequestid" {
		return a.get_CreateRequestID(p)
	} else if colname == "createtype" {
		return a.get_CreateType(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "urlhost" {
		return a.get_URLHost(p)
	} else if colname == "urlpath" {
		return a.get_URLPath(p)
	} else if colname == "repositorycreated" {
		return a.get_RepositoryCreated(p)
	} else if colname == "sourceinstalled" {
		return a.get_SourceInstalled(p)
	} else if colname == "packageid" {
		return a.get_PackageID(p)
	} else if colname == "packagename" {
		return a.get_PackageName(p)
	} else if colname == "protofilename" {
		return a.get_ProtoFilename(p)
	} else if colname == "protosubmitted" {
		return a.get_ProtoSubmitted(p)
	} else if colname == "protocommitted" {
		return a.get_ProtoCommitted(p)
	} else if colname == "minprotoversion" {
		return a.get_MinProtoVersion(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "permissionscreated" {
		return a.get_PermissionsCreated(p)
	} else if colname == "secureargscreated" {
		return a.get_SecureArgsCreated(p)
	} else if colname == "serviceid" {
		return a.get_ServiceID(p)
	} else if colname == "serviceuserid" {
		return a.get_ServiceUserID(p)
	} else if colname == "servicetoken" {
		return a.get_ServiceToken(p)
	} else if colname == "finalised" {
		return a.get_Finalised(p)
	} else if colname == "patchrepo" {
		return a.get_PatchRepo(p)
	} else if colname == "sourcerepositoryid" {
		return a.get_SourceRepositoryID(p)
	} else if colname == "notificationsent" {
		return a.get_NotificationSent(p)
	} else if colname == "artefactid" {
		return a.get_ArtefactID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBTrackerGitRepository) Tablename() string {
	return a.SQLTablename
}

func (a *DBTrackerGitRepository) SelectCols() string {
	return "id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent, artefactid"
}
func (a *DBTrackerGitRepository) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".createrequestid, " + a.SQLTablename + ".createtype, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".urlhost, " + a.SQLTablename + ".urlpath, " + a.SQLTablename + ".repositorycreated, " + a.SQLTablename + ".sourceinstalled, " + a.SQLTablename + ".packageid, " + a.SQLTablename + ".packagename, " + a.SQLTablename + ".protofilename, " + a.SQLTablename + ".protosubmitted, " + a.SQLTablename + ".protocommitted, " + a.SQLTablename + ".minprotoversion, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".permissionscreated, " + a.SQLTablename + ".secureargscreated, " + a.SQLTablename + ".serviceid, " + a.SQLTablename + ".serviceuserid, " + a.SQLTablename + ".servicetoken, " + a.SQLTablename + ".finalised, " + a.SQLTablename + ".patchrepo, " + a.SQLTablename + ".sourcerepositoryid, " + a.SQLTablename + ".notificationsent, " + a.SQLTablename + ".artefactid"
}

func (a *DBTrackerGitRepository) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.TrackerGitRepository, error) {
	var res []*savepb.TrackerGitRepository
	for rows.Next() {
		// SCANNER:
		foo := &savepb.TrackerGitRepository{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.CreateRequestID
		scanTarget_2 := &foo.CreateType
		scanTarget_3 := &foo.RepositoryID
		scanTarget_4 := &foo.URLHost
		scanTarget_5 := &foo.URLPath
		scanTarget_6 := &foo.RepositoryCreated
		scanTarget_7 := &foo.SourceInstalled
		scanTarget_8 := &foo.PackageID
		scanTarget_9 := &foo.PackageName
		scanTarget_10 := &foo.ProtoFilename
		scanTarget_11 := &foo.ProtoSubmitted
		scanTarget_12 := &foo.ProtoCommitted
		scanTarget_13 := &foo.MinProtoVersion
		scanTarget_14 := &foo.UserID
		scanTarget_15 := &foo.PermissionsCreated
		scanTarget_16 := &foo.SecureArgsCreated
		scanTarget_17 := &foo.ServiceID
		scanTarget_18 := &foo.ServiceUserID
		scanTarget_19 := &foo.ServiceToken
		scanTarget_20 := &foo.Finalised
		scanTarget_21 := &foo.PatchRepo
		scanTarget_22 := &foo.SourceRepositoryID
		scanTarget_23 := &foo.NotificationSent
		scanTarget_24 := &foo.ArtefactID
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8, scanTarget_9, scanTarget_10, scanTarget_11, scanTarget_12, scanTarget_13, scanTarget_14, scanTarget_15, scanTarget_16, scanTarget_17, scanTarget_18, scanTarget_19, scanTarget_20, scanTarget_21, scanTarget_22, scanTarget_23, scanTarget_24)
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
func (a *DBTrackerGitRepository) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null ,createtype integer not null ,repositoryid bigint not null ,urlhost text not null ,urlpath text not null ,repositorycreated boolean not null ,sourceinstalled boolean not null ,packageid text not null ,packagename text not null ,protofilename text not null ,protosubmitted boolean not null ,protocommitted boolean not null ,minprotoversion bigint not null ,userid text not null ,permissionscreated boolean not null ,secureargscreated boolean not null ,serviceid text not null ,serviceuserid text not null ,servicetoken text not null ,finalised boolean not null ,patchrepo boolean not null ,sourcerepositoryid bigint not null ,notificationsent boolean not null ,artefactid bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null ,createtype integer not null ,repositoryid bigint not null ,urlhost text not null ,urlpath text not null ,repositorycreated boolean not null ,sourceinstalled boolean not null ,packageid text not null ,packagename text not null ,protofilename text not null ,protosubmitted boolean not null ,protocommitted boolean not null ,minprotoversion bigint not null ,userid text not null ,permissionscreated boolean not null ,secureargscreated boolean not null ,serviceid text not null ,serviceuserid text not null ,servicetoken text not null ,finalised boolean not null ,patchrepo boolean not null ,sourcerepositoryid bigint not null ,notificationsent boolean not null ,artefactid bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createtype integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS urlhost text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS urlpath text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositorycreated boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS sourceinstalled boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS packageid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS packagename text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS protofilename text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS protosubmitted boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS protocommitted boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS minprotoversion bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS permissionscreated boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS secureargscreated boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS serviceid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS serviceuserid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS servicetoken text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS finalised boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS patchrepo boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS sourcerepositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS notificationsent boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS artefactid bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createrequestid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createtype integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS urlhost text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS urlpath text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositorycreated boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS sourceinstalled boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS packageid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS packagename text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS protofilename text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS protosubmitted boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS protocommitted boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS minprotoversion bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS permissionscreated boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS secureargscreated boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS serviceid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS serviceuserid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS servicetoken text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS finalised boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS patchrepo boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS sourcerepositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS notificationsent boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS artefactid bigint not null  default 0;`,
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
func (a *DBTrackerGitRepository) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

