FROM alpine:3.16

RUN addgroup -S shoya && adduser -S shoya -G shoya && chown -R shoya:shoya /home/shoya

COPY ./bin/api /home/shoya/api
RUN chmod +x /home/shoya/api

USER shoya
WORKDIR /home/shoya
CMD ["/home/shoya/api"]