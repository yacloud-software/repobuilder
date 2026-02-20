package gitrun

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	git_home_dir       string
	config_created     = false
	config_create_lock sync.Mutex
)

func GitRun(nctx context.Context, com_and_args []string, dir string) (string, error) {
	CreateGitConfig()
	/*
		ctx_s, err := ctx.SerialiseContextToString(nctx)
		if err != nil {
			return "", err
		}
	*/
	l := linux.NewWithContext(nctx)
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	l.SetEnvironment([]string{
		"HOME=" + git_home_dir,
		"PATH=" + os.Getenv("PATH"),
	})

	/*
		env := os.Environ()
		env = append(env, "GE_CTX="+ctx_s)
		l.SetEnvironment(env)
	*/
	out, err := l.SafelyExecuteWithDir(com_and_args, dir, nil)
	return out, err
}

// write a git config into ~/.gitconfig which authenticates us as repobuilder
func CreateGitConfig() error {
	config_create_lock.Lock()
	defer config_create_lock.Unlock()
	if config_created {
		return nil
	}
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
	if err != nil {
		return err
	}
	config_created = true
	return nil
}
