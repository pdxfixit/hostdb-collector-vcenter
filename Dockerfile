FROM scratch

COPY hostdb-collector-vcenter /usr/bin/
COPY config.yaml /etc/hostdb-collector-vcenter/

ENV HOSTDB_COLLECTOR_VCENTER_DEBUG=false

ENTRYPOINT [ "/usr/bin/hostdb-collector-vcenter" ]
