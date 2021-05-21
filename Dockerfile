FROM keppel.eu-de-1.cloud.sap/ccloud-dockerhub-mirror/library/python:3

LABEL source_repository="https://github.com/sapcc/concourse-awx-resource"

COPY assets/ /opt/resource/
RUN chmod +x /opt/resource/*

COPY requirements.txt /
RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt
