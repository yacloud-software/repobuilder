package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBTrackerGitRepository
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence trackergitrepository_seq;

Main Table:

 CREATE TABLE trackergitrepository (id integer primary key default nextval('trackergitrepository_seq'),createrequestid bigint not null  ,createtype integer not null  ,repositoryid bigint not null  ,urlhost text not null  ,urlpath text not null  ,repositorycreated boolean not null  ,sourceinstalled boolean not null  ,packageid text not null  ,packagename text not null  ,protofilename text not null  ,protosubmitted boolean not null  ,protocommitted boolean not null  ,minprotoversion bigint not null  ,context text not null  ,userid text not null  ,permissionscreated boolean not null  ,secureargscreated boolean not null  ,serviceid text not null  ,serviceuserid text not null  ,servicetoken text not null  ,finalised boolean not null  ,patchrepo boolean not null  ,sourcerepositoryid bigint not null  ,notificationsent boolean not null  );

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
ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS context text not null default '';
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


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE trackergitrepository_archive (id integer unique not null,createrequestid bigint not null,createtype integer not null,repositoryid bigint not null,urlhost text not null,urlpath text not null,repositorycreated boolean not null,sourceinstalled boolean not null,packageid text not null,packagename text not null,protofilename text not null,protosubmitted boolean not null,protocommitted boolean not null,minprotoversion bigint not null,context text not null,userid text not null,permissionscreated boolean not null,secureargscreated boolean not null,serviceid text not null,serviceuserid text not null,servicetoken text not null,finalised boolean not null,patchrepo boolean not null,sourcerepositoryid bigint not null,notificationsent boolean not null);
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
	default_def_DBTrackerGitRepository *DBTrackerGitRepository
)

type DBTrackerGitRepository struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// archive. It is NOT transactionally save.
func (a *DBTrackerGitRepository) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBTrackerGitRepository", "insert into "+a.SQLArchivetablename+" (id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25) ", p.ID, p.CreateRequestID, p.CreateType, p.RepositoryID, p.URLHost, p.URLPath, p.RepositoryCreated, p.SourceInstalled, p.PackageID, p.PackageName, p.ProtoFilename, p.ProtoSubmitted, p.ProtoCommitted, p.MinProtoVersion, p.Context, p.UserID, p.PermissionsCreated, p.SecureArgsCreated, p.ServiceID, p.ServiceUserID, p.ServiceToken, p.Finalised, p.PatchRepo, p.SourceRepositoryID, p.NotificationSent)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBTrackerGitRepository) Save(ctx context.Context, p *savepb.TrackerGitRepository) (uint64, error) {
	qn := "DBTrackerGitRepository_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24) returning id", p.CreateRequestID, p.CreateType, p.RepositoryID, p.URLHost, p.URLPath, p.RepositoryCreated, p.SourceInstalled, p.PackageID, p.PackageName, p.ProtoFilename, p.ProtoSubmitted, p.ProtoCommitted, p.MinProtoVersion, p.Context, p.UserID, p.PermissionsCreated, p.SecureArgsCreated, p.ServiceID, p.ServiceUserID, p.ServiceToken, p.Finalised, p.PatchRepo, p.SourceRepositoryID, p.NotificationSent)
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
func (a *DBTrackerGitRepository) SaveWithID(ctx context.Context, p *savepb.TrackerGitRepository) error {
	qn := "insert_DBTrackerGitRepository"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25) ", p.ID, p.CreateRequestID, p.CreateType, p.RepositoryID, p.URLHost, p.URLPath, p.RepositoryCreated, p.SourceInstalled, p.PackageID, p.PackageName, p.ProtoFilename, p.ProtoSubmitted, p.ProtoCommitted, p.MinProtoVersion, p.Context, p.UserID, p.PermissionsCreated, p.SecureArgsCreated, p.ServiceID, p.ServiceUserID, p.ServiceToken, p.Finalised, p.PatchRepo, p.SourceRepositoryID, p.NotificationSent)
	return a.Error(ctx, qn, e)
}

func (a *DBTrackerGitRepository) Update(ctx context.Context, p *savepb.TrackerGitRepository) error {
	qn := "DBTrackerGitRepository_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set createrequestid=$1, createtype=$2, repositoryid=$3, urlhost=$4, urlpath=$5, repositorycreated=$6, sourceinstalled=$7, packageid=$8, packagename=$9, protofilename=$10, protosubmitted=$11, protocommitted=$12, minprotoversion=$13, context=$14, userid=$15, permissionscreated=$16, secureargscreated=$17, serviceid=$18, serviceuserid=$19, servicetoken=$20, finalised=$21, patchrepo=$22, sourcerepositoryid=$23, notificationsent=$24 where id = $25", p.CreateRequestID, p.CreateType, p.RepositoryID, p.URLHost, p.URLPath, p.RepositoryCreated, p.SourceInstalled, p.PackageID, p.PackageName, p.ProtoFilename, p.ProtoSubmitted, p.ProtoCommitted, p.MinProtoVersion, p.Context, p.UserID, p.PermissionsCreated, p.SecureArgsCreated, p.ServiceID, p.ServiceUserID, p.ServiceToken, p.Finalised, p.PatchRepo, p.SourceRepositoryID, p.NotificationSent, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No TrackerGitRepository with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) TrackerGitRepository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBTrackerGitRepository) TryByID(ctx context.Context, p uint64) (*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) TrackerGitRepository with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBTrackerGitRepository) All(ctx context.Context) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" order by id")
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

// get all "DBTrackerGitRepository" rows with matching CreateRequestID
func (a *DBTrackerGitRepository) ByCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByCreateRequestID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where createrequestid = $1", p)
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
func (a *DBTrackerGitRepository) ByLikeCreateRequestID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeCreateRequestID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where createrequestid ilike $1", p)
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

// get all "DBTrackerGitRepository" rows with matching CreateType
func (a *DBTrackerGitRepository) ByCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByCreateType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where createtype = $1", p)
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
func (a *DBTrackerGitRepository) ByLikeCreateType(ctx context.Context, p uint32) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeCreateType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where createtype ilike $1", p)
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

// get all "DBTrackerGitRepository" rows with matching RepositoryID
func (a *DBTrackerGitRepository) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where repositoryid = $1", p)
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
func (a *DBTrackerGitRepository) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where repositoryid ilike $1", p)
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

// get all "DBTrackerGitRepository" rows with matching URLHost
func (a *DBTrackerGitRepository) ByURLHost(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByURLHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where urlhost = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeURLHost(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeURLHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where urlhost ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching URLPath
func (a *DBTrackerGitRepository) ByURLPath(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByURLPath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where urlpath = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLPath: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLPath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeURLPath(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeURLPath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where urlpath ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLPath: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByURLPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching RepositoryCreated
func (a *DBTrackerGitRepository) ByRepositoryCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByRepositoryCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where repositorycreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeRepositoryCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeRepositoryCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where repositorycreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching SourceInstalled
func (a *DBTrackerGitRepository) BySourceInstalled(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySourceInstalled"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where sourceinstalled = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceInstalled: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceInstalled: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeSourceInstalled(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeSourceInstalled"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where sourceinstalled ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceInstalled: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceInstalled: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PackageID
func (a *DBTrackerGitRepository) ByPackageID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPackageID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where packageid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePackageID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePackageID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where packageid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PackageName
func (a *DBTrackerGitRepository) ByPackageName(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPackageName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where packagename = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePackageName(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePackageName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where packagename ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPackageName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ProtoFilename
func (a *DBTrackerGitRepository) ByProtoFilename(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoFilename"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where protofilename = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoFilename: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoFilename: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeProtoFilename(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeProtoFilename"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where protofilename ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoFilename: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoFilename: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ProtoSubmitted
func (a *DBTrackerGitRepository) ByProtoSubmitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoSubmitted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where protosubmitted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoSubmitted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoSubmitted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeProtoSubmitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeProtoSubmitted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where protosubmitted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoSubmitted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoSubmitted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ProtoCommitted
func (a *DBTrackerGitRepository) ByProtoCommitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByProtoCommitted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where protocommitted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoCommitted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoCommitted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeProtoCommitted(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeProtoCommitted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where protocommitted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoCommitted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProtoCommitted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching MinProtoVersion
func (a *DBTrackerGitRepository) ByMinProtoVersion(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByMinProtoVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where minprotoversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMinProtoVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMinProtoVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeMinProtoVersion(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeMinProtoVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where minprotoversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMinProtoVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMinProtoVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching Context
func (a *DBTrackerGitRepository) ByContext(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByContext"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where context = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeContext(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeContext"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where context ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching UserID
func (a *DBTrackerGitRepository) ByUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PermissionsCreated
func (a *DBTrackerGitRepository) ByPermissionsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPermissionsCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where permissionscreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPermissionsCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPermissionsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePermissionsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePermissionsCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where permissionscreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPermissionsCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPermissionsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching SecureArgsCreated
func (a *DBTrackerGitRepository) BySecureArgsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySecureArgsCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where secureargscreated = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySecureArgsCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySecureArgsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeSecureArgsCreated(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeSecureArgsCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where secureargscreated ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySecureArgsCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySecureArgsCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ServiceID
func (a *DBTrackerGitRepository) ByServiceID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where serviceid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeServiceID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeServiceID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where serviceid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ServiceUserID
func (a *DBTrackerGitRepository) ByServiceUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where serviceuserid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeServiceUserID(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeServiceUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where serviceuserid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching ServiceToken
func (a *DBTrackerGitRepository) ByServiceToken(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByServiceToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where servicetoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeServiceToken(ctx context.Context, p string) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeServiceToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where servicetoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching Finalised
func (a *DBTrackerGitRepository) ByFinalised(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByFinalised"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where finalised = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinalised: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinalised: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeFinalised(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeFinalised"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where finalised ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinalised: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinalised: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching PatchRepo
func (a *DBTrackerGitRepository) ByPatchRepo(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByPatchRepo"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where patchrepo = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPatchRepo: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPatchRepo: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikePatchRepo(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikePatchRepo"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where patchrepo ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPatchRepo: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPatchRepo: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching SourceRepositoryID
func (a *DBTrackerGitRepository) BySourceRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_BySourceRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where sourcerepositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeSourceRepositoryID(ctx context.Context, p uint64) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeSourceRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where sourcerepositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBTrackerGitRepository" rows with matching NotificationSent
func (a *DBTrackerGitRepository) ByNotificationSent(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByNotificationSent"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where notificationsent = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotificationSent: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotificationSent: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTrackerGitRepository) ByLikeNotificationSent(ctx context.Context, p bool) ([]*savepb.TrackerGitRepository, error) {
	qn := "DBTrackerGitRepository_ByLikeNotificationSent"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent from "+a.SQLTablename+" where notificationsent ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotificationSent: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotificationSent: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBTrackerGitRepository) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.TrackerGitRepository, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBTrackerGitRepository) Tablename() string {
	return a.SQLTablename
}

func (a *DBTrackerGitRepository) SelectCols() string {
	return "id,createrequestid, createtype, repositoryid, urlhost, urlpath, repositorycreated, sourceinstalled, packageid, packagename, protofilename, protosubmitted, protocommitted, minprotoversion, context, userid, permissionscreated, secureargscreated, serviceid, serviceuserid, servicetoken, finalised, patchrepo, sourcerepositoryid, notificationsent"
}
func (a *DBTrackerGitRepository) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".createrequestid, " + a.SQLTablename + ".createtype, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".urlhost, " + a.SQLTablename + ".urlpath, " + a.SQLTablename + ".repositorycreated, " + a.SQLTablename + ".sourceinstalled, " + a.SQLTablename + ".packageid, " + a.SQLTablename + ".packagename, " + a.SQLTablename + ".protofilename, " + a.SQLTablename + ".protosubmitted, " + a.SQLTablename + ".protocommitted, " + a.SQLTablename + ".minprotoversion, " + a.SQLTablename + ".context, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".permissionscreated, " + a.SQLTablename + ".secureargscreated, " + a.SQLTablename + ".serviceid, " + a.SQLTablename + ".serviceuserid, " + a.SQLTablename + ".servicetoken, " + a.SQLTablename + ".finalised, " + a.SQLTablename + ".patchrepo, " + a.SQLTablename + ".sourcerepositoryid, " + a.SQLTablename + ".notificationsent"
}

func (a *DBTrackerGitRepository) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.TrackerGitRepository, error) {
	var res []*savepb.TrackerGitRepository
	for rows.Next() {
		foo := savepb.TrackerGitRepository{}
		err := rows.Scan(&foo.ID, &foo.CreateRequestID, &foo.CreateType, &foo.RepositoryID, &foo.URLHost, &foo.URLPath, &foo.RepositoryCreated, &foo.SourceInstalled, &foo.PackageID, &foo.PackageName, &foo.ProtoFilename, &foo.ProtoSubmitted, &foo.ProtoCommitted, &foo.MinProtoVersion, &foo.Context, &foo.UserID, &foo.PermissionsCreated, &foo.SecureArgsCreated, &foo.ServiceID, &foo.ServiceUserID, &foo.ServiceToken, &foo.Finalised, &foo.PatchRepo, &foo.SourceRepositoryID, &foo.NotificationSent)
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
func (a *DBTrackerGitRepository) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null  ,createtype integer not null  ,repositoryid bigint not null  ,urlhost text not null  ,urlpath text not null  ,repositorycreated boolean not null  ,sourceinstalled boolean not null  ,packageid text not null  ,packagename text not null  ,protofilename text not null  ,protosubmitted boolean not null  ,protocommitted boolean not null  ,minprotoversion bigint not null  ,context text not null  ,userid text not null  ,permissionscreated boolean not null  ,secureargscreated boolean not null  ,serviceid text not null  ,serviceuserid text not null  ,servicetoken text not null  ,finalised boolean not null  ,patchrepo boolean not null  ,sourcerepositoryid bigint not null  ,notificationsent boolean not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),createrequestid bigint not null  ,createtype integer not null  ,repositoryid bigint not null  ,urlhost text not null  ,urlpath text not null  ,repositorycreated boolean not null  ,sourceinstalled boolean not null  ,packageid text not null  ,packagename text not null  ,protofilename text not null  ,protosubmitted boolean not null  ,protocommitted boolean not null  ,minprotoversion bigint not null  ,context text not null  ,userid text not null  ,permissionscreated boolean not null  ,secureargscreated boolean not null  ,serviceid text not null  ,serviceuserid text not null  ,servicetoken text not null  ,finalised boolean not null  ,patchrepo boolean not null  ,sourcerepositoryid bigint not null  ,notificationsent boolean not null  );`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS createrequestid bigint not null default 0;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS createtype integer not null default 0;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS urlhost text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS urlpath text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS repositorycreated boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS sourceinstalled boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS packageid text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS packagename text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS protofilename text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS protosubmitted boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS protocommitted boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS minprotoversion bigint not null default 0;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS context text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS permissionscreated boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS secureargscreated boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS serviceid text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS serviceuserid text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS servicetoken text not null default '';`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS finalised boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS patchrepo boolean not null default false;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS sourcerepositoryid bigint not null default 0;`,
		`ALTER TABLE trackergitrepository ADD COLUMN IF NOT EXISTS notificationsent boolean not null default false;`,
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
func (a *DBTrackerGitRepository) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
