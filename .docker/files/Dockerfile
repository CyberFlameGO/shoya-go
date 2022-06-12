FROM alpine:3.16

RUN addgroup -S shoya && adduser -S shoya -G shoya && chown -R shoya:shoya /home/shoya

COPY ./bin/files /home/shoya/files
RUN chmod +x /home/shoya/files

USER shoya
WORKDIR /home/shoya
CMD ["/home/shoya/files"]