package compose

// DockerComposeUpOptions : Represent docker compose up command options
type DockerComposeUpOptions struct {
	// Path : path to docker-compose.yaml
	Path string
	// Services : services names
	Services []string
}

// DockerComposePullOptions represents 'docker compose pull' command options
type DockerComposePullOptions struct {
	// Path to the docker-compose.yaml
	Path string
	// Services names
	Services []string
}

// DockerComposeCreateOptions represents `docker compose create` command options
type DockerComposeCreateOptions struct {
	// Path to the docker-compose.yaml
	Path string
	// Services names
	Services []string
}

// DockerComposeBuildOptions represents `docker compose build` command options
type DockerComposeBuildOptions struct {
	// Path to the docker-compose.yaml
	Path string
	// Services names
	Services []string
}

// DockerComposePsOptions : Represents docker compose ps command options
type DockerComposePsOptions struct {
	// Path : path to docker-compose.yaml
	Path string
	// Services : use with --services to display services
	Services bool
	// Quiet : use with --quiet to display only IDs
	Quiet bool
	// ServiceName: Service argument
	ServiceName string
	// FilterRunning : use with --filter status=running
	FilterRunning bool
}

// DockerComposeLogsOptions : Represents docker compose log command options
type DockerComposeLogsOptions struct {
	// Path : path to docker-compose.yaml
	Path string
	// Services : services names
	Services []string
	// Follow : use with --follow
	Follow bool
	// Tail : if greater than 0 and Follow is False used for --tail
	Tail int
}

// DockerComposeDownOptions : Represents docker compose down command options
type DockerComposeDownOptions struct {
	// Path : path to docker-compose.yaml
	Path string
}
