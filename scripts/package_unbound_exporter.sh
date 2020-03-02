p!/bin/bash -x

# https://github.com/florentchauveau/kamailio_exporter
# releases/download/v0.2.1/kamailio_exporter-0.2.1.linux-amd64.tar.gz


# cd ${WORKSPACE}
REPONAME=fuze
PACKAGENAME=unbound_exporter
DEBPACKAGENAME=unbound-exporter
VERSION=${VERSION}

# fpm extensions:
MAINTAINER=fuze-ops@fuze.com
VENDOR=florentchauveau/kamailio_exporter
URI=https://github.com/florentchauveau/kamailio_exporter
CATEGORY=monitoring
DESCRIPTION="Prometheus Exporter for Kamailio"


ARCH=x86_64

echo $"Creating ${PACKAGENAME} of type: ${ARCH} and version: ${VERSION}"

URL="https://github.com/${REPONAME}/${PACKAGENAME}/releases/download/v${VERSION}/${TGZ}"


# create target structure
mkdir -p startup
mkdir -p target/usr/local/bin

mv bin/


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
