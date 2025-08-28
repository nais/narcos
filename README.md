# *N*ais *A*dminst*r*ator *C*LI *o*g *S*cripts

> Clies are necessary when the script is too difficult to believe - Pablo Esclibar

## Local Development

### Clone the repository:

```bash
git clone git@github.com:nais/narcos.git
cd narcos
```

### Install tools

Narcos uses [mise](https://mise.jdx.dev/) to handle dependencies and tasks. After installing mise run the following command to install tools:

```bash
mise install
```

### Run tasks

Run `mise run` to see all available tasks. These are some of the most common ones used for local development:

```bash
mise run test # run tests
mise run check # run all static code analysis tools
mise run build # build the CLI
```

### Setup shell completion for local builds

```
source <(./bin/narc completion zsh|bash|fish|powershell)
```

## Contributing

This repo uses [Conventional Commits](https://www.conventionalcommits.org/). Please read up on how to format your commit messages. Please see the [pre-commit hook](script/semantic-commit-hook.sh) to see which types we allow.
