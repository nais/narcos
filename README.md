# *N*ais *A*dminst*r*ator *C*LI *o*g *S*cripts

> Clies are necessary when the script is too difficult to believe - Pablo Esclibar

## Local Development

### Install the required go version:

```bash
mise install
```

### Build cli

```
mise run build
```

### Run tests

```
mise run test
```

### Verify cli

```
./bin/narc --version
```

### Setup shell completion for local builds

```
source <(./bin/narc completion zsh|bash|fish|powershell)
```

## Contributing

This repo uses [Conventional Commits](https://www.conventionalcommits.org/). Please read up on how to format your commit messages. Please see the [pre-commit hook](script/semantic-commit-hook.sh) to see which types we allow.
