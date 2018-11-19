#!/usr/bin/env bash
set -e

git fetch
if [ "$(git rev-parse HEAD)" != "$(git rev-parse @{u})" ]; then
    echo "branch is not up-to-date with upstream. 'git pull' to update your local branch."
    exit 1
fi

latest=$(git describe --abbrev=0 --tags)

parts=(${latest//./ })
major=${parts[0]}
minor=${parts[1]}
patch=${parts[2]}

version=${major}.$((minor+1)).0
echo $version

dir=releases
mkdir -p $dir
cd $dir

platforms=("linux/amd64" "darwin/amd64" "windows/amd64")
for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output="starlightd"
    echo "building starlightd for $GOOS/$GOARCH..."
    if [ $GOOS = "windows" ]; then
        GOOS=$GOOS GOARCH=$GOARCH go build -o starlightd.exe github.com/interstellar/starlight/cmd/starlightd
        zip starlightd-$GOOS-$GOARCH-$version.zip starlightd.exe
        rm starlightd.exe
        echo "starlightd-$GOOS-$GOARCH-$version.zip done"
    else
        GOOS=$GOOS GOARCH=$GOARCH go build -o $output github.com/interstellar/starlight/cmd/starlightd
        tar -czf starlightd-$GOOS-$GOARCH-$version.tar.gz starlightd
        rm starlightd
        echo "starlightd-$GOOS-$GOARCH-$version.tar.gz done"
    fi
done

# create new tag with specified message
git tag -a $version
echo "created tag: $(git rev-parse HEAD) as $version"
