package main

import (
	"context"
	"fmt"
	oa "golang.conradwood.net/apis/objectauth"
	pb "golang.conradwood.net/apis/repobuilder"
	"golang.conradwood.net/go-easyops/auth"
)

func (c *Creator) CreatePermissions() error {
	err := c.RestoreContext()
	if err != nil {
		return err
	}
	u := auth.GetUser(c.ctx)
	if u == nil {
		return fmt.Errorf("Missing user account")
	}
	c.Printf("Git RepositoryID: %d\n", c.tgr.RepositoryID)
	for _, gid := range c.req.VisibilityGroupIDs {
		c.Printf("   Visibility: %s\n", gid)
	}
	for _, gid := range c.req.AccessGroupIDs {
		c.Printf("   Access    : %s\n", gid)
	}
	ob := oa.GetObjectAuthClient()
	altid := ""
	if u.ID == "7" {
		altid = "1"
	}
	if u.ID == "1" {
		altid = "7"
	}
	// first grant the creator explicit rights to his new repo
	_, err = ob.GrantToUser(c.ctx, &oa.GrantUserRequest{
		ObjectType: oa.OBJECTTYPE_GitRepository,
		ObjectID:   c.tgr.RepositoryID,
		UserID:     u.ID,
		Read:       true,
		Write:      true,
		Execute:    true,
		View:       true,
	})
	if err != nil {
		return err
	}
	if altid != "" {
		ob := oa.GetObjectAuthClient()
		// first grant the creator explicit rights to his new repo
		_, err = ob.GrantToUser(c.ctx, &oa.GrantUserRequest{
			ObjectType: oa.OBJECTTYPE_GitRepository,
			ObjectID:   c.tgr.RepositoryID,
			UserID:     altid,
			Read:       true,
			Write:      true,
			Execute:    true,
			View:       true,
		})

		if err != nil {
			return err
		}
	}
	err = SetAdditionalPermissions(c.ctx, c.req, c.tgr.RepositoryID)
	if err != nil {
		return err
	}

	return nil
}

func SetAdditionalPermissions(ctx context.Context, req *pb.CreateWebRepoRequest, repoid uint64) error {
	var err error
	ob := oa.GetObjectAuthClient()
	for _, gid := range req.DeveloperGroupIDs {
		fmt.Printf("   Dev-Access: %s\n", gid)
		_, err = ob.GrantToGroup(ctx, &oa.GrantGroupRequest{
			ObjectType: oa.OBJECTTYPE_GitRepository,
			ObjectID:   repoid,
			GroupID:    gid,
			Read:       true,
			Write:      true,
			Execute:    true,
			View:       true,
		})
		if err != nil {
			return err
		}
	}

	for _, gid := range req.VisibilityGroupIDs {
		fmt.Printf("   Vis-Access: %s\n", gid)
		_, err = ob.GrantToGroup(ctx, &oa.GrantGroupRequest{
			ObjectType: oa.OBJECTTYPE_GitRepository,
			ObjectID:   repoid,
			GroupID:    gid,
			Read:       true,
			Write:      false,
			Execute:    false,
			View:       true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}






