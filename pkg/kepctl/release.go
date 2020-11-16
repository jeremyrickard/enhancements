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

package kepctl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"sigs.k8s.io/yaml"
)

type InitReleaseOpts struct {
	CommonArgs
	EnhancementTeam []string
	ReleaseLeads    []string
	ReleaseSemver   string
}

type ownersFile struct {
	Approvers []string `json:"approvers,omitempty"`
	Reviewers []string `json:"reviewers,omitempty"`
}

func (o *ownersFile) save(path string) error {
	b, err := yaml.Marshal(o)
	if err != nil {
		return nil
	}
	return ioutil.WriteFile(path, b, 0644)
}

func (i *InitReleaseOpts) Validate(args []string) error {

	if len(args) == 0 {
		return fmt.Errorf("must provide a release version")
	}

	i.ReleaseSemver = args[0]
	_, err := semver.NewVersion(i.ReleaseSemver)
	if err != nil {
		return fmt.Errorf("invalid release version: %s", err)
	}

	if len(i.EnhancementTeam) == 0 {
		return fmt.Errorf("at least one enhancement team member must be provided")
	}
	if len(i.ReleaseLeads) == 0 {
		return fmt.Errorf("at least one releae lead must be provided")
	}

	return nil
}

func (c *Client) InitRelease(opts *InitReleaseOpts) error {
	repoPath, err := c.findEnhancementsRepo(opts.CommonArgs)
	if err != nil {
		return fmt.Errorf(
			"unable to initalize release: unable to find enhancements repo: %s", err)
	}

	releasePath := filepath.Join(repoPath, "releases", opts.ReleaseSemver)
	// Make the release directory structure, if it doesn't exist
	err = os.MkdirAll(releasePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create release directory: %s", err)
	}

	approvers := []string{}
	for _, person := range opts.EnhancementTeam {
		approvers = append(approvers, fmt.Sprintf("%s // Enhancement Team", person))
	}
	for _, person := range opts.ReleaseLeads {
		approvers = append(approvers, fmt.Sprintf("%s // Release Lead", person))
	}
	sort.Strings(approvers)
	//generate owners
	owners := &ownersFile{
		Reviewers: []string{"release-team"},
		Approvers: approvers,
	}

	//release directories
	//see https://docs.google.com/document/d/1qnfXjQCBrikbbu9F38hdzWEo5ZDBd57Cqi0wG0V6aGc/edit#
	//1.20
	// |- accepted
	acceptedPath := filepath.Join(releasePath, "accepted")
	fmt.Fprintf(c.Out, "===> creating %s accepted directory: %s\n", opts.ReleaseSemver, acceptedPath)
	err = os.Mkdir(acceptedPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create accepted directory: %s", err)
	}
	owners.save(filepath.Join(acceptedPath, "OWNERS"))
	// |- exceptions
	exceptionPath := filepath.Join(releasePath, "exception")
	fmt.Fprintf(c.Out, "===> creating %s exception directory: %s\n", opts.ReleaseSemver, exceptionPath)
	err = os.Mkdir(exceptionPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create exception directory: %s", err)
	}
	owners.save(filepath.Join(exceptionPath, "OWNERS"))
	// |- proposed
	proposedPath := filepath.Join(releasePath, "proposed")
	fmt.Fprintf(c.Out, "===> creating %s proposed directory: %s\n", opts.ReleaseSemver, proposedPath)
	err = os.Mkdir(proposedPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create proposed directory: %s", err)
	}
	owners.save(filepath.Join(proposedPath, "OWNERS"))
	// |- at-risk
	atRiskPath := filepath.Join(releasePath, "at-risk")
	fmt.Fprintf(c.Out, "===>  creating %s at risk directory: %s\n", opts.ReleaseSemver, atRiskPath)
	err = os.Mkdir(atRiskPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create at-risk directory: %s", err)
	}
	owners.save(filepath.Join(atRiskPath, "OWNERS"))
	// |- removed
	removedPath := filepath.Join(releasePath, "removed")
	fmt.Fprintf(c.Out, "===>  creating %s removed directory: %s\n", opts.ReleaseSemver, removedPath)
	err = os.Mkdir(removedPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create removed directory: %s", err)
	}
	owners.save(filepath.Join(removedPath, "OWNERS"))
	return nil
}

type ReleaseOpts struct {
	CommonArgs
	Release string
	Stage   string
	Link    string //this would be like https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/1020-kubectl-staging
}

func (r *ReleaseOpts) Validate(args []string) error {
	err := r.validateAndPopulateKEP(args)
	if err != nil {
		return err
	}
	if r.Release == "" {
		return fmt.Errorf("target release is required to propose a KEP for a release")
	}
	if r.Stage == "" {
		return fmt.Errorf("stage required to target the release")
	}
	r.Link = generateKEPLink(r.CommonArgs)
	return nil

}

func generateKEPLink(opts CommonArgs) string {
	return fmt.Sprintf(
		"https://github.com/kubernetes/enhancements/tree/master/keps/%s",
		opts.KEP,
	)
}

func generateIssueLink(opts CommonArgs) string {
	return fmt.Sprintf(
		"https://github.com/kubernetes/enhancements/issues/%s",
		opts.Number,
	)
}

type releaseContent struct {
	Number string `json:"kep,omitempty"`
	Link   string `json:"link"`
	SIG    string `json:"sig"`
	Stage  string `json:"stage"`
	Issue  string `json:"issue"`
}

func (c *Client) proposeKEP(opts *ReleaseOpts) error {
	repoPath, err := c.findEnhancementsRepo(opts.CommonArgs)
	if err != nil {
		return fmt.Errorf(
			"unable to propose KEP for release: %s", err)
	}
	releasePath := filepath.Join(repoPath, "releases", opts.Release)
	_, err = os.Stat(releasePath)
	if os.IsNotExist(err) {
		return fmt.Errorf(
			"unable to propose KEP for release: release directory does not exist: %s", err)
	}
	proposal := releaseContent{
		SIG:   opts.SIG,
		Stage: opts.Stage,
		Issue: generateIssueLink(opts.CommonArgs),
		Link:  generateKEPLink(opts.CommonArgs),
	}

	fileName := fmt.Sprintf("%s.yaml", opts.Number)
	proposedPath := filepath.Join(releasePath, "proposed", fileName)
	b, err := yaml.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("unable to generate release proposal: %s", err)
	}
	ioutil.WriteFile(proposedPath, b, os.ModePerm)
	fmt.Fprintf(c.Out, "Generated release proposal for %s at %s\n", opts.Name, proposedPath)
	return nil
}

type releaseManifest struct {
	Proposed []releaseContent `json:"proposed"`
	Accepted []releaseContent `json:"accepted"`
	AtRisk   []releaseContent `json:"at-risk"`
	Removed  []releaseContent `json:"removed"`
}

type ReleaseManifestOpts struct {
	CommonArgs
	ReleaseVersion string
}

func (r *ReleaseManifestOpts) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide a release version")
	}
	r.ReleaseVersion = args[0]
	_, err := semver.NewVersion(r.ReleaseVersion)
	if err != nil {
		return fmt.Errorf("invalid release version: %s", err)
	}
	return nil
}

//GenerateRelease builds the YAML manifest for the specified release
func (c *Client) GenerateRelease(opts *ReleaseManifestOpts) error {
	repoPath, err := c.findEnhancementsRepo(opts.CommonArgs)
	if err != nil {
		return fmt.Errorf(
			"unable to generate release manifest: %s", err)
	}

	releasePath := filepath.Join(repoPath, "releases", opts.ReleaseVersion)
	_, err = os.Stat(releasePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("release directory does not exist: %s", err)
	}
	manifest := &releaseManifest{}
	//get approved
	acceptedPath := filepath.Join(releasePath, "accepted")
	files, err := ioutil.ReadDir(acceptedPath)
	if err != nil {
		return fmt.Errorf("unable to read accepted metadata: %s", err)
	}
	var content releaseContent
	acceptedKeps := []releaseContent{}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(acceptedPath, file.Name()))
		if err != nil {
			return fmt.Errorf("error reading accepted KEP: %s", err)
		}
		err = yaml.Unmarshal(b, &content)
		if err != nil {
			return fmt.Errorf("error loading accepted KEP: %s", err)
		}
		num := strings.Split(file.Name(), ".")[0]
		content.Number = num
		acceptedKeps = append(acceptedKeps, content)
	}
	manifest.Accepted = acceptedKeps
	//get at-risk
	atRiskPath := filepath.Join(releasePath, "at-risk")
	files, err = ioutil.ReadDir(atRiskPath)
	if err != nil {
		return fmt.Errorf("unable to read at-risk metadata: %s", err)
	}
	atrisk := []releaseContent{}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(atRiskPath, file.Name()))
		if err != nil {
			return fmt.Errorf("error reading at-risk KEP: %s", err)
		}
		err = yaml.Unmarshal(b, &content)
		if err != nil {
			return fmt.Errorf("error loading at-risk KEP: %s", err)
		}
		num := strings.Split(file.Name(), ".")[0]
		content.Number = num
		atrisk = append(atrisk, content)
	}
	manifest.AtRisk = atrisk
	//get removed
	removedPath := filepath.Join(releasePath, "removed")
	files, err = ioutil.ReadDir(removedPath)
	if err != nil {
		return fmt.Errorf("unable to read removed metadata: %s", err)
	}
	removed := []releaseContent{}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(removedPath, file.Name()))
		if err != nil {
			return fmt.Errorf("error reading removed KEP: %s", err)
		}
		err = yaml.Unmarshal(b, &content)
		if err != nil {
			return fmt.Errorf("error loading removed KEP: %s", err)
		}
		num := strings.Split(file.Name(), ".")[0]
		content.Number = num
		removed = append(removed, content)
	}
	manifest.Removed = removed
	//get exceptions
	b, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("error generating manifest: %s", err)
	}
	buf := new(bytes.Buffer)
	buf.WriteString("### THIS FILE IS AUTOGENERATED ###\n")
	buf.WriteString("# To regenerate run kepctl release generate\n")
	buf.WriteString("\n")
	buf.Write(b)
	manifestPath := filepath.Join(releasePath, "manifest.yaml")
	err = ioutil.WriteFile(manifestPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("unable to generate release manifest: %s", err)
	}
	return nil
}
