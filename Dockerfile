FROM amd64/golang:1.16.0-alpine3.13
RUN apk --no-cache add bash git mercurial subversion openssh-client ca-certificates

# RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
# ENV GOPATH /go
# ENV PATH /go/bin:$PATH

# setup go env
# ENV GO111MODULE=on
# ENV GIT_TERMINAL_PROMPT=1
# ENV GONOPROXY=github.com/beowulflab/*
# ENV GOPRIVATE=github.com/beowulflab/*
# ENV GOROOT /usr/lib/go
# ENV GOPATH /go
# ENV PATH /go/bin:$PATH
# RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin


ENV WSS_URL="wss://signal-conference-staging.quickom.com"
ENV USERNAME="hai"

# RUN git clone https://github.com/lamhai1401/testrtc.git
# COPY ./ /${GOPATH}/src/testrtc
# WORKDIR /${GOPATH}/src/testrtc
# WORKDIR /go/src/testrtc
# RUN go mod download
# RUN go clean
ADD ./ /${GOPATH}/src/testrtc
WORKDIR /${GOPATH}/src/testrtc
RUN ls
# RUN go build -o bot-core
RUN chmod +x ./testrtc
ADD entrypoint.sh /
ENTRYPOINT ["sh" ,"/entrypoint.sh"]