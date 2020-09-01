# splat

Lays out the structure for a new Spinup API.

## Usage

Download the latest binary for your platform.

### Initialize from a local template

`splat init -l template-directory github.com/YaleSpinup/new-api`

### Initialize from a template in Github

Initialize a new API from a repository in Github.  `Splat` works off of releases and by default will pull the latest release.

`splat init -g 'YaleSpinup/api-tmpl' github.com/YaleSpinup/new-api`

 You can specify the release tag with the `--tag` flag.

 `splat init -g 'YaleSpinup/api-tmpl' --tag v0.2.0 github.com/YaleSpinup/new-api`

## Author

E Camden Fisher <camden.fisher@yale.edu>

## License

GNU Affero General Public License v3.0 (GNU AGPLv3)
Copyright (c) 202 Yale University