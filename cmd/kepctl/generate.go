/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/spf13/cobra"
	"k8s.io/enhancements/pkg/kepctl"
)

func buildReleaseCommand(k *kepctl.Client) *cobra.Command {
	baseCmd := &cobra.Command{
		Use:   "generate",
		Short: "Manage enhancements for a release",
	}
	baseCmd.AddCommand(buildReleaseInitCommand(k))
	baseCmd.AddCommand(buildGenerateReleaseCommand(k))
	return baseCmd
}

// maybe this should be kepctl init release v1.20 or should we extract to a new tool
func buildReleaseInitCommand(k *kepctl.Client) *cobra.Command {
	opts := &kepctl.InitReleaseOpts{}
	cmd := &cobra.Command{
		Use:     "release [number]",
		Short:   "Initalize a new release",
		Long:    "Initalize a new release and create required folder structure",
		Example: `  kepctl generaet release v1.21 --enhancement-team <people> --release-leads <names>`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.InitRelease(opts)
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&opts.EnhancementTeam, "enhancement-team", []string{}, "Enhancement Team Members")
	f.StringSliceVar(&opts.ReleaseLeads, "release-leads", []string{}, "Release Leads")
	addRepoPathFlag(f, &opts.CommonArgs)
	return cmd
}

// maybe this should be kepctl generate release v1.20?
func buildGenerateReleaseCommand(k *kepctl.Client) *cobra.Command {
	opts := &kepctl.ReleaseManifestOpts{}
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Generate the release manifest",
		Example: `	kepctl generate manifest v1.21`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.GenerateRelease(opts)
		},
	}
	f := cmd.Flags()
	addRepoPathFlag(f, &opts.CommonArgs)
	return cmd
}
