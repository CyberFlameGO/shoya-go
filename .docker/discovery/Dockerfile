FROM alpine:3.16

RUN addgroup -S shoya && adduser -S shoya -G shoya && chown -R shoya:shoya /home/shoya

COPY ./bin/discovery /home/shoya/discovery
RUN chmod +x /home/shoya/discovery

USER shoya
WORKDIR /home/shoya
CMD ["/home/shoya/discovery"]