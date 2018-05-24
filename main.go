package main

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/docker/archive"
	"github.com/containers/image/docker/reference"
	"github.com/containers/image/signature"
	"github.com/containers/image/types"
)

func main() {
	fmt.Println("Starting http file sever")
	http.HandleFunc("/", HandleClient)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func HandleClient(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	image := request.URL.Query().Get("image")
	if image == "" {
		//Get not set, send a 400 bad request
		http.Error(writer, "Get 'image' not specified in url.", http.StatusPreconditionFailed)
		return
	}

	ref, err := docker.ParseReference(image)
	if err != nil {
		http.Error(writer, "Fail to parse 'image'", http.StatusNotFound)
		return
	}

	srcimg, err := ref.NewImageSource(ctx, nil)
	if err != nil {
		http.Error(writer, "Fail to parse 'image'", 400)
		return
	}
	defer srcimg.Close()
	fmt.Println("Client requests: " + image)

	refTagged, ok := ref.DockerReference().(reference.NamedTagged)
	if !ok {
		http.Error(writer, "Fail to get ref", 400)
		return
	}

	gzw := gzip.NewWriter(writer)
	defer gzw.Close()

	dest, err := archive.NewImageDestinationWriter(nil, refTagged, gzw)
	if err != nil {
		http.Error(writer, "Fail NewImageDestinationWriter", 400)
		return
	}
	defer dest.Close()

	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.tgz\"", path.Base(image)))
	writer.Header().Set("Content-Type", "application/gzip")

	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}

	sctx, err := signature.NewPolicyContext(policy)
	if err != nil {
		http.Error(writer, "Fail NewPolicyContext", 400)
		return
	}

	reportWriter := os.Stdout
	if err := copy.ImageR(ctx, sctx, dest, srcimg, &copy.Options{
		RemoveSignatures: false,
		SignBy:           "",
		ReportWriter:     reportWriter,
		SourceCtx:        nil, //s.systemContext,
		DestinationCtx:   &types.SystemContext{},
	}); err != nil {
		http.Error(writer, fmt.Sprintf("Fail to get source information, err: %s", err.Error()), 400)
		return
	}

	return
}
