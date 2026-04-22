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

## Fasit

The `narc fasit` command group provides tools to inspect and manage Fasit configurations. Run `narc fasit login` to authenticate.

### Common Flags

- `--output json|yaml`: Set output format for automation or deep inspection.

### Examples

#### Inspecting Tenants and Environments

```bash
# List all tenants
narc fasit tenant list

# Get details for a specific tenant
narc fasit tenant get my-tenant

# Inspect a specific environment
narc fasit env get my-tenant prod
```

#### Working with Features

```bash
# Get status of a feature across all environments
narc fasit feature status my-feature

# Get configuration for a feature in an environment
narc fasit env feature get my-tenant prod my-feature

# View rollout logs and helm diff for an environment feature
narc fasit env feature logs my-tenant prod my-feature

# View computed helm values
narc fasit env feature helm my-tenant prod my-feature
```

#### Rollouts and Auditing

```bash
# List recent rollouts
narc fasit rollout list

# Get details for a specific rollout
narc fasit rollout get my-feature 1.2.3

# Check audit history (note: currently a placeholder)
narc fasit env feature audit my-tenant prod my-feature
```

## Contributing

This repo uses [Conventional Commits](https://www.conventionalcommits.org/). Please read up on how to format your commit messages. Please see the [pre-commit hook](script/semantic-commit-hook.sh) to see which types we allow.
