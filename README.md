# git-backup

A cli to pull all your git repositories for backup purposes. It is completely stand-alone. You don't even need to install git.

## Download

The latest version can be downloaded from the [releases page](https://github.com/ChappIO/git-backup/releases).

## Configuration File

Example yaml Configuration:
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
    # (optional) Back up repos you own.
    # (default: true)
    owned: true
    # (optional) Back up repos you starred.
    # (default: true)
    starred: true
    # (optional) Back up repos on which you 
    # are a collaborator. (default: true)
    collaborator: true
    # (optional) Back up repos owned by 
    # organisations of which you are a member.
    # (default: true)
    org_member: true
    # (optional) Set this url to connect to
    # your self-hosted github install.
    # (default: https://api.github.com)
    url: https://github.mydomain.com
    # (optional) Exclude this list of repos
    exclude:
      - my-namespace/excluded-repository-name
# The gitlab section contains backup jobs for
# GitLab.com and GitLab on premise
gitlab:
  # (optional) The job name. This is used to
  # create a subfolder in the backup folder. 
  # (default: GitLab)
  - job_name: gitlab.com
    # (required) The GitLab access token.
    # Create one with the scopes: "api"
    # https://gitlab.com/-/profile/personal_access_tokens?scopes=api&name=git-backup
    access_token: glpat-6t78yuihy789uy8t768
    # (optional) Back up repos you own.
    # (default: true)
    owned: true
    # (optional) Back up repos you starred.
    # (default: true)
    starred: true
    # (optional) Back up repos owned by 
    # teams of which you are a member.
    # (default: true)
    member: true
    # (optional) Set this url to connect to
    # your self-hosted gitlab install.
    # (default: https://gitlab.com/)
    url: https://gitlab.mydomain.com
    # (optional) Exclude this list of repos
    # or whole organizations/users
    exclude:
      - my-excluded-org
      - my-excluded-user
      - my-namespace/excluded-repository-name
```

## Usage: CLI

```asciidoc
Usage: git-backup

Options:
  -backup.path string
      The target path to the backup folder. (default "backup")
  -config.file string
      The path to your config file. (default "git-backup.yml")
  -backup.fail-at-end
      Fail at the end of backing up repositories, rather than right away.
  -backup.bare-clone
      Make bare clones without checking out the main branch.
  -insecure
      Use this flag to disable verification of SSL/TLS certificates
  -version
      Show the version number and exit.
```

## Usage: Docker

First, create your [git-backup.yml file](#configuration-file) at `/path/to/your/backups`.

Then update your backups using the mounted volume.

```bash
docker run -v /path/to/backups:/backups ghcr.io/chappio/git-backup:1
```

### Parameters

You can specify several parameters when starting this container.

| **Parameter**                  | **Description**                                                                        |
|--------------------------------|----------------------------------------------------------------------------------------|
| `-v /path/to/backups:/backups` | Mount the folder where you want to store your backups and read you configuration file. |
| `-e TZ=Europe/Amsterdam`       | Set the timezone used for logging.                                                     |
| `-e PUID=0`                    | Set the user id of the unix user who will own the backup files in /backup.             |
| `-e PGID=0`                    | Set the group id of the unix user's group who will own the backup files.               | 
