// client create: RepoBuilderClient
/*
  Created by /home/cnw/devel/go/yatools/src/golang.yacloud.eu/yatools/protoc-gen-cnw/protoc-gen-cnw.go
*/

/* geninfo:
   filename  : protos/golang.conradwood.net/apis/repobuilder/repobuilder.proto
   gopackage : golang.conradwood.net/apis/repobuilder
   importname: ai_0
   clientfunc: GetRepoBuilder
   serverfunc: NewRepoBuilder
   lookupfunc: RepoBuilderLookupID
   varname   : client_RepoBuilderClient_0
   clientname: RepoBuilderClient
   servername: RepoBuilderServer
   gsvcname  : repobuilder.RepoBuilder
   lockname  : lock_RepoBuilderClient_0
   activename: active_RepoBuilderClient_0
*/

package repobuilder

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_RepoBuilderClient_0 sync.Mutex
  client_RepoBuilderClient_0 RepoBuilderClient
)

func GetRepoBuilderClient() RepoBuilderClient { 
    if client_RepoBuilderClient_0 != nil {
        return client_RepoBuilderClient_0
    }

    lock_RepoBuilderClient_0.Lock() 
    if client_RepoBuilderClient_0 != nil {
       lock_RepoBuilderClient_0.Unlock()
       return client_RepoBuilderClient_0
    }

    client_RepoBuilderClient_0 = NewRepoBuilderClient(client.Connect(RepoBuilderLookupID()))
    lock_RepoBuilderClient_0.Unlock()
    return client_RepoBuilderClient_0
}

func RepoBuilderLookupID() string { return "repobuilder.RepoBuilder" } // returns the ID suitable for lookup in the registry. treat as opaque, subject to change.

func init() {
   client.RegisterDependency("repobuilder.RepoBuilder")
   AddService("repobuilder.RepoBuilder")
}




