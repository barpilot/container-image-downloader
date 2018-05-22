package main

import (
	"fmt"
	"net/http"

	"github.com/containers/image/docker"
	"github.com/containers/image/docker/reference"
	"github.com/containers/image/docker/tarfile"
)

func main() {
	fmt.Println("Starting http file sever")
	http.HandleFunc("/", HandleClient)
	err := http.ListenAndServe(":80", nil)
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

	img, err := ref.NewImage(ctx, nil)
	srcimg, err := ref.NewImageSource(ctx, nil)
	if err != nil {
		http.Error(writer, "Fail to parse 'image'", 400)
		return
	}
	defer srcimg.Close()
	fmt.Println("Client requests: " + image)

	//refdest := reference.NamedTagged{}
	refTagged, err := reference.WithTag(ref.DockerReference(), "latest")
	dest := tarfile.NewDestination(writer, refTagged)

	//Send the headers
	//writer.Header().Set("Content-Type", FileContentType)

	blobs := img.LayerInfos()
	if err != nil || len(blobs) == 0 {
		http.Error(writer, "Fail to get source information", 400)
		return
	}

	reader, _, err := srcimg.GetBlob(ctx, blobs[0])
	if err != nil {
		http.Error(writer, "Fail to get source data", 400)
		return
	}

	writer.Header().Set("Content-Disposition", "attachment; filename="+image+".tar")
	//writer.Header().Set("Content-Length", size)
	dest.PutBlob(ctx, reader, blobs[0], false)
	return
}
