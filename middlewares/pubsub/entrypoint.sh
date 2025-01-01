#!/bin/bash

gcloud beta emulators pubsub start \
  --project=$PUBSUB_PROJECT_ID \
  --host-port=${PUBSUB_EMULATOR_HOST}:${PUBSUB_EMULATOR_PORT} --quiet &

while ! nc -z localhost ${PUBSUB_EMULATOR_PORT}; do
  sleep 0.1
done

. venv/bin/activate

python3 publisher.py $PUBSUB_PROJECT_ID create $PUBSUB_TOPIC_ID
python3 subscriber.py $PUBSUB_PROJECT_ID create $PUBSUB_TOPIC_ID $PUBSUB_SUBSCRIPTION_ID

fg %1
