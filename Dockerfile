FROM keppel.eu-de-1.cloud.sap/ccloud-dockerhub-mirror/library/python:3

LABEL source_repository="https://github.com/sapcc/concourse-awx-resource"

COPY SAPNetCA_G2.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates

ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

COPY assets/ /opt/resource/
RUN chmod +x /opt/resource/*

COPY requirements.txt /
RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt
