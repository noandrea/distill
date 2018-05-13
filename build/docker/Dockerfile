FROM alpine:3.7

LABEL author="no.andrea@gmail.com"

RUN ls
COPY ./dist /data
RUN mkdir -p /data/db

ENTRYPOINT [ "/data/distill" ]
CMD [ "start", "-c", "/data/configs/settings.docker.yaml"]
