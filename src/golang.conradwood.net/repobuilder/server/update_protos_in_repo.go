package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/protorenderer"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
	"sync"
)

type process_package_state struct {
	err       error
	recreated map[string]bool // prefices that were deleted/created
	rec       sync.Mutex
	vendor    string
}

func (pps *process_package_state) SetError(pp *process_package, err error) {
	if err == nil {
		return
	}
	pps.err = err
}

func (pps *process_package_state) PrefixToDir(prefix string) string {
	return pps.vendor + "/" + prefix
}

// delete/create prefix in vendor UNLESS it has been done alredy
func (pps *process_package_state) HandlePrefix(prefix string) error {
	if len(prefix) < 2 {
		panic(fmt.Sprintf("invalid prefix: \"%s\"", prefix))
	}
	if pps.recreated[prefix] {
		return nil
	}
	pps.rec.Lock()
	defer pps.rec.Unlock()
	if pps.recreated[prefix] {
		return nil
	}
	dir := pps.PrefixToDir(prefix)
	if utils.FileExists(dir) {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("Failed to remove dir \"%s\" for prefix \"%s\"\n", dir, prefix)
			return err
		}
	}
	os.MkdirAll(dir, 0777)
	pps.recreated[prefix] = true
	return nil
}

// save a .pb.go file
func (pps *process_package_state) Save(fp *protorenderer.FlatPackage, filename string, file *protorenderer.File) error {
	err := pps.HandlePrefix(fp.Prefix)
	if err != nil {
		return err
	}
	fn := pps.vendor + "/" + filename
	//	fmt.Printf("Saving file \"%s\" (prefix \"%s\")\n", fn, fp.Prefix)
	dir := filepath.Dir(fn)
	os.MkdirAll(dir, 0777)
	err = utils.WriteFile(fn, []byte(file.Content))
	if err != nil {
		return err
	}
	return nil
}

type process_package struct {
	state *process_package_state
	p     *protorenderer.FlatPackage
	ctx   context.Context
}

// process a flatpackage
func (pp *process_package) Process() {
	id := &protorenderer.ID{ID: pp.p.ID}
	fl, err := protorenderer.GetProtoRendererClient().GetFilesGO(pp.ctx, id)
	if err != nil {
		pp.state.SetError(pp, err)
		return
	}
	for _, fn := range fl.Files {
		fr := &protorenderer.FileRequest{PackageID: id, Filename: fn}
		file, err := protorenderer.GetProtoRendererClient().GetFile(pp.ctx, fr)
		if err != nil {
			pp.state.SetError(pp, err)
			return
		}
		fmt.Printf("Saving Package ID #%s (name=%s, prefix=%s)\n", id, pp.p.Name, pp.p.Prefix)
		err = pp.state.Save(pp.p, fn, file)
		if err != nil {
			pp.state.SetError(pp, err)
			return
		}
	}
}
