package main

import (
	"context"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	gitpb "golang.conradwood.net/apis/gitserver"
	oa "golang.conradwood.net/apis/objectauth"
	pb "golang.conradwood.net/apis/repobuilder"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/utils"
	"golang.conradwood.net/repobuilder/db"
	"golang.conradwood.net/repobuilder/latepatching"
	"google.golang.org/grpc"
	"os"
	"strings"
	"time"
)

const (
	USER_APP_REPO = 311
	PROBE_WEBREPO = 36 // webrequest for repo #216  https://git.conradwood.net/git/proberrepo.git
	TYPE_WEBREPO  = uint32(1)
)

var (
	port                       = flag.Int("port", 4100, "The grpc server port")
	WebRepoRequest_store       *db.DBCreateWebRepoRequest
	TrackerLog_store           *db.DBTrackerLog
	RepoCreateStatus_store     *db.DBRepoCreateStatus
	TrackerGitRepository_store *db.DBTrackerGitRepository
)

type repoBuilderServer struct {
}

func main() {
	var err error
	flag.Parse()
	fmt.Printf("Starting RepoBuilderServer...\n")
	WebRepoRequest_store = db.DefaultDBCreateWebRepoRequest()
	TrackerLog_store = db.DefaultDBTrackerLog()
	TrackerGitRepository_store = db.DefaultDBTrackerGitRepository()
	RepoCreateStatus_store = db.DefaultDBRepoCreateStatus()
	create_all_web_repos()

	sd := server.NewServerDef()
	sd.SetPort(*port)
	sd.SetRegister(server.Register(
		func(server *grpc.Server) error {
			e := new(repoBuilderServer)
			pb.RegisterRepoBuilderServer(server, e)
			return nil
		},
	))
	err = server.ServerStartup(sd)
	utils.Bail("Unable to start server", err)
	os.Exit(0)
}

/************************************
* grpc functions
************************************/
func (e *repoBuilderServer) CreateUserFirmwareRepo(ctx context.Context, req *pb.CreateWebRepoRequest) (*pb.CreateRepoResponse, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "need user")
	}
	if req.Description == "" {
		s := "Description"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.Name == "" {
		s := "Name"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}

	req.Name = strings.ToLower(req.Name)

	surl := &gitpb.SourceRepositoryURL{Host: fmt.Sprintf("git.%s", req.Domain), Path: req.Name}
	if req.Domain == "" {
		surl = &gitpb.SourceRepositoryURL{Host: "git.singingcat.net", Path: fmt.Sprintf("%s/%s", u.ID, req.Name)}
	}

	bir := &gitpb.ForkRequest{
		Description:    req.Description,
		ArtefactName:   req.Name,
		RepositoryID:   USER_APP_REPO,
		URL:            surl,
		CreateReadOnly: true,
	}
	fr, err := gitpb.GetGIT2Client().Fork(ctx, bir)
	if err != nil {
		return nil, err
	}
	ob := oa.GetObjectAuthClient()
	// first grant the creator explicit rights to his new repo
	_, err = ob.GrantToUser(ctx, &oa.GrantUserRequest{
		ObjectType: oa.OBJECTTYPE_GitRepository,
		ObjectID:   fr.ID,
		UserID:     u.ID,
		Read:       true,
		Write:      true,
		Execute:    true,
		View:       true,
	})
	if err != nil {
		return nil, err
	}
	err = SetAdditionalPermissions(ctx, req, fr.ID)
	if err != nil {
		return nil, err
	}
	crr := &pb.CreateRepoResponse{
		RequestID: 0,
		Finished:  true,
		Success:   true,
		Error:     "",
		URL:       fmt.Sprintf("https://%s/git/%s.git", surl.Host, surl.Path),
	}
	lpq := &pb.LatePatchingQueue{RepositoryID: fr.ID, EntryCreated: uint32(time.Now().Unix())}
	_, err = db.DefaultDBLatePatchingQueue().Save(ctx, lpq)
	if err != nil {
		return nil, err
	}
	latepatching.Trigger()
	return crr, nil
}
func (e *repoBuilderServer) CreateWebRepo(ctx context.Context, req *pb.CreateWebRepoRequest) (*pb.CreateRepoResponse, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "need user")
	}
	if req.Description == "" {
		s := "Description"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.Name == "" {
		s := "Name"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.Language == 0 {
		s := "Language"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.Domain == "" {
		s := "Domain"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.RepoName == "" {
		s := "RepoName"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.ServiceName == "" {
		s := "ServiceName"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.ProtoDomain == "" {
		req.ProtoDomain = req.Domain
	}

	id, err := WebRepoRequest_store.Save(ctx, req)
	if err != nil {
		return nil, err
	}

	// store stuff that we need to process it
	tu := &pb.TrackerGitRepository{
		CreateRequestID: id,
		CreateType:      TYPE_WEBREPO,
		UserID:          u.ID,
		PatchRepo:       true, // patch repo to match servicename and deployment etc
	}
	tu.SourceRepositoryID = LanguageToRepo(req.Language)
	_, err = TrackerGitRepository_store.Save(ctx, tu)
	if err != nil {
		return nil, err
	}
	err = trigger_create_web_repo(tu)
	if err != nil {
		fmt.Printf("Failed to create web repo: %s\n", utils.ErrorString(err))
		return nil, err
	}

	res := &pb.CreateRepoResponse{RequestID: id}
	return res, nil
}

func (e *repoBuilderServer) RecreateWebRepo(ctx context.Context, req *pb.CreateWebRepoRequest) (*pb.CreateRepoResponse, error) {
	var err error
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.AccessDenied(ctx, "cannot recreate w/o user")
	}
	err = errors.NeedsRoot(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Resetting webrequest\n")
	nreq, err := WebRepoRequest_store.ByID(ctx, PROBE_WEBREPO)
	if err != nil {
		return nil, err
	}

	nreq.Language = req.Language

	err = WebRepoRequest_store.Update(ctx, nreq)
	if err != nil {
		return nil, err
	}

	tgsX, err := TrackerGitRepository_store.ByCreateRequestID(ctx, nreq.ID)
	if err != nil {
		return nil, err
	}
	if len(tgsX) == 0 {
		return nil, fmt.Errorf("no tracker for request")
	}
	tgs := tgsX[0]

	// tell gitserver that we're going again..
	bir := &gitpb.ByIDRequest{ID: tgs.RepositoryID}
	_, err = gitpb.GetGIT2Client().ResetRepository(authremote.Context(), bir)
	if err != nil {
		fmt.Printf("Failed to reset repository: %s\n", utils.ErrorString(err))
		return nil, err
	}

	tgs.UserID = u.ID
	// which steps do we want to repeat? marking those as 'false'
	tgs.SourceInstalled = false
	tgs.Finalised = false
	tgs.ProtoSubmitted = false
	tgs.ProtoCommitted = false
	err = TrackerGitRepository_store.Update(ctx, tgs)
	if err != nil {
		return nil, err
	}

	// also delete any status we might still have:
	rcsX, err := RepoCreateStatus_store.ByCreateRequestID(ctx, nreq.ID)
	for _, rcs := range rcsX {
		err = RepoCreateStatus_store.DeleteByID(ctx, rcs.ID)
		if err != nil {
			return nil, err
		}
	}
	err = trigger_create_web_repo(tgs)
	if err != nil {
		fmt.Printf("Failed to create web repo: %s\n", utils.ErrorString(err))
		return nil, err
	}

	res := &pb.CreateRepoResponse{RequestID: nreq.ID}
	return res, nil

}

func (e *repoBuilderServer) Fork(ctx context.Context, req *pb.ForkRequest) (*pb.CreateRepoResponse, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "need user")
	}
	if req.Name == "" {
		s := "Name"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.Domain == "" {
		s := "Domain"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}
	if req.RepoName == "" {
		s := "RepoName"
		return nil, errors.InvalidArgs(ctx, s+" required", s+" required")
	}

	wr := &pb.CreateWebRepoRequest{
		Name:     req.Name,
		Domain:   req.Domain,
		RepoName: req.RepoName,
	}

	id, err := WebRepoRequest_store.Save(ctx, wr)
	if err != nil {
		return nil, err
	}

	// store stuff that we need to process it
	tu := &pb.TrackerGitRepository{
		CreateRequestID:    id,
		CreateType:         TYPE_WEBREPO,
		UserID:             u.ID,
		PatchRepo:          false, // patch repo to match servicename and deployment etc
		SourceRepositoryID: req.RepositoryID,
	}
	_, err = TrackerGitRepository_store.Save(ctx, tu)
	if err != nil {
		return nil, err
	}
	err = trigger_create_web_repo(tu)
	if err != nil {
		fmt.Printf("Failed to create web repo: %s\n", utils.ErrorString(err))
		return nil, err
	}

	res := &pb.CreateRepoResponse{RequestID: id}
	return res, nil

}
func (e *repoBuilderServer) GetRepoChoices(ctx context.Context, req *common.Void) (*pb.Choices, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "authentication required")
	}
	res := &pb.Choices{}
	for k, v := range common.ProgrammingLanguage_name {
		l := &pb.Language{ID: uint64(k), Name: v}
		res.Languages = append(res.Languages, l)
	}
	for _, g := range u.Groups {
		res.Groups = append(res.Groups, g)
	}
	res.Domains = []*pb.RepoDomain{
		&pb.RepoDomain{Domain: "yacloud.eu"},
		&pb.RepoDomain{Domain: "youritguru.com"},
		&pb.RepoDomain{Domain: "safeservers.eu"},
	}
	if auth.IsRoot(ctx) {
		res.Domains = append(res.Domains, &pb.RepoDomain{Domain: "conradwood.net"})
		res.Domains = append(res.Domains, &pb.RepoDomain{Domain: "singingcat.net"})
	}

	return res, nil
}
func (e *repoBuilderServer) RetriggerAll(ctx context.Context, req *common.Void) (*common.Void, error) {
	if !auth.IsRoot(ctx) {
		return nil, errors.AccessDenied(ctx, "this is a root only function")
	}
	go create_all_web_repos()
	return req, nil
}
func (e *repoBuilderServer) GetRepoStatus(ctx context.Context, req *pb.RepoStatusRequest) (*pb.CreateRepoResponse, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "not authenticated")
	}
	wrr, err := WebRepoRequest_store.ByID(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}
	tgrs, err := TrackerGitRepository_store.ByCreateRequestID(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}
	if len(tgrs) == 0 {
		return nil, errors.NotFound(ctx, "no such request")
	}
	tgr := tgrs[0]
	if tgr.UserID != u.ID {
		return nil, errors.AccessDenied(ctx, "access to repository create request denied (user \"%s\" asked for repo by user \"%s\")", u.ID, tgr.UserID)
	}
	tss, err := RepoCreateStatus_store.ByCreateRequestID(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}
	if len(tss) == 0 {
		return nil, errors.NotFound(ctx, "no repocreate estatus")
	}
	rcs := tss[0]
	res := &pb.CreateRepoResponse{
		RequestID: wrr.ID,
		Finished:  rcs.Success,
		Success:   rcs.Success,
		Error:     rcs.Error,
	}
	return res, nil
}

// given a language, will return a repo to clone
func LanguageToRepo(language common.ProgrammingLanguage) uint64 {
	if language == common.ProgrammingLanguage_GO {
		return 64
	}
	return 0
}



