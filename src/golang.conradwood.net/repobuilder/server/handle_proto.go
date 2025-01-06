package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"golang.yacloud.eu/apis/protomanager"
	//	"time"
)

type file_result interface {
	GetFilename() string
	GetContent() []byte
}

// this will block until proto is ready!
func (c *Creator) SubmitProto() error {
	rerr := c.CloneRepo()
	if rerr != nil {
		return rerr
	}
	fn := c.tgr.ProtoFilename
	c.Printf("Handling proto %s\n", fn)
	fullfile := c.GitDir() + fn
	if !utils.FileExists(fullfile) {
		return errors.Errorf("File \"%s\" does not exist", fullfile)
	}
	b, err := utils.ReadFile(fullfile)
	if err != nil {
		return errors.Wrap(err)
	}
	err = c.RestoreContext()
	if err != nil {
		return err
	}
	// compile .proto:
	var files []file_result

	fn = strings.TrimPrefix(fn, "/")
	fn = strings.TrimPrefix(fn, "protos/")
	cr := &protomanager.SimpleCompileRequest{
		Filename: fn,
		Content:  b,
	}
	v, err := protomanager.GetProtoManagerClient().SimpleCompileGo(c.ctx, cr)
	if err != nil {
		fmt.Printf("(1a) Compile Error: %s\n", err)
		return err
	}
	if v.ErrorMessage != "" {
		fmt.Printf("(2a) Compile Error: %s\n", v.ErrorMessage)
		return errors.Errorf("%s", v.ErrorMessage)
	}
	for _, f := range v.GetFiles() {
		fmt.Printf("Received file \"%s\"\n", f.Filename)
		files = append(files, f)
	}

	c.tgr.ProtoSubmitted = true

	fmt.Printf("Directory: %s\n", c.gitdir)
	for _, f := range files {
		fname := f.GetFilename()
		fmt.Printf("Saving file \"%s\"\n", fname)
		fname = c.fixCompiledProtoFilename(fname)
		if fname == "" {
			continue
		}
		fmt.Printf("Saving proto file to \"%s\"\n", fname)
		err = utils.WriteFile(fname, f.GetContent())
		if err != nil {
			return errors.Wrap(err)
		}
		gname := strings.TrimPrefix(fname, c.GitDir())
		gname = strings.TrimPrefix(gname, "/")
		out, err := linux.New().SafelyExecuteWithDir([]string{"git", "add", gname}, c.GitDir(), nil)
		if err != nil {
			fmt.Printf("git add \"%s\" failed.\n", gname)
			fmt.Printf("git said: %s\n", out)
			fmt.Printf("git add failed: %s\n", err)
		}
	}
	return nil
}

func (c *Creator) fixCompiledProtoFilename(fname string) string {
	fname = strings.TrimPrefix(fname, "/")
	if !strings.HasPrefix(fname, "golang") {
		return ""
	}
	fname = strings.TrimPrefix(fname, "golang/")
	fname = strings.TrimPrefix(fname, "protos")
	fname = c.GitDir() + "/src/" + fname
	dir := filepath.Dir(fname)
	os.MkdirAll(dir, 0777)
	return fname

}
