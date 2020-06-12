#  Version tests

get_latest_release() {
  curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

version=$(darknode --version | cut -d ' ' -f 4)
printf "Injected version = %s\n" "$version"

expectVer=$(cat ../VERSION | tr -d "[:space:]")
printf "Expected version = %s\n" "$expectVer"

latestReleaseVer=$(get_latest_release "renproject/darknode-cli")
printf "Latest release = %s\n" "$latestReleaseVer"

#  1. Check the version is not equal to the latest release on github
if [ "$version" = "$latestReleaseVer" ]; then
    echo "you forget to pump the version number"
    exit 1
fi

#  2. The binary has the correct version number
if [ "$version" != "$expectVer" ]; then
    printf "invalid version number of the binary, have = %s, expected = %s \n" "$version" "$expectVer"
    exit 1
fi

