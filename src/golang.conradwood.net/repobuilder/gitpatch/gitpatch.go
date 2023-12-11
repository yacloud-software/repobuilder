package gitpatch

import (
	"context"
	"flag"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type GitReference struct {
	repoid       uint64
	commit       string
	localgitdir  string // top of git repository
	workdir      string
	inuse        bool
	addedfiles   []string
	needs_commit bool
}

var (
	late_patch_workdir  = flag.String("late_patch_workdir", "lwd", "working directory for late patch files")
	clone_counter       = uint64(0)
	clone_counter_mutex sync.Mutex
)

func GetNextCloneCounter() uint64 {
	clone_counter_mutex.Lock()
	clone_counter++
	clone_counter_mutex.Unlock()
	return clone_counter
}

// get's latest. reference refers to a repository and a commit
func GetRepoReferenceByID(ctx context.Context, repoid uint64) (*GitReference, error) {
	repo, err := gitpb.GetGIT2Client().RepoByID(ctx, &gitpb.ByIDRequest{ID: repoid})
	if err != nil {
		return nil, err
	}

	gr := &GitReference{inuse: true}
	l := linux.New()
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	cc := GetNextCloneCounter()
	gr.workdir = fmt.Sprintf("%s/%d/%d", *late_patch_workdir, repoid, cc)
	gr.workdir, err = filepath.Abs(gr.workdir)
	if err != nil {
		return nil, err
	}
	gr.localgitdir = gr.workdir + "/repo"
	os.MkdirAll(gr.workdir, 0777)

	surl := repo.URLs[0]
	url := fmt.Sprintf("https://%s/git//%s", surl.Host, surl.Path)
	fmt.Printf("Cloning git repo %s into %s...\n", url, gr.workdir)
	//	c.GitSetAuth(url)
	out, err := l.SafelyExecuteWithDir([]string{"git", "clone", url, "repo"}, gr.workdir, nil)
	if err != nil {
		fmt.Printf("Error (clone). Git said: %s\n", out)
		return nil, err
	}
	return gr, nil
}
func (gr *GitReference) Close() {
	gr.inuse = false
	//	os.RemoveAll(gr.workdir)
}
func (gr *GitReference) AddFile(filename string, content []byte) error {
	ffilename := fmt.Sprintf("%s/%s", gr.localgitdir, filename)
	fp := filepath.Dir(ffilename)
	os.MkdirAll(fp, 0777)
	r, err := utils.ReadFile(ffilename)
	if err == nil && string(r) == string(content) {
		// exists already
		return nil
	}
	err = utils.WriteFile(ffilename, content)
	if err != nil {
		return err
	}
	fmt.Printf("Adding file \"%s\"\n", filename)
	l := linux.New()
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	out, err := l.SafelyExecuteWithDir([]string{"git", "add", filename}, gr.localgitdir, nil)
	if err != nil {
		fmt.Printf("Error (add). Git said: %s\n", out)
		return err
	}
	gr.needs_commit = true
	return nil
}
func (gr *GitReference) NeedsCommit() bool {
	return gr.needs_commit
}
func (gr *GitReference) CommitAndPush() error {
	fmt.Printf("Commit and Pushing...\n")
	l := linux.New()
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	out, err := l.SafelyExecuteWithDir([]string{"git", "commit", "-a", "-m", "repobuilder patches"}, gr.localgitdir, nil)
	if err != nil {
		fmt.Printf("Error (commit). Git said: %s\n", out)
		return err
	}
	l = linux.New()
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	out, err = l.SafelyExecuteWithDir([]string{"git", "push"}, gr.localgitdir, nil)
	if err != nil {
		fmt.Printf("Error (push). Git said: %s\n", out)
		return err
	}
	gr.needs_commit = false
	return nil
}

func (gr *GitReference) GitDirAbsFilename() string {
	return gr.localgitdir
}


