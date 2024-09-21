FROM golang:1.23

RUN mkdir /jwt-auth
WORKDIR /jwt-auth

COPY . .
RUN chmod a+x docker/*.sh