# claude-commit-msg-gen

A CLI tool that automatically generates [Conventional Commits](https://www.conventionalcommits.org/) messages from staged diffs using the [Anthropic API](https://www.anthropic.com/) and [Lefthook](https://lefthook.dev/).

For architecture and design details, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Setup

### 1. Set Anthropic API key

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. Install Lefthook

```sh
brew install lefthook
# or
pnpm add -g lefthook
```

### 3. Install the binary

```sh
pnpm install -g @himenon/claude-commit-msg-gen
# or
npm install -g @himenon/claude-commit-msg-gen
```

### 4. Enable the hook

```sh
lefthook install
```

`git commit` will now auto-generate commit messages.

## Keep the API key out of your shell profile

Write the API key in `lefthook-local.yml` instead. This file is listed in `.gitignore` and will never be committed.

```yaml
# lefthook-local.yml (.gitignore'd)
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      env:
        ANTHROPIC_API_KEY: "sk-ant-..."
```

`lefthook-local.yml` is merged on top of `lefthook.yml`. Only the keys you specify are overridden; everything else continues to use the values from `lefthook.yml`.

## Troubleshooting

**`Binary not found` is shown**

```sh
pnpm run build
```

**`ANTHROPIC_API_KEY is not set` is shown**

```sh
echo $ANTHROPIC_API_KEY
```

**Temporarily disable auto-generation**

```sh
LEFTHOOK=0 git commit
```

> For the shell-script alternative using the `claude` CLI, see [docs/SHELL-SCRIPT-ALTERNATIVE.md](docs/SHELL-SCRIPT-ALTERNATIVE.md).
