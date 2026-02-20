package gitrun

import (
	"context"
	"time"

	"golang.conradwood.net/go-easyops/linux"
)

func GitRun(nctx context.Context, com_and_args []string, dir string) (string, error) {
	/*
		ctx_s, err := ctx.SerialiseContextToString(nctx)
		if err != nil {
			return "", err
		}
	*/
	l := linux.NewWithContext(nctx)
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	/*
		env := os.Environ()
		env = append(env, "GE_CTX="+ctx_s)
		l.SetEnvironment(env)
	*/
	out, err := l.SafelyExecuteWithDir(com_and_args, dir, nil)
	return out, err
}
