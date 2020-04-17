#!/bin/bash

set -ex

: ${ORGDIR:="/src/github.com/TykTechnologies"}
: ${SIGNKEY:="12B5D62C28F57592D1575BD51ED14C59E37DAC20"}
: ${BUILDPKGS:="1"}
: ${SIGNPKGS:="1"}
TYK_IB_SRC_DIR=$ORGDIR/tyk-sync
BUILDTOOLSDIR=$TYK_IB_SRC_DIR/build_tools

if [ $BUILDPKGS == "1" ]; then
    echo Configuring gpg-agent-config to accept a passphrase
    mkdir ~/.gnupg && chmod 700 ~/.gnupg
    cat >> ~/.gnupg/gpg-agent.conf <<EOF
allow-preset-passphrase
debug-level expert
log-file /tmp/gpg-agent.log
EOF
    gpg-connect-agent reloadagent /bye

    echo "Importing signing key"
    gpg --list-keys | grep -w $SIGNKEY && echo "Key exists" || gpg --batch --import $BUILDTOOLSDIR/tyk.io.signing.key
    bash $BUILDTOOLSDIR/unlock-agent.sh $SIGNKEY
fi

DESCRIPTION="Tyk Sync"
RELEASE_DIR="$TYK_IB_SRC_DIR/build"
BUILD="tyk-sync-$VERSION"
BUILD_DIR="$RELEASE_DIR/$BUILD"

cd $TYK_IB_SRC_DIR

echo "Creating build folder ($BUILD_DIR)"
mkdir -p $BUILD_DIR

# ---- APP BUILD START ---
echo "Building application"
gox -osarch="linux/arm64 linux/amd64 linux/386"
# ---- APP BUILD END ---

# ---- CREATE TARGET FOLDER ---
echo "Copying files"
cd $TYK_IB_SRC_DIR
cp LICENSE.md $BUILD_DIR/
cp README.md $BUILD_DIR/

cd $RELEASE_DIR
echo "Removing old builds"
rm -f *.deb
rm -f *.rpm
rm -f *.tar.gz

echo "LINUX"
FPMCOMMON=(
    --name tyk-sync
    --description "$DESCRIPTION"
    -v $VERSION
    --vendor "Tyk Technologies Ltd"
    -m "<info@tyk.io>"
    --url "https://tyk.io"
    -s dir
    -C $BUILD_DIR
)

for arch in i386 amd64 arm64
do
    echo "Creating $arch Tarball"
    cd $TYK_IB_SRC_DIR
    mv tyk-sync_linux_${arch/i386/386} $BUILD_DIR/tyk-sync
    cd $RELEASE_DIR
    tar -pczf $RELEASE_DIR/tyk-sync-$arch-$VERSION.tar.gz $BUILD/

    if [ $BUILDPKGS == "1" ]; then
        echo "Building $arch packages"
        fpm "${FPMCOMMON[@]}" -a $arch -t deb ./=/opt/tyk-sync
        fpm "${FPMCOMMON[@]}" -a $arch -t rpm ./=/opt/tyk-sync

        if [ $SIGNPKGS == "1" ]; then
            echo "Signing $arch RPM"
            rpm --define "%_gpg_name Team Tyk (package signing) <team@tyk.io>" \
                --define "%__gpg /usr/bin/gpg" \
                --addsign *.rpm || (cat /tmp/gpg-agent.log; exit 1)
            echo "Signing $arch DEB"
            for i in *.deb
            do
                dpkg-sig --sign builder -k $SIGNKEY $i || (cat /tmp/gpg-agent.log; exit 1)
            done
        fi

    fi
done
