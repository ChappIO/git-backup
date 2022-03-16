# git-backup

A cli to pull all your GitHub repositories for backup purposes.

## Download

The latest version can be downloaded from the [releases page](https://github.com/ChappIO/git-backup/releases).

## Configuration

Example Configuration:
```yaml
# The github section contains backup jobs for
# GitHub and GitHub Enterprise
github:
    # (optional) The job name. This is used to
    # create a subfolder in the backup folder. 
    # (default: GitHub)
  - job_name: github.com
    # (required) The GitHub personal access
    # token. Create one with the scopes:
    # "read:org, repo"
    # https://github.com/settings/tokens/new?scopes=repo,read:org
    access_token: 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
```

## Usage

```asciidoc
Usage: git-backup

Options:
  -backup.path string
        The target path to the backup folder. (default "backup")
  -config.file string
        The path to your config file. (default "git-backup.yml")
```