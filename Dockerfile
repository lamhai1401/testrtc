FROM ubuntu

RUN apt-get update && apt dist-upgrade -y
RUN apt-get install software-properties-common -y
RUN apt-get install apt-transport-https ca-certificates curl software-properties-common -y
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable"
RUN apt-get update
RUN apt-cache policy docker-ce

ENV WSS_URL="wss://signal-conference-staging.quickom.com"
ENV USERNAME="hai"

RUN mkdir testrtc
WORKDIR /testrtc
ADD ./classroom-core /testrtc
RUN chmod +x ./classroom-core

ADD entrypoint.sh /
ENTRYPOINT ["sh" ,"/entrypoint.sh"]