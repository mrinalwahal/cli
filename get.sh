#!/usr/bin/env bash

# Some helpful functions
yell() { echo -e "${RED}FAILED> $* ${NC}" >&2; }
die() { yell "$*"; exit 1; }
try() { "$@" || die "failed executing: $*"; }
log() { echo -e "--> $*"; }

# Colors for colorizing
RED='\033[0;31m'
GREEN='\033[0;32m'
PURPLE='\033[0;35m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

INSTALL_PATH=${INSTALL_PATH:-"/usr/local/bin"}
NEED_SUDO=0

REPO="mrinalwahal/cli"

function maybe_sudo() {
    if [[ "$NEED_SUDO" == '1' ]]; then
        sudo "$@"
    else
        "$@"
    fi
}

# check for curl
hasCurl=$(which curl)
if [ "$?" = "1" ]; then
    die "You need to install curl to use this script."
fi

release=${1:-latest}

log "Getting $release version..."

version=$(curl --silent "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' |  sed -E 's/.*"([^"]+)".*/\1/')

if [ ! $version ]; then
    log "${YELLOW}"
    log "Failed while attempting to install Nhost CLI. Please manually install:"
    log ""
    log "2. Open your web browser and go to https://github.com/$REPO/releases/latest"
    log "2. Download the CLI from latest release for your platform. Name it 'nhost'."
    log "3. chmod +x ./nhost"
    log "4. mv ./nhost /usr/local/bin"
    log "${NC}"
    die "exiting..."
fi

log "Latest version is $version"

# check for existing nhost installation
hasCli=$(which nhost)
if [ "$?" = "0" ]; then
    log ""
    log "${GREEN}You already have the Nhost CLI at '${hasCli}'${NC}"
    export n=3
    log "${YELLOW}Downloading again in $n seconds... Press Ctrl+C to cancel.${NC}"
    log ""
    sleep $n
fi

# get platform and arch
platform='unknown'
unamestr=`uname`
if [[ "$unamestr" == 'Linux' ]]; then
    platform='linux'
elif [[ "$unamestr" == 'Darwin' ]]; then
    platform='darwin'
elif [[ "$unamestr" == 'Windows' ]]; then
    platform='windows'
fi

if [[ "$platform" == 'unknown' ]]; then
    die "Unknown OS platform"
fi

arch='unknown'
archstr=`uname -m`
if [[ "$archstr" == 'x86_64' ]]; then
    arch='amd64'
else
    arch='386'
fi

# some variables
suffix="-${platform}-${arch}"
targetFile="nhost-$version$suffix.tar.gz"

if [ -e $targetFile ]; then
    rm $targetFile
fi

log "${PURPLE}Downloading Nhost for $platform-$arch to ${targetFile}${NC}"
url=https://github.com/mrinalwahal/cli/releases/download/$version/$targetFile

try curl -L -f -o $targetFile "$url"
try chmod +x $targetFile
try rm /usr/local/bin/nhost
try tar -xvf $targetFile -C /usr/local/bin
rm ./$targetFile

log "${GREEN}Download complete!${NC}"
echo
nhost version
echo
log "${BLUE}Use Nhost CLI with: nhost --help${NC}"
