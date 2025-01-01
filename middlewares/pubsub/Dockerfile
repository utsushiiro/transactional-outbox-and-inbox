# https://github.com/GoogleCloudPlatform/cloud-sdk-docker/tree/master?tab=readme-ov-file#docker-image-options
# https://cloud.google.com/sdk/docs/downloads-docker#components_installed_in_each_tag
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:489.0.0

RUN apt-get update

# https://github.com/docker/for-mac/issues/7255#issuecomment-2540082464
RUN apt-get install -y ca-certificates

# > The virtual environment was not created successfully because ensurepip is not
# > available.  On Debian/Ubuntu systems, you need to install the python3-venv
# > package using the following command.
# >
# >     apt install python3-venv
# https://packages.debian.org/bookworm/python3-venv
RUN apt-get install -y git python3-venv netcat-traditional

# https://github.com/googleapis/python-pubsub/tree/main/samples/snippets
RUN git clone https://github.com/googleapis/python-pubsub.git
WORKDIR /python-pubsub/samples/snippets
RUN python3 -m venv venv
RUN venv/bin/pip install -r requirements.txt

COPY ./entrypoint.sh ./
EXPOSE 10002
ENTRYPOINT ["./entrypoint.sh"]
