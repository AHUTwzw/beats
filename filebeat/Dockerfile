FROM alpine:3.15

ENV PATH=/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/go/bin

WORKDIR /usr/share/filebeat

COPY ./filebeat.exe /usr/share/filebeat/filebeat

ENTRYPOINT ["./filebeat"]

CMD [ "-e" ]
