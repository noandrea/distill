FROM alpine:3.7

LABEL author="no.andrea@gmail.com"

RUN ls
COPY ./dist /data
RUN mkdir -p /data/db

ENTRYPOINT [ "/data/ilij", "-c", "/data/ilij.conf.yaml" ]
CMD ["start"]
