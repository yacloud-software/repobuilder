package protos

import (
	"context"
	"fmt"
	pr "golang.conradwood.net/apis/protorenderer"
	"golang.conradwood.net/go-easyops/utils"
	"path/filepath"
)

type ProtoCompileResult struct {
	Files []*File // files relative to dir, which are new or changed
}
type File struct {
	Filename string
	Content  []byte
}

func Compile(ctx context.Context, repoid uint64, dir, filename string) (*ProtoCompileResult, error) {
	fmt.Printf("Compiling %s\n", dir+"/"+filename)
	b, err := utils.ReadFile(dir + "/" + filename)
	if err != nil {
		return nil, err
	}
	creq := &pr.CompileRequest{
		Compilers: []pr.CompilerType{pr.CompilerType_NANOPB},
		AddProtoRequest: &pr.AddProtoRequest{
			Content:      string(b),
			Name:         filename,
			RepositoryID: repoid,
		},
	}
	if utils.FileExists(dir + "/proto_files") {
		creq.Compilers = append(creq.Compilers, pr.CompilerType_NANOPB)
		fmt.Printf("NANOPB compile...\n")
	}
	if len(creq.Compilers) == 0 {
		return nil, fmt.Errorf("unable to determine which compiler to use for this repo")
	}
	cr, err := pr.GetProtoRendererClient().CompileFile(ctx, creq)
	if err != nil {
		return nil, err
	}
	pcr := &ProtoCompileResult{}
	for _, cf := range cr.Files {
		var fs []*File

		if cf.Compiler == pr.CompilerType_NANOPB {
			sname := "proto_files/nanopb/" + filepath.Base(cf.Filename)
			fs = append(fs, &File{Filename: sname, Content: cf.Content})
		}
		pcr.Files = append(pcr.Files, fs...)
		//		fmt.Printf("Compiled file (%v): %s\n", cf.Compiler, cf.Filename)
	}
	return pcr, nil
}




