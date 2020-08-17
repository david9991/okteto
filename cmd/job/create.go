// Copyright 2020 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

import (
	"context"
	"errors"
	"fmt"

	"github.com/okteto/okteto/cmd/utils"
	"github.com/okteto/okteto/pkg/analytics"
	"github.com/okteto/okteto/pkg/cmd/login"
	k8Client "github.com/okteto/okteto/pkg/k8s/client"
	"github.com/okteto/okteto/pkg/k8s/jobs"
	"github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"
	"github.com/spf13/cobra"
)

//Create creates a job
func Create(ctx context.Context) *cobra.Command {
	var namespace string
	var devPath string

	cmd := &cobra.Command{
		Use:   "job <name>",
		Short: fmt.Sprintf("Creates a dev version of the job [beta]"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := login.WithEnvVarIfAvailable(ctx); err != nil {
				return err
			}

			dev, err := utils.LoadDev(devPath)
			if err != nil {
				return err
			}

			if err := dev.UpdateNamespace(namespace); err != nil {
				return err
			}

			err = executeCreateJob(ctx, dev, args[0])
			analytics.TrackCreateJob(err == nil)
			return err
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("create job requires one argument")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&devPath, "file", "f", utils.DefaultDevManifest, "path to the manifest file")
	return cmd
}

func executeCreateJob(ctx context.Context, dev *model.Dev, job string) error {
	if len(dev.Jobs) == 0 {
		return fmt.Errorf("Unable to find any job with the provided name")
	}

	if job == "" {
		job = dev.Jobs[0].Name
	}

	client, _, namespace, err := k8Client.GetLocal()
	if err != nil {
		return err
	}

	if dev.Namespace == "" {
		dev.Namespace = namespace
	}

	for _, j := range dev.Jobs {
		if j.Name == job {
			n, err := jobs.CreateDevJob(j, dev, client)
			if err != nil {
				return fmt.Errorf("failed to create job: %s", err)
			}

			log.Success("started job: %s", n)
			return nil
		}
	}

	return fmt.Errorf("Unable to find any job with the provided name")
}
