#!/bin/bash

CUR_VERSION=""
NEW_VERSION="$(curl -s https://api.github.com/repos/Tenderly/tenderly-cli/releases/latest | grep tag_name | cut -d'v' -f2 | cut -d'"' -f1)"
EXISTS="$(command -v tenderly)"

if [ "$EXISTS" != "" ]; then
  CUR_VERSION="$(tenderly version | sed -n 1p | cut -d'v' -f3)"
  echo -e "\nCurrent Version: $CUR_VERSION => New Version: $NEW_VERSION\n"
fi

if [ "$NEW_VERSION" != "$CUR_VERSION" ]; then

  echo -e "Installing version $NEW_VERSION\n"

  cd /tmp/

  curl -s https://api.github.com/repos/Tenderly/tenderly-cli/releases/latest \
  | grep "browser_download_url.*Linux_amd64\.tar\.gz" \
  | cut -d ":" -f 2,3 \
  | tr -d \" \
  | xargs curl -sLOJ

  tarball="$(find . -name "*Linux_amd64.tar.gz" 2>/dev/null)"
  tar -xzf $tarball

  chmod +x tenderly

  echo -e "Moving CLI to /usr/local/bin/\n"

  mv tenderly /usr/local/bin/

  cd -

  location="$(which tenderly)"
  echo -e "Tenderly CLI installed to: $location\n"

  version="$(tenderly version | sed -n 1p | cut -d'v' -f3)"
  echo -e "New Tenderly version installed: $version\n"

else
  echo -e "Latest version already installed\n"
fi
