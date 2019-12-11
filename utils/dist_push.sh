#!/bin/bash
: ${ORGDIR:="/src/github.com/TykTechnologies"}
: ${SOURCEBINPATH:="${ORGDIR}/tyk-sync"}
: ${DEBVERS:="ubuntu/xenial ubuntu/bionic debian/jessie debian/stretch debian/buster"}
: ${RPMVERS:="el/6 el/7"}
: ${PKGNAME:="tyk-sync"}

RELEASE_DIR="$SOURCEBINPATH/build"
export PACKAGECLOUDREPO=$PC_TARGET

cd $RELEASE_DIR/

for arch in i386 amd64 arm64
do
    debName="${PKGNAME}_${VERSION}_${arch}.deb"
    rpmName="${PKGNAME}-$VERSION-1.${arch/amd64/x86_64}.rpm"

    for ver in $DEBVERS
    do
        echo "Pushing $debName to PackageCloud $ver"
        package_cloud push tyk/$PACKAGECLOUDREPO/$ver $debName
    done

    for ver in $RPMVERS
    do
        echo "Pushing $rpmName to PackageCloud $ver"
        package_cloud push tyk/$PACKAGECLOUDREPO/$ver $rpmName
    done
done
