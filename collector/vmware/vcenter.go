package vmware

import (
	"net/url"

	"github.com/brunotm/tact"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type vcenter struct {
	client *govmomi.Client
	mors   []types.ManagedObjectReference
}

func newClient(sess *tact.Session) (vc *vcenter, err error) {
	u, err := url.Parse(sess.Node().APIURL)
	if err != nil {
		return nil, err
	}
	u.User = url.UserPassword(sess.Node().APIUser, sess.Node().APIPassword)

	vc = &vcenter{}
	vc.client, err = govmomi.NewClient(sess.Context(), u, true)
	if err != nil {
		return nil, err
	}

	var perfmanager mo.PerformanceManager
	err = vc.client.RetrieveOne(sess.Context(), *vc.client.ServiceContent.PerfManager, nil, &perfmanager)
	if err != nil {
		return nil, err
	}
	return vc, nil
}
