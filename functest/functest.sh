#!/bin/bash
cd $(dirname $0)
set -ex

rm -rf scratch
mkdir -p scratch/token
export SOFTHSM2_CONF=./token.conf
softhsm2-util --slot=0 --init-token --label=functest --pin=123456 --so-pin=12345678
relic="relic -c ./testconf.yml"
$relic import-key -k rsa2048 -f testkeys/rsa2048.key
$relic serve &
spid=$!
trap "kill $spid" EXIT INT QUIT TERM

signed=scratch/signed
mkdir -p $signed
echo

### RPM
pkg="zlib-1.2.8-10.fc24.i686.rpm"
relic verify --cert "testkeys/RPM-GPG-KEY-fedora-25-i386" "packages/$pkg"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.pgp" "$signed/$pkg"
echo

### Starman
pkg="zlib-1.2.8-10.fc24.i686.tar"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.pgp" "$signed/$pkg"
echo

### DEB
pkg="zlib1g_1.2.8.dfsg-5_i386.deb"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.pgp" "$signed/$pkg"
echo

### PGP
relic verify --cert "testkeys/ubuntu2012.pgp" "packages/InRelease"
relic verify --cert "testkeys/ubuntu2012.pgp" "packages/Release.gpg" --content "packages/Release"
$relic remote sign-pgp -u rsa2048 -ba "packages/Release" -o "$signed/Release.gpg"
relic verify --cert "testkeys/rsa2048.pgp" "$signed/Release.gpg" --content "packages/Release"
$relic remote sign-pgp -u rsa2048 --clearsign "packages/Release" -o "$signed/InRelease"
relic verify --cert "testkeys/rsa2048.pgp" "$signed/InRelease"
echo

### JAR
pkg="hello.jar"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### EXE
pkg="ClassLibrary1.dll"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### MSI
pkg="dummy.msi"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### appx
pkg="App1_1.0.3.0_x64.appx"
relic verify --cert "testkeys/ralph.crt" "packages/$pkg"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### CAB
pkg="dummy.cab"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### CAT
pkg="hyperv.cat"
relic verify --cert "testkeys/msroot.crt" "packages/$pkg"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### XAP
pkg="dummy.xap"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### Powershell
pkg="hello.ps1"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
pkg="hello.ps1xml"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
pkg="hello.mof"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

### ClickOnce
pkg="WindowsFormsApplication1.exe.manifest"
$relic remote sign -k rsa2048 -f "packages/$pkg" -o "$signed/$pkg"
relic verify --cert "testkeys/rsa2048.crt" "$signed/$pkg"
echo

set +x
echo
echo OK
echo
