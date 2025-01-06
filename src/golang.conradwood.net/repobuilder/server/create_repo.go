package main

import (
	"context"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/artefact"
	"golang.conradwood.net/apis/buildrepo"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/email"
	gitpb "golang.conradwood.net/apis/gitserver"
	pb "golang.conradwood.net/apis/repobuilder"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"os"

	//	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	git_home_dir     string
	git_max_run_time = flag.Duration("git_max_runtime", time.Duration(600)*time.Second, "max runtime in seconds of a git process before it is killed")
	git_dir          = flag.String("git_dir", "/tmp/git_dirs", "the directory in which to clone git")
	creatorLock      sync.Mutex
)

type Creator struct {
	req         *pb.CreateWebRepoRequest
	ctx         context.Context
	tgr         *pb.TrackerGitRepository
	rcs         *pb.RepoCreateStatus
	gitdir      string
	needscommit bool
	clonedRepo  bool // true if the repo has been cloned to the local filesystem
	err         error
	errtopic    string
}

/******************************************************************************
* pick up all incomplete repos and try to complete them...
******************************************************************************/
func create_all_web_repos() {
	creatorLock.Lock()
	defer creatorLock.Unlock()
	fmt.Printf("Running through repository requests again...\n")
	err := utils.RecreateSafely(*git_dir)
	if err != nil {
		fmt.Printf("Failed to create dir \"%s\": %s\n", *git_dir, err)
		return
	}
	ctx := context.Background()
	rs, err := TrackerGitRepository_store.All(ctx)
	if err != nil {
		fmt.Printf("Could not get webrequests: %s\n", err)
		return
	}
	for _, r := range rs {
		if r.Finalised {
			continue
		}
		err = create_web_repo(r)
		if err != nil {
			fmt.Printf("#%d Failed to create web repo: %s\n", r.ID, utils.ErrorString(err))
		} else {
			fmt.Printf("#%d Repo complete.\n", r.ID)
		}
	}
}

/******************************************************************************
* the main sequence of steps for a repo to be created...
******************************************************************************/
func create_web_repo(req *pb.TrackerGitRepository) error {
	if req.Finalised {
		return nil
	}
	err := createGitConfig()
	if err != nil {
		return errors.Errorf("failed to create git config %w", err)
	}
	c := &Creator{tgr: req}
	err = c.setup()
	if err != nil {
		return errors.Errorf("failed to setup %w", err)
	}
	c.create()
	return nil
}
func trigger_create_web_repo(req *pb.TrackerGitRepository) error {
	if req.Finalised {
		return nil
	}
	err := createGitConfig()
	if err != nil {
		return err
	}
	c := &Creator{tgr: req}
	err = c.setup()
	if err != nil {
		return err
	}
	if c.rcs.Success {
		// it's done. do not do anything.
		return nil
	}

	go c.create()
	return nil
}

// step1 quick/databasy stuff (within RPC speed)
func (c *Creator) setup() error {
	err := c.RestoreContext()
	if err != nil {
		return err
	}
	err = c.LoadCreateReq()
	if err != nil {
		return err
	}
	err = c.LoadStatus()
	if err != nil {
		return err
	}
	if c.rcs.Success {
		c.Printf("Repository completed already\n")
		// it's done. do not do anything.
		return nil
	}

	err = c.GitRepository()
	if err != nil {
		return err
	}
	c.Printf("setup of git Repo at: %s\n", c.GitCloneURL())
	if c.rcs.Success {
		return nil
	}

	c.Printf("Got Git Repository %d at %s for tracker #%d\n", c.tgr.RepositoryID, c.GitCloneURL(), c.tgr.ID)
	return nil
}

// step2 - long running stuff
func (c *Creator) create() {
	if c.rcs.Success {
		c.Printf("Repository completed already\n")
		// it's done. do not do anything.
		return
	}
	err := c.RestoreContext()
	if err != nil {
		return
	}
	if c.SetError("loadcreatereq", c.LoadCreateReq()) {
		return
	}
	// execute stuff
	c.createWithCleanup()

	// cleanup and store state afterwards

	if c.needscommit {
		if c.SetError("gitcommit", c.GitCommit()) {
			return
		}
		c.tgr.SourceInstalled = true
	}
	c.SetGitHooks(true) // make sure we at least do a best effort attempt at re-enabling the githooks
	if c.err == nil {
		c.rcs.Success = true
		c.tgr.Finalised = true
	}

	c.SetError("save_progress", c.SaveProgress()) // always save progress so far
	c.SetError("save_status", c.SaveStatus())
	details := fmt.Sprintf("[trackergitrepo id=%d, createwebreporequest id=%d, repocreatestatus id = %d] ", c.tgr.ID, c.req.ID, c.rcs.ID)
	if c.err != nil {
		c.Printf("%sGit repository %s at %s failed.\n", details, c.req.RepoName, c.GitCloneURL())
		c.Printf("failed task \"%s\": %s\n", c.errtopic, errors.ErrorStringWithStackTrace(c.err))
	} else {
		c.RelinquishRepo()
		c.Printf("Git repository %s at %s completed successfully.\n", c.req.RepoName, c.GitCloneURL())
		c.notifyCreated()
	}
}

func (c *Creator) createWithCleanup() {

	if !c.tgr.PatchRepo {
		if c.SetError("clonerepo_nopatch", c.CloneRepo()) {
			return
		}
		s := fmt.Sprintf("created by repobuilder at %v\n", time.Now())
		if c.SetError("write_info", utils.WriteFile(c.GitDir()+"repo_info.txt", []byte(s))) {
			return
		}
		if c.SetError("add_nopatch", c.gitadd()) {
			return
		}
		c.needscommit = true
		return
	}
	if !c.tgr.SourceInstalled {
		if c.SetError("modify_template", c.ModifyTemplate()) {
			return
		}
	}

	if !c.tgr.ProtoSubmitted {
		if c.SetError("submit_proto", c.SubmitProto()) {
			return
		}
	}
	if !c.tgr.ProtoCommitted {
		if c.SetError("clone_repo", c.CloneRepo()) {
			return
		}

		/*
			repopath, err := filepath.Abs(c.GitDir())
			if err != nil {
				c.SetError("Absolute filepath", err)
				return
			}
		*/
		err := c.RestoreContext()
		if err != nil {
			c.SetError("pre_update_protos", err)
			return
		}

		c.needscommit = true
		c.tgr.ProtoCommitted = true
	}
	if c.SetError("create_service_account", c.CreateServiceAccount()) {
		return
	}
	if !c.tgr.SecureArgsCreated {
		if c.SetError("create_args", c.CreateSecureArgs()) {
			return
		}
		c.tgr.SecureArgsCreated = true
	}
	if !c.tgr.PermissionsCreated {
		if c.SetError("create_permissions", c.CreatePermissions()) {
			return
		}
		c.tgr.PermissionsCreated = true
	}
}

func (c *Creator) SetGitHooks(run bool) error {
	// we always attempt to use a new context. the old one might have expired...
	err := c.RestoreContext()
	if err != nil {
		return err
	}
	sr := &gitpb.SetRepoFlagsRequest{
		RepoID:         c.tgr.RepositoryID,
		RunPostReceive: run,
		RunPreReceive:  run,
	}
	_, err = gitpb.GetGIT2Client().SetRepoFlags(c.ctx, sr)
	if err != nil {
		c.Printf("Failed to set repo flags: %s\n", utils.ErrorString(err))
	}
	c.Printf("Git repo hooks set to %v\n", run)
	return err
}

// commit and push the repo
func (c *Creator) GitCommit() error {
	err := c.SetGitHooks(false)
	if err != nil {
		return err
	}
	c.Printf("Committing git repo...\n")
	dir := fmt.Sprintf("%s/%d/repo", *git_dir, c.req.ID)
	os.MkdirAll(dir, 0777)
	url := c.GitCloneURL()
	c.GitSetAuth(url)
	out, err := rungit([]string{"git", "commit", "-a", "-m", "new repository created"}, dir, nil)
	if err != nil {
		c.Printf("Error Committing (%s). Git said: %s\n", err, out)
		return err
	}
	out, err = rungit([]string{"git", "push", "--set-upstream", "origin", "master"}, dir, nil)
	if err != nil {
		c.Printf("Error Push. Git said: %s\n", out)
		return err
	}
	c.needscommit = false
	c.Printf("Committed Git Repo\n")
	return nil
}

// ensure there is a git repository checked out which contains the template files
// and it's git "origin" points to the url of the new repository
func (c *Creator) CloneRepo() error {
	if c.clonedRepo {
		return nil
	}
	var err error
	if c.tgr.SourceInstalled {
		err = c.GitClone()
	} else {
		err = c.GitCloneSkel()
	}
	if err != nil {
		return errors.Errorf("Failed to clone repo: %s", err)
	}
	c.clonedRepo = true
	return nil
}

// create git repository if necessary, otherwise load it from db
func (c *Creator) GitRepository() error {
	gpath := strings.ToLower(c.req.RepoName) + ".git"
	c.tgr.URLHost = "git." + strings.ToLower(c.req.Domain)
	c.tgr.URLPath = "/git/" + gpath

	if c.tgr.RepositoryCreated {
		// if repo is created already, reset it
		bir := &gitpb.ByIDRequest{ID: c.tgr.RepositoryID}
		c.Printf("Resetting repository: %d\n", bir.ID)
		_, err := gitpb.GetGIT2Client().ResetRepository(authremote.Context(), bir)
		if err != nil {
			c.Printf("Failed to reset repo: %s\n", utils.ErrorString(err))
			return err
		}

	} else {
		// create repo if it has not been created yet
		crr := &gitpb.CreateRepoRequest{
			ArtefactName: c.req.RepoName,
			URL: &gitpb.SourceRepositoryURL{
				Host: c.tgr.URLHost,
				Path: gpath,
			},
			Description: c.req.Description,
		}
		sr, err := gitpb.GetGIT2Client().CreateRepo(c.ctx, crr)
		if err != nil {
			return err
		}
		c.Printf("Repository created: %d\n", sr.ID)
		c.tgr.RepositoryID = sr.ID
		c.tgr.RepositoryCreated = true
		err = c.SaveProgress()
		if err != nil {
			return err
		}
	}
	ctx := authremote.Context()
	bm, err := buildrepo.GetBuildRepoManagerClient().GetManagerInfo(ctx, &common.Void{})
	if err != nil {
		return err
	}
	brepodomain := bm.Domain

	ar, err := artefact.GetArtefactClient().CreateArtefactIfRequired(ctx,
		&artefact.CreateArtefactRequest{
			OrganisationID:  "repobuilder",
			ArtefactName:    c.req.RepoName,
			BuildRepoDomain: brepodomain,
		},
	)
	if err != nil {
		return err
	}
	c.tgr.ArtefactID = ar.Meta.ID

	return nil
}

func (c *Creator) SaveProgress() error {
	err := c.RestoreContext()
	if err != nil {
		return err
	}
	err = TrackerGitRepository_store.Update(c.ctx, c.tgr)
	return err
}
func (c *Creator) SaveStatus() error {
	err := c.RestoreContext()
	if err != nil {
		return err
	}
	err = RepoCreateStatus_store.Update(c.ctx, c.rcs)
	if err != nil {
		return err
	}
	return nil
}
func (c *Creator) LoadStatus() error {
	tss, err := RepoCreateStatus_store.ByCreateRequestID(c.ctx, c.req.ID)
	if err != nil {
		return err
	}
	if len(tss) != 0 {
		c.rcs = tss[0]
		return nil
	}
	ts := &pb.RepoCreateStatus{
		CreateRequestID: c.req.ID,
		CreateType:      TYPE_WEBREPO,
	}
	_, err = RepoCreateStatus_store.Save(c.ctx, ts)
	if err != nil {
		return err
	}
	c.rcs = ts
	return nil
}

// restore context from database
func (c *Creator) RestoreContext() error {
	ctx, err := authremote.ContextForUserID(c.tgr.UserID)
	if err != nil {
		return errors.Errorf("Unable to get context for user: %w", err)
	}
	c.ctx = ctx
	return nil
}

// return url for git
func (c *Creator) GitCloneURL() string {
	return fmt.Sprintf("https://%s%s", c.tgr.URLHost, c.tgr.URLPath)
}

// ensures it's cloned somewhere
func (c *Creator) GitClone() error {
	dir := fmt.Sprintf("%s/%d", *git_dir, c.req.ID)
	os.MkdirAll(dir, 0777)
	url := c.GitCloneURL()
	c.Printf("Cloning git repo %s...\n", url)
	c.GitSetAuth(url)
	out, err := rungit([]string{"git", "clone", url, "repo"}, dir, nil)
	if err != nil {
		c.Printf("Error. Git said: %s\n", out)
		return err
	}
	c.gitdir = dir + "/repo"
	return nil
}

// check out the "skel" and modify the URL to point back to the proper url
func (c *Creator) GitCloneSkel() error {
	dir := fmt.Sprintf("%s/%d", *git_dir, c.req.ID)
	os.RemoveAll(dir)
	os.MkdirAll(dir, 0777)
	repo, err := c.getGitRepo(c.tgr.SourceRepositoryID)
	if err != nil {
		return err
	}
	url := "https://git.conradwood.net/git/skel-go.git"
	url = fmt.Sprintf("https://%s/git/%s", repo.Host, repo.Path)
	c.Printf("Cloning git repo %s...\n", url)
	c.GitSetAuth(url)
	out, err := rungit([]string{"git", "clone", url, "repo"}, dir, nil)
	if err != nil {
		c.Printf("Error. Git said: %s\n", out)
		return err
	}
	config := `[core]
        repositoryformatversion = 0
        filemode = true
        bare = false
        logallrefupdates = true
[remote "origin"]
        url = %s
        fetch = +refs/heads/*:refs/remotes/origin/*
`
	cs := fmt.Sprintf(config, c.GitCloneURL())
	err = utils.WriteFile(dir+"/repo/.git/config", []byte(cs))
	if err != nil {
		return err
	}
	c.gitdir = dir + "/repo"
	return nil
}

// no error, presumably later stuff will fail
// TODO - authentication required here
func (c *Creator) GitSetAuth(url string) {
	// isn't that handled by gitconfig token thing?
}

func (c *Creator) GitDir() string {
	return c.gitdir
}

func (c *Creator) SetError(step string, err error) bool {
	if err == nil {
		return false
	}
	c.Printf("Error in %s creating repository: %s\n", step, errors.ErrorStringWithStackTrace(err))
	c.err = err
	c.errtopic = step
	return true
}
func (c *Creator) LoadCreateReq() error {
	if c.req != nil {
		return nil
	}
	crs, err := WebRepoRequest_store.ByID(c.ctx, c.tgr.CreateRequestID)
	if err != nil {
		return err
	}
	c.req = crs
	return nil

}

// last thing we do, is to tell gitserver that repobuilder relinguishes control
// of this repo. (This is one-way, once we done that we cannot get control back).
// this must be the very last and only when it is completed.
func (c *Creator) RelinquishRepo() error {
	c.Printf("Relinquishing control of repo to gitserver access checks\n")
	bir := &gitpb.ByIDRequest{ID: c.tgr.RepositoryID}
	_, err := gitpb.GetGIT2Client().RepoBuilderComplete(authremote.Context(), bir)
	if err != nil {
		c.SetError("set_git_ready", err)
		fmt.Printf("Failed to tell gitserver repobuilderComplete\n")
		return err
	}
	return nil
}
func (c *Creator) Printf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	prefix := "[noreq] "
	if c.req != nil {
		prefix = fmt.Sprintf("[%d %d %s] ", c.tgr.ID, c.req.ID, c.req.Name)
	}
	fmt.Printf("%s%s", prefix, s)
}

// write a git config into ~/.gitconfig which authenticates us as repobuilder
func createGitConfig() error {
	h, err := utils.HomeDir()
	if err != nil {
		return errors.Errorf("failed to get homedir %w", err)
	}
	git_home_dir = h + "/repobuilder"
	gitconfig := `[user]
        name = RepoBuilder
        email = repobuilder@services.yacloud.eu

[http]
        postBuffer = 524288000
        extraHeader = "Authorization: Bearer %s"
`
	gc := fmt.Sprintf(gitconfig, tokens.GetServiceTokenParameter())
	err = utils.WriteFileCreateDir(git_home_dir+"/.gitconfig", []byte(gc))
	return err
}

func (c *Creator) getGitRepo(repoid uint64) (*gitpb.SourceRepositoryURL, error) {
	ctx := c.ctx
	bir := &gitpb.ByIDRequest{ID: repoid}
	repo, err := gitpb.GetGIT2Client().RepoByID(ctx, bir)
	if err != nil {
		return nil, err
	}
	if len(repo.URLs) == 0 {
		return nil, errors.Errorf("Repo %d has no URLs", repoid)
	}
	return repo.URLs[0], nil
}

func (c *Creator) notifyCreated() {
	if c.tgr.NotificationSent {
		return
	}

	u := auth.GetUser(c.ctx)
	ter := &email.TemplateEmailRequest{
		TemplateName: "standard",
		Sender:       " repobuilder@yacloud.eu",
		Recipient:    u.Email,
		Values:       make(map[string]string),
	}
	ter.Values["subject"] = fmt.Sprintf("Repository %s created", c.req.Name)
	ter.Values["textbody"] = fmt.Sprintf(`Repository %s created and is now available at
%s`, c.req.Name, c.GitCloneURL())
	_, err := email.GetEmailClient().SendTemplate(c.ctx, ter)
	if err != nil {
		fmt.Printf("failed to send email: %s\n", utils.ErrorString(err))
	}
	c.tgr.NotificationSent = true
	c.RestoreContext()
	err = TrackerGitRepository_store.Update(c.ctx, c.tgr)
	if err != nil {
		fmt.Printf("Failed to store trackergitrepo: %s\n", err)
	}
}

func rungit(com []string, dir string, r io.Reader) (string, error) {
	l := linux.New()
	l.SetMaxRuntime(*git_max_run_time)
	out, err := l.SafelyExecuteWithDir([]string{"git", "commit", "-a", "-m", "new repository created"}, dir, r)
	return out, err
}
