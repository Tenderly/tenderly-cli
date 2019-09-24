#!/bin/bash

CUR_VERSION=""
NEW_VERSION="$(curl -s https://api.github.com/repos/Tenderly/tenderly-cli/releases/latest | grep tag_name | cut -d'v' -f2 | cut -d'"' -f1)"
EXISTS="$(command -v tenderly)"

if [ "$EXISTS" != "" ]; then
  CUR_VERSION="$(tenderly version | sed -n 1p | cut -d'v' -f3)"
  printf "\nCurrent Version: %s => New Version: %s\n" $CUR_VERSION $NEW_VERSION
fi

if [ "$NEW_VERSION" != "$CUR_VERSION" ]; then

  printf "Installing version %s\n" $NEW_VERSION

  cd /tmp/ > /dev/null

  curl -s https://api.github.com/repos/Tenderly/tenderly-cli/releases/latest \
  | grep "browser_download_url.*Darwin_amd64\.tar\.gz" \
  | cut -d ":" -f 2,3 \
  | tr -d \" \
  | xargs curl -sLOJ

  tarball="$(find . -name "*Darwin_amd64.tar.gz" 2>/dev/null)"
  tar -xzf $tarball

  chmod +x tenderly

  printf "Moving CLI to /usr/local/bin/\n"

  mv tenderly /usr/local/bin/

  cd - > /dev/null

  location="$(which tenderly)"
  printf "Tenderly CLI installed to: %s\n" $location

  version="$(tenderly version | sed -n 1p | cut -d'v' -f3)"
  printf "New Tenderly version installed: %s\n" $version

else
  printf "Latest version already installed\n"
fi
