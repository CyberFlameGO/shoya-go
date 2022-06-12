FROM alpine:3.16

RUN addgroup -S shoya && adduser -S shoya -G shoya && chown -R shoya:shoya /home/shoya

USER shoya
WORKDIR /home/shoya
COPY --chown=shoya:shoya ./bin/files /home/shoya/files
RUN chmod +x /home/shoya/files

CMD ["/home/shoya/files"]