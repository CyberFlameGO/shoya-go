FROM alpine:3.16

RUN apk add libc6-compat && addgroup -S shoya && adduser -S shoya -G shoya && chown -R shoya:shoya /home/shoya

USER shoya
WORKDIR /home/shoya
COPY --chown=shoya:shoya ./bin/shoya /home/shoya/shoya
RUN chmod +x /home/shoya/shoya

CMD ["/home/shoya/shoya", "files", "serve"]