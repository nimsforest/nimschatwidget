FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY nimschatwidget /usr/local/bin/nimschatwidget
ENTRYPOINT ["nimschatwidget"]
CMD ["serve"]
