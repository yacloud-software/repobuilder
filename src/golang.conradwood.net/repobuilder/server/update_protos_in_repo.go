package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/common"
	pr "golang.conradwood.net/apis/protorenderer"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	upd_protoclient pr.ProtoRendererServiceClient
)

// updated .pb.go files in repo from protorenderer server
func UpdateProtosInRepo(ctx context.Context, repopath string) error {
	if !*use_vendor {
		return nil
	}
	if repopath[0] != '/' {
		return fmt.Errorf("UpdateProtosInRepo: requires absolute path (not %s)", repopath)
	}
	if upd_protoclient == nil {
		upd_protoclient = pr.GetProtoRendererClient()
	}
	fmt.Printf("Updating protos in git repository \"%s\"...\n", repopath)
	vendor, err := findVendor(repopath)
	if err != nil {
		return err
	}
	fmt.Printf("Vendor directory: \"%s\"\n", vendor)
	fls, err := upd_protoclient.GetPackages(ctx, &common.Void{})
	if err != nil {
		fmt.Printf("Failed to get proto packages: %s\n", err)
		return err
	}

	ppr := &process_package_state{
		recreated: make(map[string]bool),
		vendor:    vendor,
	}
	wg := &sync.WaitGroup{}
	for _, p := range fls.Packages {
		wg.Add(1)
		nctx := authremote.Context()
		pp := &process_package{state: ppr, ctx: nctx, p: p}
		go func(ppl *process_package) {
			ppl.Process()
			wg.Done()
		}(pp)
	}
	wg.Wait()
	if ppr.err != nil {
		fmt.Printf("Failed to update protos: %s\n", ppr.err)
		return ppr.err
	}
	fmt.Printf("git-adding .pb.go proto files to git repo\n")
	// add all pb.go to repo...
	cmd := []string{"git", "add"}
	for k, _ := range ppr.recreated {
		fmt.Printf("Git add: \"%s\"\n", k)
		l := linux.New()
		nc := append(cmd, k)
		o, err := l.SafelyExecuteWithDir(nc, vendor, nil)
		if err != nil {
			fmt.Printf("Failed to add pb.go files: \n%s\n %s\n", o, err)
			return err
		}
	}
	//return fmt.Errorf("UpdateProtosInRepo: Not implemented")
	return nil
}

// find vendor dir in repo
func findVendor(repopath string) (string, error) {
	files, err := ioutil.ReadDir(repopath + "/src")
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("UpdateProtosInRepo: no files found under \"src\"...")
	}
	for _, f := range files {
		v := repopath + "/src/" + f.Name() + "/vendor"
		if utils.FileExists(v) {
			return v, nil
		}
	}
	return "", fmt.Errorf("UpdateProtosInRepo: no vendor dir found under \"src\"...")

}

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
func (pps *process_package_state) Save(fp *pr.FlatPackage, filename string, file *pr.File) error {
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
	p     *pr.FlatPackage
	ctx   context.Context
}

// process a flatpackage
func (pp *process_package) Process() {
	id := &pr.ID{ID: pp.p.ID}
	fl, err := upd_protoclient.GetFilesGO(pp.ctx, id)
	if err != nil {
		pp.state.SetError(pp, err)
		return
	}
	for _, fn := range fl.Files {
		fr := &pr.FileRequest{PackageID: id, Filename: fn}
		file, err := upd_protoclient.GetFile(pp.ctx, fr)
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



