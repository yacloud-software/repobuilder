package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
	use_vendor = flag.Bool("use_vendor", false, "if true update vendor directories and install protos and stuff into it")
)

func (c *Creator) UseVendor() bool {
	return *use_vendor
}

// modify the template to match new parameters
func (c *Creator) ModifyTemplate() error {
	rerr := c.CloneRepo()
	if rerr != nil {
		return rerr
	}

	dir := c.GitDir()
	c.moveFiles()
	files, err := findTemplatableFiles(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		err = c.modifyFile(f)
		if err != nil {
			c.Printf("Failed to modify file  \"%s\": %s\n", f, err)
			return err
		}
	}
	c.needscommit = true
	return nil
}
func (c *Creator) GoDomainDir() string {
	s := "golang." + c.req.Domain
	return strings.ToLower(s)
}

// e.g. /fulldir/src/golang.conradwood.net/repobuilder
func (c *Creator) GoPackageDir() string {
	s := c.GitDir() + "/src/golang." + c.req.Domain + "/" + c.req.Name
	return strings.ToLower(s)
}
func (c *Creator) RepositoryIDString() string {
	return fmt.Sprintf("%d", c.tgr.RepositoryID)
}
func (c *Creator) ServiceName() string {
	res := c.req.ServiceName
	a := res[:1]
	b := res[1:]
	res = strings.ToUpper(a) + b
	return res
}
func (c *Creator) ClientName() string {
	res := c.ServiceName() + "Client"
	return res
}

// the path used by .go files to import the pb.go file, e.g. "golang.conradwood.net/apis/common"
func (c *Creator) GoProtoImportPath() string {
	res := c.ProtoPackagePrefix() + "/" + c.ProtoPackageName()
	return strings.ToLower(res)
}
func (c *Creator) ProtoPackageName() string {
	res := c.req.Name
	return strings.ToLower(res)
}
func (c *Creator) ProtoJavaPackageName() string {
	s := strings.Join(reverseDomain(c.req.Domain), ".")
	s = s + ".apis." + c.req.Name
	return strings.ToLower(s)
}
func (c *Creator) shortName() string {
	return strings.ToLower(c.req.Name)
}

// e.g. yacloud.eu/apis or golang.conradwood.net/apis
func (c *Creator) ProtoPackagePrefix() string {
	s := c.req.ProtoDomain + "/apis"
	s = strings.ToLower(s)
	return s
}

// e.g. /fulldir/src/golang.conradwood.net/repobuilder
func (c *Creator) GoVendorDir() string {
	s := c.GitDir() + "/src/golang." + c.req.Domain + "/vendor"
	return strings.ToLower(s)
}

func (c *Creator) gitadd() error {
	l := linux.New()
	l.SetMaxRuntime(time.Duration(300) * time.Second)
	dir := c.GitDir()
	os.MkdirAll(dir, 0777)
	url := c.GitCloneURL()
	c.GitSetAuth(url)
	out, err := l.SafelyExecuteWithDir([]string{"git", "add", "."}, dir, nil)
	if err != nil {
		c.Printf("Git add Error. Git said: %s\n", out)
		return err
	}
	return nil
}

// move template files around
func (c *Creator) moveFiles() error {
	s := c.GoPackageDir()
	c.Printf("Creating go files from template in \"%s\"...\n", s)
	os.MkdirAll(s, 0777)

	templdir := c.GitDir() + "/src/golang.conradwood.net/template"
	// rename the client/server .go
	err := os.Rename(templdir+"/client/template-client.go", templdir+"/client/"+c.shortName()+"-client.go")
	if err != nil {
		return err
	}
	err = os.Rename(templdir+"/server/template-server.go", templdir+"/server/"+c.shortName()+"-server.go")
	if err != nil {
		return err
	}
	// copy the template dir
	err = linux.CopyDir(templdir, s)
	if err != nil {
		return err
	}
	err = os.RemoveAll(templdir)
	if err != nil {
		return err
	}
	// copy the proto
	np := "/protos/" + c.ProtoPackagePrefix() + "/" + c.shortName()
	np = strings.ToLower(np)
	os.MkdirAll(c.GitDir()+np, 0777)
	np = np + "/" + c.shortName() + ".proto"
	c.tgr.ProtoFilename = np
	err = os.Rename(c.GitDir()+"/skel.proto", c.GitDir()+np)
	if err != nil {
		return err
	}

	// move vendor
	ovendor := c.GitDir() + "/src/golang.conradwood.net/vendor"
	nvendor := c.GoVendorDir()
	if ovendor != nvendor {
		err = os.Rename(ovendor, nvendor)
		if err != nil {
			return err
		}
	}

	// put the proto.skel to the right place

	// these _may_ be empty, so attempt to delete them
	// they may not be empty if we chose a domain that matches either of these
	os.Remove(c.GitDir() + "/src/golang.conradwood.net") // remove if empty only
	os.Remove(c.GitDir() + "/src/golang.singingcat.net") // remove if empty only
	return nil
}

type replace struct {
	orig string
	repl string
}

func (c *Creator) modifyFile(absfile string) error {
	c.Printf("Modifing \"%s\"...\n", absfile)
	replace := []*replace{
		&replace{orig: "skel-go", repl: c.req.RepoName},
		&replace{orig: "src/golang.conradwood.net", repl: "src/" + c.GoDomainDir()},
		&replace{orig: "template-client.go", repl: c.shortName() + "-client.go"},
		&replace{orig: "template-server.go", repl: c.shortName() + "-server.go"},
		&replace{orig: "template-server", repl: c.shortName() + "-server"},
		&replace{orig: "PROTOPACKAGE", repl: c.ProtoPackageName()},
		&replace{orig: "PROTOIMPORTPATH", repl: c.GoProtoImportPath()},
		&replace{orig: "SERVICENAME", repl: c.ServiceName()},
		&replace{orig: "GODOMAIN", repl: c.GoDomainDir()},
		&replace{orig: "golang.conradwood.net/template/appinfo", repl: c.GoDomainDir() + "/" + c.req.Name + "/appinfo"},
		&replace{orig: "golang.conradwood.net/template", repl: c.GoDomainDir() + "/" + c.req.Name},
		&replace{orig: "JAVAPACKAGE", repl: c.ProtoJavaPackageName()},
		&replace{orig: "REPOSITORYID", repl: c.RepositoryIDString()},
		&replace{orig: `pb "golang.conradwood.net/apis/echoservice"`, repl: "pb \"" + c.GoProtoImportPath() + "\""},
		&replace{orig: "pb.EchoServiceClient", repl: "pb." + c.ClientName()},
		&replace{orig: "pb.GetEchoServiceClient", repl: "pb.Get" + c.ClientName()},
		&replace{orig: "pb.RegisterEchoServiceServer", repl: "pb.Register" + c.ServiceName() + "Server"},
		&replace{orig: "Starting EchoServiceServer", repl: "Starting " + c.ServiceName() + "Server"},
	}

	ct, err := utils.ReadFile(absfile)
	if err != nil {
		return err
	}
	nfct := string(ct)
	for _, r := range replace {
		nfct = strings.ReplaceAll(nfct, r.orig, r.repl)
	}
	err = utils.WriteFile(absfile, []byte(nfct))
	if err != nil {
		return err
	}
	err = c.gitadd()
	if err != nil {
		return err
	}
	return nil
}

// returns absolute filenames
func findTemplatableFiles(dir string) ([]string, error) {
	res, err := findInDirWithSuffix(dir+"/src", ".go")
	if err != nil {
		return nil, err
	}

	fs, err := findInDirWithSuffix(dir+"/src", ".java")
	if err != nil {
		return nil, err
	}
	res = append(res, fs...)

	fs, err = findInDirWithSuffix(dir+"/src", "Makefile")
	if err != nil {
		return nil, err
	}
	res = append(res, fs...)

	fs, err = findInDirWithSuffix(dir+"/src", "go.mod")
	if err != nil {
		return nil, err
	}
	res = append(res, fs...)

	fs, err = findInDirWithSuffix(dir+"/protos", ".proto")
	if err != nil {
		return nil, err
	}
	res = append(res, fs...)

	res = append(res, dir+"/deployment/deploy.yaml")
	res = append(res, dir+"/autobuild.sh")
	res = append(res, dir+"/update-dbs.sh")

	return res, nil
}
func findInDirWithSuffix(dir string, suffix string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, f := range files {
		fn := f.Name()
		if f.IsDir() {
			if fn == "vendor" {
				continue
			}
			fs, err := findInDirWithSuffix(dir+"/"+fn, suffix)
			if err != nil {
				return nil, err
			}
			res = append(res, fs...)
			continue
		}
		if !strings.HasSuffix(fn, suffix) {
			continue
		}
		res = append(res, dir+"/"+fn)
	}
	return res, nil
}

func reverseDomain(domain string) []string {
	s := strings.Split(domain, ".")
	res := make([]string, len(s))
	for i, _ := range s {
		res[len(s)-i-1] = s[i]
	}
	return res
}

