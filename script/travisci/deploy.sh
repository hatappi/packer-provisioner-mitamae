#!/bin/sh -ex
VERSION=`echo $TRAVIS_TAG | sed -e "s/v//g"`

echo "*** $VERSION deploy start ***"

goxc \
  -arch="386 amd64" \
  -os="linux darwin" \
  -+tasks=clean,compile,archive \
  -o="{{.Dest}}{{.PS}}{{.Version}}{{.PS}}packer-provisioner-mitamae-{{.Os}}-{{.Arch}}{{.Ext}}" \
  -resources-exclude=".go,LICENSE,README.md" \
  -pv=$VERSION \
    publish-github \
      -owner=hatappi \
      -repository=packer-provisioner-mitamae \
      -apikey=$GITHUB_TOKEN \
      -include="*"

echo "*** $VERSION deploy end ***"
