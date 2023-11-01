#!/bin/bash
#exit when any command fails
set -e
cd /home/runner/work/eigenlayer/eigenlayer/eigenlayer
go install github.com/golang/mock/mockgen@v1.6.0
go generate ./...
mkdir -p build/package/debian/src/github.com/NethermindEth/eigenlayer/
rsync -aq . build/package/debian/src/github.com/NethermindEth/eigenlayer/ --exclude build/ --exclude .git/
cd build/package/debian/src/github.com/NethermindEth/eigenlayer/ && go mod vendor
cd /home/runner/work/eigenlayer/eigenlayer/eigenlayer

export SVERSION=${VERSION#v}
echo "eigenlayer ($SVERSION) jammy; urgency=medium

  * EigenLayer ($SVERSION release)

 -- Nethermind <devops@nethermind.io>  $( date -R )" > /home/runner/work/eigenlayer/eigenlayer/eigenlayer/build/package/debian/debian/changelog

cd build/package/debian
debuild -S -uc -us
cd ..
echo 'Signing package'
debsign -p 'gpg --batch --yes --no-tty --pinentry-mode loopback --passphrase-file /tmp/PASSPHRASE' -S -k $PPA_GPG_KEYID eigenlayer_${SVERSION}_source.changes
echo 'Uploading'
dput --force --debug ppa:nethermindeth/eigenlayer eigenlayer_${SVERSION}_source.changes
echo "Publishing Eigenlayer to PPA complete"
echo 'Cleanup'
rm -r eigenlayer_$SVERSION*
