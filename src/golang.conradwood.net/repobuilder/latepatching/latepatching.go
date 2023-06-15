package latepatching

import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"golang.conradwood.net/repobuilder/db"
	"sort"
	"time"
)

var (
	trigger_chan = make(chan *trigger, 1)
	debug        = flag.Bool("debug_late_patching", false, "debug late patching code")
)

type trigger struct {
}

func init() {
	go late_patch_loop()
}
func Trigger() {
	t := &trigger{}
	select {
	case trigger_chan <- t:
	//
	default:
		// do not block
	}
}
func late_patch_loop() {
	wait := 3
	for {
		select {
		case <-trigger_chan:
		// triggered
		case <-time.After(time.Duration(wait) * time.Second):
			// timeout
		}
		handle_late_patchers()
		wait = 30
	}
}

func handle_late_patchers() {
	if *debug {
		fmt.Printf("Late patching...\n")
	}
	ctx := authremote.Context()
	lpqs, err := db.DefaultDBLatePatchingQueue().All(ctx)
	if err != nil {
		fmt.Printf("Late patching failed: %s\n", err)
		return
	}
	if *debug {
		fmt.Printf("Late patching %d entries...\n", len(lpqs))
	}
	sort.Slice(lpqs, func(i, j int) bool {
		return lpqs[i].LastAttempt < lpqs[j].LastAttempt
	})
	for _, lpq := range lpqs {
		err = patch(lpq)
		ctx = authremote.Context()
		if err != nil {
			fmt.Printf("Patching failed: %s\n", utils.ErrorString(err))
			lpq.LastAttempt = uint32(time.Now().Unix())
			db.DefaultDBLatePatchingQueue().Update(ctx, lpq)
			continue
		}
		fmt.Printf("Successfully patched repository #%d\n", lpq.RepositoryID)
		db.DefaultDBLatePatchingQueue().DeleteByID(ctx, lpq.ID)
	}
}
