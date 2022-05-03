package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	googleworkspace "github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	opts := tfsdk.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		Name: "registry.terraform.io/hashicorp/googleworkspace",
	}

	err := tfsdk.Serve(context.Background(), googleworkspace.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
