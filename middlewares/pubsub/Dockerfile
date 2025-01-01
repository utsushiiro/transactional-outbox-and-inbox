
# https://github.com/GoogleCloudPlatform/cloud-sdk-docker?tab=readme-ov-file#installing-a-docker-image
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:489.0.0-stable

RUN apt-get update
RUN apt-get install -y git python3-pip netcat
RUN git clone https://github.com/googleapis/python-pubsub.git

# https://github.com/googleapis/python-pubsub/tree/main/samples/snippets
WORKDIR /python-pubsub/samples/snippets
RUN pip3 install -r requirements.txt

COPY ./entrypoint.sh ./
EXPOSE 10002
ENTRYPOINT ["./entrypoint.sh"]
