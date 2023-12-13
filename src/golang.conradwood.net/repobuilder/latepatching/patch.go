package latepatching

import (
	"bytes"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	pb "golang.conradwood.net/apis/repobuilder"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"golang.conradwood.net/repobuilder/gitpatch"
	"golang.conradwood.net/repobuilder/protos"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type TemplateData struct {
	FQDN_GO_PROTOPACKAGE string
	GO_PROTOPACKAGE      string
	JAVA_PROTOPACKAGE    string
}

func valid_proto_go(in string) string {
	if len(in) < 2 {
		return in + "_" + utils.RandomString(20)
	}
	fc := in[0]
	if (fc >= 'a' && fc <= 'z') || (fc >= 'A' && fc <= 'Z') {
		return in
	}
	return "x" + in
}
func patch(lpq *pb.LatePatchingQueue) error {
	ctx := authremote.Context()
	repo, err := gitpb.GetGIT2Client().RepoByID(ctx, &gitpb.ByIDRequest{ID: lpq.RepositoryID})
	if err != nil {
		return err
	}
	fmt.Printf("Patching Repo #%d (%s)\n", repo.ID, repo.ArtefactName)
	gr, err := gitpatch.GetRepoReferenceByID(ctx, repo.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Checked out into %s\n", gr.GitDirAbsFilename())
	defer gr.Close()
	pn := valid_proto_go(repo.ArtefactName)
	protofilename := fmt.Sprintf("protos/userprotos.singingcat.net/%s/%s/%s.proto", repo.CreateUser, pn, pn)
	td := &TemplateData{
		FQDN_GO_PROTOPACKAGE: to_fqdn_go_proto_package_name(protofilename),
		GO_PROTOPACKAGE:      to_proto_package_name(repo.ArtefactName),
		JAVA_PROTOPACKAGE:    "repobuilder.latepatching does not yet support java packages",
	}
	content, err := create_patch_file("protofile.template", td)
	if err != nil {
		return err
	}
	err = gr.AddFile(protofilename, []byte(content))
	if err != nil {
		return err
	}
	err = compileProtos(gr, repo.ID, protofilename)
	if err != nil {
		return err
	}
	if gr.NeedsCommit() {
		err = gr.CommitAndPush()
		if err != nil {
			return err
		}
	}
	ctx = authremote.Context()
	urs := &gitpb.UpdateRepoStatusRequest{
		RepoID:   repo.ID,
		ReadOnly: gitpb.NewRepoState_SET_FALSE,
	}
	_, err = gitpb.GetGIT2Client().UpdateRepoStatus(ctx, urs)
	if err != nil {
		return err
	}
	return nil
}

func create_patch_file(template_filename string, td *TemplateData) (string, error) {
	tct, err := utils.ReadFile("configs/" + template_filename)
	if err != nil {
		return "", err
	}
	t, err := template.New(template_filename).Parse(string(tct))
	if err != nil {
		return "", err
	}
	b := &bytes.Buffer{}
	err = t.Execute(b, td)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func compileProtos(gr *gitpatch.GitReference, repoid uint64, filename string) error {
	dir := gr.GitDirAbsFilename()
	ctx := authremote.Context()
	pcr, err := protos.Compile(ctx, repoid, dir, filename)
	if err != nil {
		return err
	}
	fmt.Printf("res: %#v\n", pcr)
	for _, file := range pcr.Files {
		err = gr.AddFile(file.Filename, file.Content)
		if err != nil {
			return err
		}
		fmt.Printf("File: %#v\n", file.Filename)
	}
	return nil
}
func writeFile(filename string, content []byte) error {
	dir := filepath.Dir(filename)
	os.MkdirAll(dir, 0777)
	return utils.WriteFile(filename, content)
}

func to_proto_package_name(pkg_name string) string {
	res := strings.ToLower(pkg_name)
	res = strings.ReplaceAll(res, "-", "_")
	return res
}

// wants a path, like "protos/golang.yacloud.eu/apis/foo/foo.proto"
func to_fqdn_go_proto_package_name(pkg_name string) string {
	res := strings.ToLower(pkg_name)
	res = strings.ReplaceAll(res, "-", "_")
	res = filepath.Dir(res)
	res = strings.TrimPrefix(res, "protos/")
	return res
}




