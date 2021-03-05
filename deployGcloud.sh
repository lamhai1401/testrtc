#!/bin/bash
sudo docker stop $(sudo docker ps -a -q)
sudo docker rm $(sudo docker ps -a -q)
sudo docker build -t docker-test-rtc:test .

# push to gcloud images
sudo docker tag docker-test-rtc:test  gcr.io/livestreaming-241004/docker-test-rtc:test 
gcloud auth print-access-token | sudo docker login -u oauth2accesstoken --password-stdin https://gcr.io/livestreaming-241004
sudo docker push gcr.io/livestreaming-241004/docker-test-rtc:test 

# run docker locally
# sudo docker run -it docker-test-rtc:test 