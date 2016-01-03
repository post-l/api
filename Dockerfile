FROM golang:1.5.1
MAINTAINER Ludovic Post <ludovic.post@epitech.eu>

EXPOSE 3000

COPY . src/github.com/post-l/api

RUN go get github.com/post-l/api
RUN go install github.com/post-l/api

CMD ["api"]
