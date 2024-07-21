package wrapper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	d_client "github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	containerWaitInterval = 1 * time.Second
)

var imageLoadSuccess = regexp.MustCompile(`Loaded image: ([\w\-]+:[\w\-]+)`)

type DockerClientWrapper struct {
	client *d_client.Client
}

func NewDockerClientWrapper(ctx context.Context) (*DockerClientWrapper, error) {
	log.Debug().Msg("Connecting to docker daemon")
	cli, err := d_client.NewClientWithOpts(d_client.FromEnv, d_client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}

	resp, err := cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping docker daemon: %v", err)
	}
	log.Info().
		Str("APIVersion", resp.APIVersion).
		Str("OSType", resp.OSType).
		Bool("Experimental", resp.Experimental).
		Str("BuilderVersion", string(resp.BuilderVersion)).
		Msg("Docker daemon connected")

	return &DockerClientWrapper{
		client: cli,
	}, nil
}

func (w *DockerClientWrapper) Close() error {
	err := w.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close docker client: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) ListImages(ctx context.Context) ([]image.Summary, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("Listing docker images")
	images, err := w.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %v", err)
	}
	return images, nil
}

func (w *DockerClientWrapper) PullImage(ctx context.Context, imageReference string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Pulling docker image: %s", imageReference)
	reader, err := w.client.ImagePull(ctx, imageReference, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull docker image: %v", err)
	}
	io.Copy(io.Discard, reader)
	return nil
}

func (w *DockerClientWrapper) InspectImage(ctx context.Context, imageReference string) (*types.ImageInspect, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Inspecting docker image: %s", imageReference)
	image, _, err := w.client.ImageInspectWithRaw(ctx, imageReference)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect docker image: %v", err)
	}
	return &image, nil
}

func (w *DockerClientWrapper) ImportImage(ctx context.Context, input io.Reader, imageReference string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Importing docker image: %s", imageReference)
	reader, err := w.client.ImageImport(ctx, image.ImportSource{
		Source:     input,
		SourceName: "-",
	}, imageReference, image.ImportOptions{})
	if err != nil {
		return fmt.Errorf("failed to import docker image: %v", err)
	}
	io.Copy(io.Discard, reader)
	return nil
}

func (w *DockerClientWrapper) LoadImage(ctx context.Context, data []byte) (string, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Loading docker image of %d bytes", len(data))
	reader := bytes.NewReader(data)
	resp, err := w.client.ImageLoad(ctx, reader, true)
	if err != nil {
		return "", fmt.Errorf("failed to load docker image: %v", err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read load result: %v", err)
	}
	log.Debug().Msgf("Load command output: %s", b)
	cmdOutput := string(b)
	cmdOutput = cmdOutput[0 : len(cmdOutput)-2] // remove \r\n

	matches := imageLoadSuccess.FindStringSubmatch(cmdOutput)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to load docker image: '%s'", cmdOutput)
	}

	imageName := matches[1]
	log.Debug().Msgf("Docker image loaded: %s", imageName)
	return imageName, nil
}

func (w *DockerClientWrapper) RemoveImage(ctx context.Context, imageReference string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Removing docker image: %s", imageReference)
	if _, err := w.client.ImageRemove(ctx, imageReference, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	}); err != nil {
		return fmt.Errorf("failed to remove docker image: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) CreateVolume(ctx context.Context, volumeName, driver string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Creating new docker volume: %s", volumeName)
	_, err := w.client.VolumeCreate(ctx, volume.CreateOptions{
		Name:   volumeName,
		Driver: driver,
	})
	if err != nil {
		return fmt.Errorf("failed to create docker volume: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) RemoveVolume(ctx context.Context, volumeName string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Removing docker volume: %s", volumeName)
	err := w.client.VolumeRemove(ctx, volumeName, true)
	if err != nil {
		return fmt.Errorf("failed to remove docker volume: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) RemoveAllVolumes(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("Removing all docker volumes")
	vols, err := w.client.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list docker volumes: %v", err)
	}
	for _, vol := range vols.Volumes {
		if err := w.RemoveVolume(ctx, vol.Name); err != nil {
			return err
		}
	}
	return nil
}

func (w *DockerClientWrapper) RunContainer(ctx context.Context, contCfg *container.Config, hostCfg *container.HostConfig, networkCfg *network.NetworkingConfig, containerName string) (string, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Starting docker container: %s", containerName)
	resp, err := w.client.ContainerCreate(ctx, contCfg, hostCfg, networkCfg, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to run docker container: %v", err)
	}
	err = w.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		if err2 := w.RemoveContainer(ctx, containerName); err2 != nil {
			return "", fmt.Errorf("failed to remove docker container after unsuccessfull start: %v", err)
		}
		return "", fmt.Errorf("failed to run docker container: %v", err)
	}
	return resp.ID, nil
}

func (w *DockerClientWrapper) StopContainer(ctx context.Context, containerName string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Stopping docker container: %s", containerName)
	err := w.client.ContainerStop(ctx, containerName, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop docker container: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) ContainerExists(ctx context.Context, containerName string) (bool, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Checking if container exists: %s", containerName)
	conts, err := w.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return false, fmt.Errorf("failed to list docker containers: %v", err)
	}
	for _, con := range conts {
		if con.ID == containerName {
			return true, nil
		}
		for _, name := range con.Names {
			if name == "/"+containerName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (w *DockerClientWrapper) RemoveContainer(ctx context.Context, containerName string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Removing docker container: %s", containerName)
	err := w.client.ContainerRemove(ctx, containerName, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker container: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) RemoveAllContainers(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("Removing all docker containers")
	conts, err := w.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to list docker containers: %v", err)
	}
	for _, con := range conts {
		cName := con.ID
		if len(con.Names) > 0 {
			cName = con.Names[0]
		}
		if err := w.RemoveContainer(ctx, cName); err != nil {
			return err
		}
	}
	return nil
}

func (w *DockerClientWrapper) WaitForContainer(ctx context.Context, containerRef string) error {
	log := zerolog.Ctx(ctx)
	for {
		cont, err := w.client.ContainerInspect(ctx, containerRef)
		if err != nil {
			return fmt.Errorf("failed to inspect docker container: %v", err)
		}
		status := cont.State.Health.Status
		if status == types.Healthy {
			log.Debug().Msgf("Container %s status: %s", containerRef, status)
			return nil
		}
		log.Debug().Msgf("Waiting for docker container: %s, current status: %s", containerRef, status)
		time.Sleep(containerWaitInterval)
	}
}

func (w *DockerClientWrapper) CreateNetwork(ctx context.Context, networkName string, options network.CreateOptions) (string, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Creating new docker network: %s", networkName)
	resp, err := w.client.NetworkCreate(ctx, networkName, options)
	if err != nil {
		return "", fmt.Errorf("failed to create docker network: %v", err)
	}
	return resp.ID, nil
}

func (w *DockerClientWrapper) ListNetworks(ctx context.Context) ([]network.Inspect, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("Listing docker networks")
	nets, err := w.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list docker networks: %v", err)
	}
	return nets, nil
}

func (w *DockerClientWrapper) RemoveNetwork(ctx context.Context, networkID string) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msgf("Removing docker network: %s", networkID)
	err := w.client.NetworkRemove(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to remove docker network: %v", err)
	}
	return nil
}

func (w *DockerClientWrapper) RemoveAllNetworks(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("Removing all docker networks")
	nets, err := w.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list docker networks: %v", err)
	}
	for _, net := range nets {
		if net.Name == "host" || net.Name == "none" || net.Name == "bridge" {
			continue // pre-defined networks cannot be removed
		}
		if err := w.RemoveNetwork(ctx, net.ID); err != nil {
			return err
		}
	}
	return nil
}
