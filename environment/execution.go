package environment

import (
	"path/filepath"

	"github.com/nhost/cli/hasura"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
)

//  Explain what it does
func (e *Environment) Execute() error {

	var err error

	//  Update environment state
	e.UpdateState(Executing)

	//	Cancel the execution context as soon as this function completed
	defer e.ExecutionCancel()

	//  check if this is the first time dev env is running
	firstRun := !util.PathExists(filepath.Join(nhost.DOT_NHOST, "db_data"))

	//  Validate the availability of required docker images,
	//  and download the ones that are missing
	status.Set("Validating required images")
	if err := e.CheckImages(); err != nil {
		return err
	}

	//	Generate configuration for every service.
	//	This generates all env vars, mount points and commands
	if err := e.Config.Init(e.Port); err != nil {
		return err
	}

	//	Create the Nhost network if it doesn't exist
	if err := e.PrepareNetwork(); err != nil {
		return err
	}

	//	Create and start the containers
	status.Set("Starting services")
	for _, item := range e.Config.Services {

		//	Only those services which have a container configuration
		//	This is being done to exclude FUNCTIONS
		if item.Config != nil {

			//	We are passing execution context, and not parent context,
			//	because if this execution is cancelled in between,
			//	we want docker to abort this procedure.
			if err := item.Run(e.Docker, e.ExecutionContext, e.Network); err != nil {
				return err
			}
		}
	}

	//
	//	Update the ports and IDs of services against the running ones
	//
	//	Fetch list of existing containers
	containers, err := e.GetContainers()
	if err != nil {
		return err
	}

	//	Wrap fetched containers as services in the environment
	_ = e.WrapContainersAsServices(containers)

	//	status.Info("Running a quick health check on services")
	if err := e.HealthCheck(e.ExecutionContext); err != nil {
		return err
	}

	e.UpdateState(Executing)
	//	e.Status.Set("Preparing your data")

	//  Now that Hasura container is active,
	//  initialize the Hasura client.
	e.Hasura = &hasura.Client{}
	if err := e.Hasura.Init(
		e.Config.Services["hasura"].Address,
		util.ADMIN_SECRET,
		nil,
	); err != nil {
		return err
	}

	//
	//  Apply migrations and metadata
	//
	//	status.Info("Preparing your data")
	if err = e.Prepare(); err != nil {
		return err
	}

	//
	//	Fixes #169
	//
	//	Detect inconsistent metadata,
	//	and restart equivalent containers to fix breaking metadata changes.
	inconsistentMetadata, err := e.Hasura.GetInconsistentMetadata()
	if !inconsistentMetadata.IsConsistent {

		status.Set("Fixing inconsistent metadata")

		for _, object := range inconsistentMetadata.InconsistentObjects {
			log.Debug(object.Reason)

			//	Fetch the equivalent container
			for x := range e.Config.Services {
				if x == object.Definition.Schema {

					//	Restart the container
					log.Debugf("Restarting %s container", x)
					if err := e.Docker.ContainerRestart(e.Context, e.Config.Services[x].ID, nil); err != nil {
						return err
					}
				}
			}
		}
	}

	//	Re-run the healthcheck after restarting containers
	if err := e.HealthCheck(e.ExecutionContext); err != nil {
		return err
	}

	//	End Fix #169

	//
	//  Apply Seeds if required
	//
	if firstRun && util.PathExists(filepath.Join(nhost.SEEDS_DIR, nhost.DATABASE)) {
		if err = e.Seed(filepath.Join(nhost.SEEDS_DIR, nhost.DATABASE)); err != nil {
			log.Debug(err)
			e.Cleanup()
		}
	}

	return err
}
