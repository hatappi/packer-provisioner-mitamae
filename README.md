This is [Packer Provisioner Plugin](https://www.packer.io/docs/extending/custom-provisioners.html) for use [MItamae](https://github.com/itamae-kitchen/mitamae).

## Installation

```sh
# require go, git
$ go get github.com/hatappi/packer-provisioner-mitamae
```

## Usage

Add the provisioner to packer file.

```json
{
  "builders": [ ... ],
  "provisioners": [{
    "type": "mitamae",
    "recipe_path": "/tmp/recipe.rb"
  }]
}
```

sample is [here](./sample)

## Configuration
**Required**

- recipe_path (string) : recipe path of remote server.

**Optional**

- sudo (bool) : Execute with sudo authority. By default is False.
- mitamae_version (string) : MItamae version. By default this is v1.4.5.
- bin_dir (string) : bin_dir is is the path to download MItamae. By default this is `/usr/local/bin`
- option (string) : It is an option when execute `mitamae local`. For example '-l debug'
- working_directory (string): specify the directory at runtime.
