# Go template from Nethermind Angkor team

This is a template for creating a new Go application made by the Nethermind Angkor team.

## Clone this template

To create a new repository from this template you can use the Github UI. See the
[Github documentation](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-from-a-template)
for more information about creating a repository from a template.

## Initialization

After cloning this repository you need to set up the Go module of the project and
the app name. To do that follow the steps below:

1.  Initialize Go module with the following command:

    ```bash
    go mod init [module-name]
    ```

    The module-name could be for instance `github.com/NethermindEth/project-name`.

2.  Run the `go mod tidy` command to download dependencies. Also you can use the
    `make gomod_tidy` command.

3. Rename the `cmd/app` directory to the name of your application.

3.  Set the `APP_NAME` value in the [.env](.env) and in the [Dockerfile](Dockerfile) to the name
    of your application. Make sure is the same name as the directory you renamed
    in the previous step.

4. Replace the `<repo>` value in the [CONTRIBUTING.md](CONTRIBUTING.md) file with url of your repository
    to make the links work.

5. Replace the `<APP_NAME>` value in the [CONTRIBUTING.md](CONTRIBUTING.md) file with the name of your application.

5. Check `CODEOWNERS` file, currently is an example file and should be updated. You can
    find it in the `.github` directory. Also, see [this GitHub documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
    about `CODEOWNERS` file.

## Documentation

Inside the `docs` directory you can find a Docusaurus project with the initial structure
for the documentation, read the [README.md](docs/README.md) file inside the directory for
more information, also you can read the Docusaurus [documentation](https://docusaurus.io/docs).

## Github Actions

This template has a set of Github Actions workflows that can be used to automate
the CI/CD process of your application. The workflows are located in the [.github/workflows](.github/workflows)
directory.

## TODO

- Support [devcontainer](https://containers.dev/)
- Template for debian package (for PPA)
