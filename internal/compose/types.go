package compose

// DockerComposeUpOptions defines the options for the 'docker compose up' command.
type DockerComposeUpOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Services lists the names of the services to be started.
	Services []string
}

// DockerComposePullOptions defines the options for the 'docker compose pull' command.
type DockerComposePullOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Services lists the names of the services for which images should be pulled.
	Services []string
}

// DockerComposeCreateOptions defines the options for the 'docker compose create' command.
type DockerComposeCreateOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Services lists the names of the services to be created.
	Services []string
	// Build specifies whether to build images before starting containers.
	Build bool
}

// DockerComposeBuildOptions defines the options for the 'docker compose build' command.
type DockerComposeBuildOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Services lists the names of the services to be built.
	Services []string
}

// DockerComposePsOptions defines the options for the 'docker compose ps' command.
type DockerComposePsOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Services, when true, displays the services.
	Services bool
	// Quiet, when true, displays only IDs.
	Quiet bool
	// ServiceName specifies the name of a service.
	ServiceName string
	// FilterRunning, when true, filters to display only running services.
	FilterRunning bool
	// Format specifies the format of the output.
	Format string
	// All, when true, displays all containers.
	All bool
}

// DockerComposeLogsOptions defines the options for the 'docker compose logs' command.
type DockerComposeLogsOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Services lists the names of the services for which logs should be displayed.
	Services []string
	// Follow, when true, follows the log output.
	Follow bool
	// Tail specifies the number of lines from the end of the logs to display.
	// If greater than 0 and Follow is false, it is used for the --tail option.
	Tail int
}

// DockerComposeStopOptions defines the options for the 'docker compose stop' command.
type DockerComposeStopOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
}

// DockerComposeDownOptions defines the options for the 'docker compose down' command.
type DockerComposeDownOptions struct {
	// Path specifies the location of the docker-compose.yaml file.
	Path string
	// Remove named volumes declared in the "volumes" section of the Compose file and anonymous volumes attached to containers.
	Volumes bool
}
