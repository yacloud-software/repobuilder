package main

import (
	"fmt"
	pr "golang.conradwood.net/apis/protorenderer"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
	"strings"
	//	"time"
)

// this will block until proto is ready!
func (c *Creator) SubmitProto() error {
	rerr := c.CloneRepo()
	if rerr != nil {
		return rerr
	}
	fn := c.tgr.ProtoFilename
	c.Printf("Handling proto %s\n", fn)
	b, err := utils.ReadFile(c.GitDir() + fn)
	if err != nil {
		return err
	}
	err = c.RestoreContext()
	if err != nil {
		return err
	}
	// compile .proto:
	protoclient := pr.GetProtoRendererClient()

	pf := &pr.AddProtoRequest{
		Name:    fn,
		Content: string(b),
	}
	v, err := protoclient.CompileFile(c.ctx, &pr.CompileRequest{
		Compilers:       []pr.CompilerType{pr.CompilerType_GOLANG},
		AddProtoRequest: pf})
	if err != nil {
		fmt.Printf("Compile Error: %s\n", v.CompileError)
		return err
	}
	c.tgr.PackageName = "foo_compiled_handle_proto.package"
	c.tgr.ProtoSubmitted = true

	fmt.Printf("Directory: %s\n", c.gitdir)
	for _, f := range v.Files {
		fname := f.Filename
		fmt.Printf("Saving file \"%s\"\n", fname)
		fname = c.fixCompiledProtoFilename(fname)
		fmt.Printf("Saving proto file to \"%s\"\n", fname)
		err = utils.WriteFile(fname, f.Content)
		if err != nil {
			return err
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
	fname = strings.TrimPrefix(fname, "protos")
	if c.UseVendor() {
		fname = c.GoVendorDir() + "/" + fname
	} else {
		fname = c.GitDir() + "/src/" + fname
	}
	dir := filepath.Dir(fname)
	os.MkdirAll(dir, 0777)
	return fname

}





