package main

import (
	"context"
	"fmt"
	"os"

	"github.com/quorumcontrol/decentragit-remote/runner"
	"github.com/quorumcontrol/decentragit-remote/storage/readonly"
	"github.com/quorumcontrol/decentragit-remote/storage/split"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

func storer() storage.Storer {
	gitStore := filesystem.NewStorage(osfs.New(os.Getenv("GIT_DIR")), cache.NewObjectLRUDefault())
	readonlyStore := readonly.NewStorage(gitStore)

	// git-remote-helper expects this script to write git objects, but nothing else
	// therefore initialize a go-git storage with the ability to write objects & shallow
	// but make reference, index, and config read only ops
	return split.NewStorage(&split.StorageMap{
		ObjectStorage:    gitStore,
		ShallowStorage:   gitStore,
		ReferenceStorage: readonlyStore,
		IndexStorage:     readonlyStore,
		ConfigStorage:    readonlyStore,
	})
}

func main() {
	fmt.Fprintf(os.Stderr, "decentragit loaded for %s\n", os.Getenv("GIT_DIR"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	local, err := git.Open(storer(), nil)
	if err != nil {
		panic(err)
	}

	r, err := runner.New(local)
	if err != nil {
		panic(err)
	}

	if err := r.Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
