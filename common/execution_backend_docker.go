// TODO: this should be split into a generic interface "ContainerBackend" and
// two implementations: one for Docker and a similar one for Rkt
package dccommon

import (
	"context"
	"fmt"
	"log"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"
	dockerCli "github.com/docker/docker/client"
	uuid "github.com/satori/go.uuid"
)

type DockerBackend struct {
	ExecutionBackend

	HostDataFolder string
}

func NewDockerBackend(hostDataFolder string) (b *DockerBackend, err error) {
	if err != nil {
		return nil, fmt.Errorf("Error creating Docker client: %s", err)
	}
	return &DockerBackend{
		HostDataFolder: hostDataFolder,
	}, nil
}

func (b *DockerBackend) Train(modelID, dataID uuid.UUID) (score float64, err error) {
	// TODO: implement that
	b.RunInUntrustedContainer("test", []string{"sleep", "10s"}, 10*time.Second)
	return 1.0, nil
}

func (b *DockerBackend) Test(modelID, dataID uuid.UUID) (score float64, err error) {
	// TODO: implement that
	return 1.0, nil
}

func (b *DockerBackend) Predict(modelID, dataID uuid.UUID) (prediction []byte, err error) {
	// TODO: implement that
	return []byte(" Irma"), nil
}

func (b *DockerBackend) RunInUntrustedContainer(containerName string, args []string, timeout time.Duration) error {
	log.Printf("[INFO][docker-backend] Running `%s` in untrusted container %s", args, containerName)

	apiClient, err := dockerCli.NewEnvClient()

	ctx, _ := context.WithTimeout(context.Background(), timeout)

	imageName := "alpine"
	log.Print("[DEBUG][docker-backend] Docker context created !")

	// Let's create the container and run the command in it
	containerCreateBody, err := apiClient.ContainerCreate(
		ctx,
		&dockerContainer.Config{
			// Hostname: containerName,
			// Domainname:   "",
			User:         "root:root", // <-- eheheheh
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
			OpenStdin:    false,
			Env:          []string{},
			Cmd:          []string{"echo", "hello"},
			// TODO: make sure not setting the entrypoint makes Docker use the one defined in the image
			// TODO: attach a health check ?
			Image: imageName,
			// TODO: volumes
			WorkingDir:      "/",
			NetworkDisabled: true,
			Labels:          map[string]string{},
			// StopSignal:
			// StopTimeout:
			// Shell
		},
		&dockerContainer.HostConfig{
			AutoRemove: true,
			// TODO: more stuff here
		},
		&dockerNetwork.NetworkingConfig{
		// TODO: same over here
		},
		containerName,
	)
	log.Print("[DEBUG][docker-backend] Docker container created")
	log.Printf("%s", err)
	if err != nil {
		return fmt.Errorf("Error creating Docker container %s: %s", containerName, err)
	}

	// Let's log any warning that was trigger
	for n, warning := range containerCreateBody.Warnings {
		log.Printf("[WARNING %d][docker-backend] Warning creating container: %s", n, warning)
	}

	log.Print("La chatte a ta mere")
	err = apiClient.ContainerStart(
		ctx,
		containerCreateBody.ID,
		dockerTypes.ContainerStartOptions{},
	)
	if err != nil {
		return fmt.Errorf("Error starting Docker container %s: %s", containerName, err)
	}

	// Let's wait for the command to be over
	status, err := apiClient.ContainerWait(ctx, containerCreateBody.ID)
	if err != nil {
		return fmt.Errorf("Error waiting for untrusted container to exit: %s", err)
	}

	log.Printf("[INFO][docker-backend] Untrusted container ran command, status code: %d", status)

	return nil
}
