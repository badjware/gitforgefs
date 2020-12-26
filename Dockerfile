FROM alpine

COPY ./bin/gitlabfs /usr/bin/gitlabfs

ENTRYPOINT ["gitlabfs"]

