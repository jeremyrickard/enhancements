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

func buildCreateCommand(k *kepctl.Client) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource [kep, reciept]",
		Long:  "Create a new enhancements resource, either a kep or a reciept for a given release.",
	}
	cmd.AddCommand(buildCreateKEPCommand(k))
	cmd.AddCommand(buildCreateReceiptCommand(k))
	return cmd
}

func buildCreateKEPCommand(k *kepctl.Client) *cobra.Command {
	opts := kepctl.CreateOpts{}
	cmd := &cobra.Command{
		Use:     "kep [KEP]",
		Short:   "Create a new KEP",
		Long:    "Create a new KEP using the current KEP template for the given type",
		Example: `  kepctl create kep sig-architecture/000-mykep`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.Create(opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.Title, "title", "", "KEP Title")
	f.StringArrayVar(&opts.Authors, "authors", []string{}, "Authors")
	f.StringArrayVar(&opts.Reviewers, "reviewers", []string{}, "Reviewers")
	f.StringVar(&opts.Type, "type", "feature", "KEP Type")
	f.StringVarP(&opts.State, "state", "s", "provisional", "KEP State")
	f.StringArrayVar(&opts.SIGS, "sigs", []string{}, "Participating SIGs")
	f.StringArrayVar(&opts.PRRApprovers, "prr-approver", []string{}, "PRR Approver")

	addRepoPathFlag(f, &opts.CommonArgs)
	return cmd
}

func buildCreateReceiptCommand(k *kepctl.Client) *cobra.Command {
	opts := kepctl.PromoteOpts{}
	cmd := &cobra.Command{
		Use:     "receipt [KEP]",
		Short:   "Target a KEP for a release",
		Long:    "Target a KEP for a release",
		Example: `  kepctl create receipt sig-architecture/000-mykep --release v1.22 --stage beta`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.Promote(opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.Release, "release", "", "Release To Target")
	f.StringVar(&opts.Stage, "stage", "", "Stage KEP will be promoted to")
	addRepoPathFlag(f, &opts.CommonArgs)
	return cmd
}
