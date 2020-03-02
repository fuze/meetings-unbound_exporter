p!/bin/bash -x

# https://github.com/florentchauveau/kamailio_exporter
# releases/download/v0.2.1/kamailio_exporter-0.2.1.linux-amd64.tar.gz


# cd ${WORKSPACE}
REPONAME=florentchauveau
PACKAGENAME=kamailio_exporter
# note: work-around as debs and fpm do not like underscores
# https://github.com/jordansissel/fpm/commit/808cadee4218ff28cce2f5625bb5c48a104df482
DEBPACKAGENAME=kamailio-exporter

VERSION=${VERSION}

# fpm extensions:
MAINTAINER=fuze-ops@fuze.com
VENDOR=florentchauveau/kamailio_exporter
URI=https://github.com/florentchauveau/kamailio_exporter
CATEGORY=monitoring
DESCRIPTION="Prometheus Exporter for Kamailio"


if [[ -z "${VERSION}" ]]; then
  echo $"Usage: $0 <VERSION> [ARCH]"
  exit 1
fi


ARCH=x86_64

case "${ARCH}" in
  i386)
    TGZ=${PACKAGENAME}-${VERSION}.linux-386.tgz
    ;;
  x86_64)
    TGZ=${PACKAGENAME}-${VERSION}.linux-amd64.tar.gz
    ;;
  *)
    echo $"Unknown architecture ${ARCH}" >&2
    exit 1
    ;;
esac

echo $"Creating ${PACKAGENAME} of type: ${ARCH} and version: ${VERSION}"

URL="https://github.com/${REPONAME}/${PACKAGENAME}/releases/download/v${VERSION}/${TGZ}"

echo $"DEBUG: The package is: ${TGZ}"
echo $"DEBUG: The url is: ${URL}"

# fetching release
curl --fail -k -L -o $TGZ $URL || {
    echo $"URL or version not found!" >&2
    exit 1
}

# clear target foler
rm -rf startup/*
rm -rf target/*

# create target structure
mkdir -p startup
mkdir -p target/usr/local/bin

# unzip
tar -zxf ${TGZ} -C target/usr/local/bin/ --strip-components 1

# SystemD Config
tee -a startup/kamailio_exporter.service <<END
[Unit]
Description=kamailio_exporter for Prometheus metrics
Wants=network-online.target
After=network-online.target
[Service]
User=root
ExecStart=/usr/local/bin/kamailio_exporter -m tm.stats,sl.stats,core.shmmem,core.uptime,dispatcher.list
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
[Install]
WantedBy=default.target
END


echo $PWD

echo "removing ${TGZ} .."
rm ${TGZ}

# create deb
fpm -s dir -t deb -f --verbose \
  -C target \
  --category $CATEGORY --url $URI --vendor $VENDOR \
  -n ${DEBPACKAGENAME} \
  -v ${VERSION} \
  -a ${ARCH} \
  -m $MAINTAINER \
  -p target \
  --iteration ${BUILD_NUMBER} \
  --description "Prometheus Exporter for Kamailio" \
  --deb-no-default-config-files \
  --deb-systemd startup/kamailio_exporter.service

if [[ $? -gt 0 ]]; then
   echo "fpm found a problem"
   rm -rf target/usr
   exit 255
fi

rm -rf target/usr

#if [ -f ./target/${DEBPACKAGENAME}_${VERSION}$_amd64.deb ]; then
#    exit 0
#fi

echo "some debugging for s3 upload"

# INTG: Upload: 
deb-s3 upload --bucket fuze-apt-bionic-rc --access-key-id=$AWS_ACCESS_KEY_ID --secret-access-key=$AWS_SECRET_ACCESS_KEY \
    --codename bionic --arch amd64 --preserve-versions \
    ./target/${DEBPACKAGENAME}_${VERSION}-${BUILD_NUMBER}_amd64.deb
# Verify: 
deb-s3 verify --bucket fuze-apt-bionic-rc --access-key-id=$AWS_ACCESS_KEY_ID --secret-access-key=$AWS_SECRET_ACCESS_KEY
# master/main: upload
deb-s3 upload --bucket fuze-apt-bionic-master --access-key-id=$AWS_ACCESS_KEY_ID --secret-access-key=$AWS_SECRET_ACCESS_KEY \
    --codename bionic --arch amd64 --preserve-versions \
    ./target/${DEBPACKAGENAME}_${VERSION}-${BUILD_NUMBER}_amd64.deb
# INTG: Verify
deb-s3 verify --bucket fuze-apt-bionic-master --access-key-id=$AWS_ACCESS_KEY_ID --secret-access-key=$AWS_SECRET_ACCESS_KEY


if ${PROD}; then
# deb-s3 upload --bucket fuze-apt-trusty-prod --access-key-id=$AWS_ACCESS_KEY_ID_PROD --secret-access-key=$AWS_SECRET_ACCESS_KEY_PROD \
#     --codename trusty --arch amd64 --preserve-versions .target/${DEBPACKAGENAME}_${VERSION}_amd64.deb
# deb-s3 verify --bucket fuze-apt-trusty-prod --access-key-id=$AWS_ACCESS_KEY_ID_PROD --secret-access-key=$AWS_SECRET_ACCESS_KEY_PROD
  deb-s3 upload --bucket fuze-apt-bionic-prod --access-key-id=$AWS_ACCESS_KEY_ID_PROD --secret-access-key=$AWS_SECRET_ACCESS_KEY_PROD \
    --codename bionic --arch amd64 --preserve-versions \
    ./target/${DEBPACKAGENAME}_${VERSION}-${BUILD_NUMBER}_amd64.deb
  deb-s3 verify --bucket fuze-apt-bionic-prod --access-key-id=$AWS_ACCESS_KEY_ID_PROD --secret-access-key=$AWS_SECRET_ACCESS_KEY_PROD
fi
