package main

import (
	"fmt"
	"net/http"
	"os"

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
	//First of check if Get is set in the URL

	ctx := request.Context()

	image := request.URL.Query().Get("image")
	if image == "" {
		//Get not set, send a 400 bad request
		http.Error(writer, "Get 'image' not specified in url.", 400)
		return
	}

	ref, err := docker.ParseReference(image)
	if err != nil {
		http.Error(writer, "Fail to parse 'image'", 400)
		return
	}

	//img, err := ref.NewImage(ctx, nil)
	srcimg, err := ref.NewImageSource(ctx, nil)
	if err != nil {
		http.Error(writer, "Fail to parse 'image'", 400)
		return
	}
	defer srcimg.Close()
	fmt.Println("Client requests: " + image)

	//refdest := reference.NamedTagged{}
	//refTagged, err := reference.WithTag(ref.DockerReference(), "latest")
	refTagged, ok := ref.DockerReference().(reference.NamedTagged)
	if !ok {
		http.Error(writer, "Fail to get ref", 400)
		return
	}
	dest, err := archive.NewImageDestinationWriter(nil, refTagged, writer)
	if err != nil {
		http.Error(writer, "Fail NewImageDestinationWriter", 400)
		return
	}

	writer.Header().Set("Content-Disposition", "attachment; filename="+image+".tar")

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
		http.Error(writer, "Fail to get source information, err", 400)
		return
	}

	//Send the headers
	//writer.Header().Set("Content-Type", FileContentType)

	// blobs := img.LayerInfos()
	// img.ConfigInfo()
	// if err != nil || len(blobs) == 0 {
	// 	http.Error(writer, "Fail to get source information", 400)
	// 	return
	// }

	// manifest, _, err := srcimg.GetManifest(ctx, nil)
	// dest.PutManifest(ctx, manifest)
	// for _, blob := range blobs {
	// 	reader, _, err := srcimg.GetBlob(ctx, blob)
	// 	if err != nil {
	// 		http.Error(writer, "Fail to get source data", 400)
	// 		return
	// 	}
	// 	defer reader.Close()
	//
	// 	//writer.Header().Set("Content-Length", size)
	// 	dest.PutBlob(ctx, reader, blob, false)
	// }
	return
}
