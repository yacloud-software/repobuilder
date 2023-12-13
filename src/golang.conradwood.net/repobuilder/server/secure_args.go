package main

import (
	apb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/secureargs"
	"golang.conradwood.net/go-easyops/authremote"
)

func (c *Creator) CreateServiceAccount() error {
	if c.tgr.ServiceUserID != "" {
		return nil
	}
	ctx := authremote.Context()
	// create the serviceaccount
	ns, err := authremote.GetAuthManagerClient().CreateService(ctx, &apb.CreateServiceRequest{
		ServiceName: c.ServiceName(),
	})
	if err != nil {
		return err
	}

	c.tgr.ServiceUserID = ns.User.ID
	c.tgr.ServiceToken = ns.Token
	err = c.SaveProgress() // save serviceid etc
	if err != nil {
		return err
	}
	return nil
}

func (c *Creator) CreateSecureArgs() error {
	ctx := authremote.Context()
	_, err := secureargs.GetSecureArgsClient().SetArg(ctx, &secureargs.SetArgRequest{
		RepositoryID: c.tgr.RepositoryID,
		ArtefactID:   c.tgr.ArtefactID,
		Name:         "TOKEN",
		Value:        c.tgr.ServiceToken,
	})
	if err != nil {
		return err
	}
	return nil
}



