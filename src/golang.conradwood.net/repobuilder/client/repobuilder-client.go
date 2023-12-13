package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/repobuilder"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"os"
)

var (
	description = flag.String("description", "", "new repo description")
	name        = flag.String("name", "", "new repo name")
	domain      = flag.String("domain", "", "new repo domain name")
	reponame    = flag.String("reponame", "", "new repo name")
	servicename = flag.String("service", "", "new repo service name")
	trigger     = flag.Bool("retrigger", false, "trigger processing again")
	probe       = flag.Bool("probe", false, "will run through the 'prober' recreation of a repo")
	echoClient  pb.RepoBuilderClient
)

func main() {
	flag.Parse()

	echoClient = pb.GetRepoBuilderClient()

	// a context with authentication
	ctx := authremote.Context()
	if *probe {
		_, err := echoClient.RecreateWebRepo(ctx, &pb.CreateWebRepoRequest{})
		utils.Bail("failed to probe", err)
		fmt.Printf("Done\n")
		os.Exit(0)
	}
	if *trigger {
		_, err := echoClient.RetriggerAll(ctx, &common.Void{})
		utils.Bail("failed to trigger", err)
		fmt.Printf("Done\n")
		os.Exit(0)
	}
	empty := &pb.CreateWebRepoRequest{
		Description:        *description,
		Name:               *name,
		Language:           1,
		Domain:             *domain,
		RepoName:           *reponame,
		ServiceName:        *servicename,
		VisibilityGroupIDs: visgroups(),
		AccessGroupIDs:     accessgroups(),
		DeveloperGroupIDs:  devgroups(),
	}
	response, err := echoClient.CreateWebRepo(ctx, empty)
	utils.Bail("Failed to ping server", err)
	fmt.Printf("RequestID: %d\n", response.RequestID)

	fmt.Printf("Done.\n")
	os.Exit(0)
}
func visgroups() []string {
	return nil
}

func accessgroups() []string {
	return nil
}
func devgroups() []string {
	return nil
}



